package internal

import (
	"context"

	"github.com/jakottelaar/relay-microservices/services/users/internal/queries"
	"go.uber.org/zap"
)

type UserService interface {
	CreateUser(ctx context.Context, req *CreateUserRequest) error
}

type userService struct {
	repo *UserRepository
	log *zap.Logger
}

func NewUserService(repo *UserRepository, log *zap.Logger) *userService {
	return &userService{
		repo: repo,
		log: log,
	}
}

func (s *userService) CreateUser(ctx context.Context, req *CreateUserRequest) error {
	s.log.Info(
		"Creating user profile",
		zap.Int64("user_id", req.UserID),
		zap.String("username", req.Username),
	)

	_, err := s.repo.CreateUser(ctx, queries.CreateUserParams{
		ID: req.UserID,
		Username: req.Username,
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