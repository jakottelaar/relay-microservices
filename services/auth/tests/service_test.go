package tests

import (
	"context"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jakottelaar/relay-microservices/services/auth/config"
	"github.com/jakottelaar/relay-microservices/services/auth/internal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/jackc/pgx/v5/pgxpool"
)

type MockJWTManager struct {
    mock.Mock
}

func (m *MockJWTManager) GenerateToken(userID int64) (string, error) {
    args := m.Called(userID)
    return args.String(0), args.Error(1)
}

func (m *MockJWTManager) ValidateToken(tokenString string) (*internal.Claims, error) {
    args := m.Called(tokenString)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*internal.Claims), args.Error(1)
}

func setUpTestDb(t *testing.T) (*pgxpool.Pool, func()) {
	ctx := context.Background()

	pgContainer, err := postgres.Run(ctx, 
		"postgres:18-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second)),
	)
	require.NoError(t, err)

	err = internal.InitSnowflake()
	require.NoError(t, err)

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	pool, err := pgxpool.New(ctx, connStr)
	require.NoError(t, err)

	runMigrations(t, connStr)

	return pool, func() {
		pool.Close()
		pgContainer.Terminate(ctx)
	}
}

func runMigrations(t *testing.T, connStr string) {

	m, err := migrate.New(
		"file://../migrations",
		connStr,
	)
	require.NoError(t, err)

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		require.NoError(t, err)
	}

}

func TestSignUp(t *testing.T) {
	pool, cleanup := setUpTestDb(t)
	defer cleanup()

	repo := internal.NewAuthRepository(pool)
	jwtManager := new(MockJWTManager)
	cfg := &config.Config{}
	service := internal.NewAuthService(repo, jwtManager, cfg)

	ctx := context.Background()

	t.Run("Successful sign up", func(t *testing.T) {
		jwtManager.On("GenerateToken", mock.AnythingOfType("int64")).Return("mock-jwt-token", nil)
		
		req := internal.SignUpRequest{
			Email: "test@mail.com",
			Password: "Password1234!",
		}

		resp, err := service.SignUp(ctx, req, internal.SessionMetadata{
			UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
			IPAddress:        "192.168.1.100",
		})
		require.NoError(t, err)
		assert.NotEmpty(t, resp.Account.ID)
		assert.NotZero(t, resp.Account.ID)
		assert.Equal(t, req.Email, resp.Account.Email)
	})

	t.Run("duplicate email fails", func(t *testing.T) {
		jwtManager.On("GenerateToken", mock.AnythingOfType("int64")).Return("mock-jwt-token", nil)
		
		req := internal.SignUpRequest{
			Email:    "duplicate@example.com",
			Password: "password123",
		}

		_, err := service.SignUp(ctx, req, internal.SessionMetadata{
			UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
			IPAddress:        "192.168.1.100",
		})
		require.NoError(t, err)

		_, err = service.SignUp(ctx, req, internal.SessionMetadata{
			UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
			IPAddress:        "192.168.1.100",
		})
		require.Error(t, err)
		assert.Equal(t, "Email already registered", err.Error())
	})

	t.Run("password is hashed", func(t *testing.T) {
		jwtManager.On("GenerateToken", mock.AnythingOfType("int64")).Return("mock-jwt-token", nil)
		
		req := internal.SignUpRequest{
			Email:    "hash@example.com",
			Password: "mypassword",
		}

		resp, err := service.SignUp(ctx, req, internal.SessionMetadata{
			UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
			IPAddress:        "192.168.1.100",
		})
		require.NoError(t, err)

		// Verify password is not stored in plain text
		dbAccount, err := repo.GetAccountByID(ctx, resp.Account.ID)
		require.NoError(t, err)
		assert.NotEqual(t, req.Password, dbAccount.PasswordHash)
		assert.NotEmpty(t, dbAccount.PasswordHash)
	})

}

func TestSignIn(t *testing.T) {
	pool, cleanup := setUpTestDb(t)
	defer cleanup()

	repo := internal.NewAuthRepository(pool)
	jwtManager := new(MockJWTManager)
	cfg := &config.Config{}
	service := internal.NewAuthService(repo, jwtManager, cfg)
	ctx := context.Background()

	// Common test data
	email := "signin@mail.com"
	password := "Password1234!"
	
	jwtManager.On("GenerateToken", mock.AnythingOfType("int64")).Return("mock-jwt-token", nil)
	
	sessionMeta := internal.SessionMetadata{
		UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
		IPAddress: "192.168.1.100",
	}

	// Setup: Create test user once
	_, err := service.SignUp(ctx, internal.SignUpRequest{
		Email:    email,
		Password: password,
	}, sessionMeta)
	require.NoError(t, err)

	t.Run("Successful sign in", func(t *testing.T) {
		resp, err := service.SignIn(ctx, internal.SignInRequest{
			Email:    email,
			Password: password,
		}, sessionMeta)

		require.NoError(t, err)
		assert.Equal(t, "mock-jwt-token", resp.AccessToken)
		assert.Equal(t, email, resp.Account.Email)
	})

	t.Run("invalid password fails", func(t *testing.T) {
		_, err := service.SignIn(ctx, internal.SignInRequest{
			Email:    email,
			Password: "WrongPassword!",
		}, sessionMeta)

		require.Error(t, err)
		assert.Equal(t, "invalid email or password", err.Error())
	})

	t.Run("non-existent email fails", func(t *testing.T) {
		_, err := service.SignIn(ctx, internal.SignInRequest{
			Email:    "non-existent@mail.com",
			Password: "SomePassword1234!",
		}, sessionMeta)
		require.Error(t, err)
		assert.Equal(t, "invalid email or password", err.Error())
	})
}
