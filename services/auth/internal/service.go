package internal

import (
	"context"
	"net/netip"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jakottelaar/relay-microservices/services/auth/config"
	"github.com/jakottelaar/relay-microservices/services/auth/internal/queries"
)

type AuthService interface {
	SignUp(ctx context.Context, req SignUpRequest, metadata SessionMetadata) (*AuthResponse, error)
}

type authService struct {
	repo *AuthRepository
	jwtManager *JWTManager
	config     *config.Config
}

func NewAuthService(repo *AuthRepository, jwtManager *JWTManager, config *config.Config) *authService {
	return &authService{
		repo: repo,
		jwtManager: jwtManager,
		config: config,
	}
}

func (s *authService) SignUp(ctx context.Context, req SignUpRequest, metadata SessionMetadata) (*AuthResponse, error) {
	_, err := s.repo.Queries.GetAccountByEmail(ctx, req.Email)
	if err == nil {
		return nil, NewDuplicateError("Email already registered")
	}

	hashedPassword, err := argon2id.CreateHash(req.Password, argon2id.DefaultParams)
	if err != nil {
		return nil, NewInternalServerError("failed to create user")
	}

	accountId, err := sf.NextID()
	if err != nil {
		return nil, NewInternalServerError("failed to generate account ID")
	}

	createdAccount, err := s.repo.Queries.CreateAccount(ctx, queries.CreateAccountParams{
		ID:           int64(accountId),
		Email:        req.Email,
		PasswordHash: hashedPassword,
	})
	if err != nil {
		return nil, NewInternalServerError("failed to create user: " + err.Error())
	}

	authResp, err := s.createAuthResponse(ctx, &Account{
		ID:        createdAccount.ID,
		Email:     createdAccount.Email,
		CreatedAt: createdAccount.CreatedAt.Time,
	}, metadata)
	if err != nil {
		return nil, err
	}

	return authResp, nil
}

func (s *authService) ValidateToken(ctx context.Context, token string) (*Claims, error) {
	return s.jwtManager.ValidateToken(token)
}

func (s *authService) createAuthResponse(ctx context.Context, account *Account, metadata SessionMetadata) (*AuthResponse, error) {
	count, err := s.repo.Queries.CountActiveSessions(ctx, account.ID)
	if err != nil {
		return nil, NewInternalServerError("failed to count sessions")
	}
	if count >= int64(s.config.MaxSessionsPerUser) {
		if err := s.repo.Queries.RevokeAllUserSessions(ctx, account.ID); err != nil {
			return nil, NewInternalServerError("failed to revoke old sessions")
		}
	}


	accessToken, err := s.jwtManager.GenerateToken(account.ID)
	if err != nil {
		return nil, NewInternalServerError("failed to generate access token")
	}

	refreshToken, err := GenerateRefreshToken()
	if err != nil {
		return nil, NewInternalServerError("failed to generate refresh token")
	}

	refreshTokenHash := HashRefreshToken(refreshToken)

	sessionID, err := sf.NextID()
	if err != nil {
		return nil, NewInternalServerError("failed to generate session ID")
	}

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

	expiresAt := pgtype.Timestamptz{}
	expiresAt.Scan(time.Now().Add(s.config.RefreshTokenExpiry))

	_, err = s.repo.Queries.CreateSession(ctx, queries.CreateSessionParams{
		ID:               int64(sessionID),
		UserAccountID:    account.ID,
		RefreshTokenHash: refreshTokenHash,
		UserAgent:        userAgent,
		IpAddress:        ipAddr,
		ExpiresAt:        expiresAt,
	})
	if err != nil {
		return nil, NewInternalServerError("failed to create session: " + err.Error())
	}

	return &AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int(s.config.AccessTokenExpiry.Seconds()),
		Account:      account,
	}, nil
}