package internal

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jakottelaar/relay-microservices/services/users/internal/queries"
)

type UserRepository struct {
	*queries.Queries
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{
		Queries: queries.New(pool),
		pool:    pool,
	}
}

func (r *UserRepository) WithTx(ctx context.Context, fn func(*queries.Queries) error) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	q := queries.New(tx)
	if err := fn(q); err != nil {
		return err
	}

	return tx.Commit(ctx)
}