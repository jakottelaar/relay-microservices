package tests

import (
	"context"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jakottelaar/relay-microservices/services/auth/internal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/jackc/pgx/v5/pgxpool"
)

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
	service := internal.NewAuthService(repo)

	ctx := context.Background()

	t.Run("Successful sign up", func(t *testing.T) {
		req := internal.SignUpRequest{
			Email: "test@mail.com",
			Password: "Password1234!",
		}

		account, err := service.SignUp(ctx, req)
		require.NoError(t, err)
		assert.NotEmpty(t, account.ID)
		assert.NotZero(t, account.ID)
		assert.Equal(t, req.Email, account.Email)
	})

	t.Run("duplicate email fails", func(t *testing.T) {
		req := internal.SignUpRequest{
			Email:    "duplicate@example.com",
			Password: "password123",
		}

		_, err := service.SignUp(ctx, req)
		require.NoError(t, err)

		_, err = service.SignUp(ctx, req)
		require.Error(t, err)
		assert.Equal(t, "Email already registered", err.Error())
	})

	t.Run("password is hashed", func(t *testing.T) {
		req := internal.SignUpRequest{
			Email:    "hash@example.com",
			Password: "mypassword",
		}

		account, err := service.SignUp(ctx, req)
		require.NoError(t, err)

		// Verify password is not stored in plain text
		dbAccount, err := repo.GetAccountByID(ctx, account.ID)
		require.NoError(t, err)
		assert.NotEqual(t, req.Password, dbAccount.PasswordHash)
		assert.NotEmpty(t, dbAccount.PasswordHash)
	})

}