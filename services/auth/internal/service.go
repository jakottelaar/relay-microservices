package internal

import (
	"context"
	"encoding/json"
	"net/netip"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jakottelaar/relay-microservices/services/auth/config"
	"github.com/jakottelaar/relay-microservices/services/auth/internal/queries"
	"github.com/jakottelaar/relay-microservices/shared/errors"
	"github.com/jakottelaar/relay-microservices/shared/events"
	"github.com/jakottelaar/relay-microservices/shared/sonyflake"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

type AuthService interface {
    SignUp(ctx context.Context, req SignUpRequest, metadata SessionMetadata) (*AuthResponse, error)
    SignIn(ctx context.Context, req SignInRequest, metadata SessionMetadata) (*AuthResponse, error)
    RefreshToken(ctx context.Context, refreshToken string, metadata SessionMetadata) (*AuthResponse, error)
    SignOut(ctx context.Context, refreshToken string) error
    RevokeAllSessions(ctx context.Context, userID int64) error
    RevokeSessionById(ctx context.Context, sessionID int64) error
    GetSessionByID(ctx context.Context, sessionID int64) (*SessionResponse, error)
}

type authService struct {
    repo       *AuthRepository
    jwtManager JWTManager
    config     *config.Config
    nc         *nats.Conn
    log *zap.Logger
}

func NewAuthService(repo *AuthRepository, jwtManager JWTManager, config *config.Config, nc *nats.Conn, log *zap.Logger) *authService {
    return &authService{
        repo:       repo,
        jwtManager: jwtManager,
        config:     config,
        nc:         nc,
        log:        log,
    }
}

func (s *authService) SignUp(ctx context.Context, req SignUpRequest, metadata SessionMetadata) (*AuthResponse, error) {
    _, err := s.repo.Queries.GetAccountByEmail(ctx, req.Email)
    if err == nil {
        s.log.Warn("Sign-up attempt with already registered email",
            zap.String("email", req.Email),
        )
        return nil, errors.NewDuplicateError("Email already registered")
    }

    hashedPassword, err := argon2id.CreateHash(req.Password, argon2id.DefaultParams)
    if err != nil {
        s.log.Error("Failed to hash password",
            zap.Error(err),
        )
        return nil, errors.NewInternalServerError("failed to create user")
    }

    accountId, err := sonyflake.GenerateSonyFlakeID()
    if err != nil {
        return nil, errors.NewInternalServerError("failed to generate account ID")
    }

    createdAccount, err := s.repo.Queries.CreateAccount(ctx, queries.CreateAccountParams{
        ID:           int64(accountId),
        Email:        req.Email,
        PasswordHash: hashedPassword,
    })
    if err != nil {
        s.log.Error("Database error creating account",
            zap.Error(err),
            zap.String("email", req.Email),
        )
        return nil, errors.NewInternalServerError("failed to create user: " + err.Error())
    }

    authResp, err := s.createSessionAndTokens(ctx, &Account{
        ID:        createdAccount.ID,
        Email:     createdAccount.Email,
        CreatedAt: createdAccount.CreatedAt.Time,
    }, metadata)
    if err != nil {
        return nil, err
    }

    event := events.UserAccountCreatedEvent{
        UserID:   createdAccount.ID,
        Username: req.Username,
        CreatedAt: createdAccount.CreatedAt.Time,
    }

    eventData , err := json.Marshal(event)
    if err != nil {
        s.log.Error("Failed to marshal user account created event",
            zap.Error(err),
            zap.Int64("user_id", createdAccount.ID),
        )
    } else {
        err := s.nc.Publish(events.SubjectUserAccountCreated, eventData)
        if err != nil {
            s.log.Error("Failed to publish user account created event",
                zap.Error(err),
                zap.Int64("user_id", createdAccount.ID),
            )
        }
    }

    return authResp, nil
}

func (s *authService) SignIn(ctx context.Context, req SignInRequest, metadata SessionMetadata) (*AuthResponse, error) {
    account, err := s.repo.Queries.GetAccountByEmail(ctx, req.Email)
    if err != nil {
        return nil, errors.NewUnauthorizedError("invalid email or password")
    }

    match, err := argon2id.ComparePasswordAndHash(req.Password, account.PasswordHash)
    if err != nil {
        return nil, errors.NewInternalServerError("failed to verify password")
    }
    if !match {
        return nil, errors.NewUnauthorizedError("invalid email or password")
    }

    authResp, err := s.createSessionAndTokens(ctx, &Account{
        ID:        account.ID,
        Email:     account.Email,
        CreatedAt: account.CreatedAt.Time,
    }, metadata)
    if err != nil {
        return nil, err
    }

    return authResp, nil
}

func (s *authService) RefreshToken(ctx context.Context, refreshToken string, metadata SessionMetadata) (*AuthResponse, error) {
    tokenHash := HashRefreshToken(refreshToken)

    session, err := s.repo.Queries.GetSessionByTokenHash(ctx, tokenHash)
    if err != nil {
        if err == pgx.ErrNoRows {
            return nil, errors.NewUnauthorizedError("invalid or expired refresh token")
        }
        return nil, errors.NewInternalServerError("failed to fetch session")
    }

    // Check if session was revoked (reuse detection)
    if session.RevokedAt.Valid {
        // Revoke all user sessions as security measure
        _ = s.repo.Queries.RevokeAllUserSessions(ctx, session.UserAccountID)
        return nil, errors.NewUnauthorizedError("refresh token reused - all sessions revoked for security")
    }

    if session.ExpiresAt.Time.Before(time.Now()) {
        return nil, errors.NewUnauthorizedError("refresh token expired")
    }

    // Revoke old session (refresh token rotation)
    if err := s.repo.Queries.RevokeSession(ctx, session.ID); err != nil {
        return nil, errors.NewInternalServerError("failed to revoke old session")
    }

    account, err := s.repo.Queries.GetAccountByID(ctx, session.UserAccountID)
    if err != nil {
        return nil, errors.NewInternalServerError("failed to fetch account")
    }

    authResp, err := s.createSessionAndTokens(ctx, &Account{
        ID:        account.ID,
        Email:     account.Email,
        CreatedAt: account.CreatedAt.Time,
    }, metadata)
    if err != nil {
        return nil, err
    }

    return authResp, nil
}

func (s *authService) SignOut(ctx context.Context, refreshToken string) error {
    tokenHash := HashRefreshToken(refreshToken)

    session, err := s.repo.Queries.GetSessionByTokenHash(ctx, tokenHash)
    if err != nil {
        if err == pgx.ErrNoRows {
            // Already logged out or invalid token - not an error
            return nil
        }
        return errors.NewInternalServerError("failed to fetch session")
    }

    if err := s.repo.Queries.RevokeSession(ctx, session.ID); err != nil {
        return errors.NewInternalServerError("failed to revoke session")
    }

    return nil
}

func (s *authService) RevokeAllSessions(ctx context.Context, userID int64) error {
    if err := s.repo.Queries.RevokeAllUserSessions(ctx, userID); err != nil {
        return errors.NewInternalServerError("failed to revoke all sessions")
    }
    return nil
}

func (s *authService) RevokeSessionById(ctx context.Context, sessionID int64) error {
    if err := s.repo.Queries.RevokeSession(ctx, sessionID); err != nil {
        if err == pgx.ErrNoRows {
            return errors.NewNotFoundError("session not found")
        }
        return errors.NewInternalServerError("failed to revoke session")
    }

    return nil
}

func (s *authService) GetSessionByID(ctx context.Context, sessionID int64) (*SessionResponse, error) {
    session, err := s.repo.Queries.GetSessionByID(ctx, sessionID)
    if err != nil {
        if err == pgx.ErrNoRows {
            return nil, errors.NewNotFoundError("session not found")
        }
        return nil, errors.NewInternalServerError("failed to fetch session")
    }

    var approxLastTimeUsed time.Time
    if session.LastUsedAt.Valid {
        approxLastTimeUsed = session.LastUsedAt.Time
    } else {
        approxLastTimeUsed = session.CreatedAt.Time
    }

    response := &SessionResponse{
        SessionID:          session.ID,
        ApproxLastTimeUsed: approxLastTimeUsed.Format(time.RFC3339),
    }
    if session.UserAgent.Valid {
        response.Client.UserAgent = session.UserAgent.String
    }
    
    return response, nil
}

func (s *authService) createSessionAndTokens(
    ctx context.Context,
    account *Account,
    metadata SessionMetadata,
) (*AuthResponse, error) {
    // Check active session count
    count, err := s.repo.Queries.CountActiveSessions(ctx, account.ID)
    if err != nil {
        return nil, errors.NewInternalServerError("failed to count sessions")
    }

    // If max sessions reached, revoke oldest
    if count >= int64(s.config.MaxSessionsPerUser) {
        if err := s.repo.Queries.RevokeOldestSession(ctx, account.ID); err != nil {
            return nil, errors.NewInternalServerError("failed to revoke oldest session")
        }
    }

    // Generate refresh token
    refreshToken, err := GenerateRefreshToken()
    if err != nil {
        return nil, errors.NewInternalServerError("failed to generate refresh token")
    }
    refreshTokenHash := HashRefreshToken(refreshToken)

    // Generate session ID
    sessionID, err := sonyflake.GenerateSonyFlakeID()
    if err != nil {
        return nil, errors.NewInternalServerError("failed to generate session ID")
    }

    // Prepare session data
    expiresAt := pgtype.Timestamptz{}
    expiresAt.Scan(time.Now().Add(s.config.RefreshTokenExpiry))

    userAgent := pgtype.Text{}
    if metadata.UserAgent != "" {
        userAgent = pgtype.Text{String: metadata.UserAgent, Valid: true}
    }

    var ipAddr *netip.Addr
    if metadata.IPAddress != "" {
        if addr, err := netip.ParseAddr(metadata.IPAddress); err == nil {
            ipAddr = &addr
        }
    }

    _, err = s.repo.Queries.CreateSession(ctx, queries.CreateSessionParams{
        ID:               int64(sessionID),
        UserAccountID:    account.ID,
        RefreshTokenHash: refreshTokenHash,
        UserAgent:        userAgent,
        IpAddress:        ipAddr,
        ExpiresAt:        expiresAt,
    })
    if err != nil {
        return nil, errors.NewInternalServerError("failed to create session")
    }

    accessToken, err := s.jwtManager.GenerateToken(account.ID, int64(sessionID))
    if err != nil {
        return nil, errors.NewInternalServerError("failed to generate access token")
    }

    return &AuthResponse{
        AccessToken:  accessToken,
        RefreshToken: refreshToken,
        ExpiresIn:    int(s.config.AccessTokenExpiry.Seconds()),
        Account:      account,
    }, nil
}