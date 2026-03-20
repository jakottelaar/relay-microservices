package internal

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jakottelaar/relay-microservices/services/users/config"
	"github.com/jakottelaar/relay-microservices/services/users/internal/queries"
	"github.com/jakottelaar/relay-microservices/shared/errors"
	"github.com/minio/minio-go/v7"
	"go.uber.org/zap"
)

type UserService interface {
    CreateUser(ctx context.Context, req *CreateUserRequest) error
    GetUserProfile(ctx context.Context, userID int64) (*ProfileResponse, error)
}

type userService struct {
    repo    *UserRepository
    log     *zap.Logger
    storage *minio.Client
    cfg     *config.Config
}

func NewUserService(repo *UserRepository, log *zap.Logger, storage *minio.Client, cfg *config.Config) *userService {
    return &userService{
        repo:    repo,
        log:     log,
        storage: storage,
        cfg:     cfg,
    }
}

func (s *userService) CreateUser(ctx context.Context, req *CreateUserRequest) error {
    s.log.Info("Creating user profile",
        zap.Int64("user_id", req.UserID),
        zap.String("username", req.Username),
    )

    _, err := s.repo.CreateUser(ctx, queries.CreateUserParams{
        ID:       req.UserID,
        Username: req.Username,
        Avatar: pgtype.Text{String: s.cfg.Storage.DefaultAvatarURL, Valid: true},
    })
    if err != nil {
        s.log.Error("Failed to create user profile",
            zap.Error(err),
            zap.Int64("user_id", req.UserID),
            zap.String("username", req.Username),
        )
        return err
    }

    s.log.Info("User profile created successfully",
        zap.Int64("user_id", req.UserID),
        zap.String("username", req.Username),
    )
    return nil
}

func (s *userService) GetUserProfile(ctx context.Context, userID int64) (*ProfileResponse, error) {
    user, err := s.repo.GetUserByID(ctx, userID)
    if err != nil {
        if err == pgx.ErrNoRows {
            s.log.Warn("User profile not found",
                zap.Int64("user_id", userID),
            )
            return nil, errors.NewNotFoundError("user not found")
        }

        s.log.Error("Failed to get user profile",
            zap.Error(err),
            zap.Int64("user_id", userID),
        )
        return nil, err
    }

    userResp := &ProfileResponse{
        ID:         user.ID,
        Username:   user.Username,
        AvatarURL:  user.Avatar.String,
        Bio:        user.Bio.String,
        CreatedAt:  user.CreatedAt.Time,
    }

    s.log.Info("User profile retrieved successfully",
        zap.Int64("user_id", userID),
    )
    return userResp, nil
}