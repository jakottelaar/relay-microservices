package internal

import (
	"context"

	"github.com/alexedwards/argon2id"
	"github.com/jakottelaar/relay-microservices/services/auth/internal/queries"
)

type AuthService interface {
	SignUp(ctx context.Context, req SignUpRequest) (*Account, error)
}

type authService struct {
	repo *AuthRepository
}

func NewAuthService(repo *AuthRepository) *authService {
	return &authService{
		repo: repo,
	}
}

func (s *authService) SignUp(ctx context.Context, req SignUpRequest) (*Account, error) {
	_, err := s.repo.Queries.GetAccountByEmail(ctx, req.Email)
	if err == nil {
		return nil, NewDuplicateError("Email already registered")
	}

	hashedPassword, err := argon2id.CreateHash(req.Password, argon2id.DefaultParams)
	if err != nil {
		return nil, NewInternalServerError("failed to create user")
	}

	req.Password = hashedPassword

	accountId, err := sf.NextID()
	if err != nil {
		return nil, NewInternalServerError("failed to generate account ID")
	}

	createdAccount, err := s.repo.Queries.CreateAccount(ctx, queries.CreateAccountParams{
		ID:           int64(accountId),
		Email:        req.Email,
		PasswordHash: req.Password,
	})
	if err != nil {
		return nil, NewInternalServerError("failed to create user: " + err.Error())
	}

	account := &Account{
		ID:           createdAccount.ID,
		Email:        createdAccount.Email,
		CreatedAt:    createdAccount.CreatedAt.Time,
	}

	return account, nil
}