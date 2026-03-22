package internal

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jakottelaar/relay-microservices/services/guilds/internal/queries"
)

type GuildRepository struct {
	*queries.Queries
	pool *pgxpool.Pool
}

func NewGuildRepository(pool *pgxpool.Pool) *GuildRepository {
	return &GuildRepository{
		Queries: queries.New(pool),
		pool:    pool,
	}
}

func (r *GuildRepository) WithTx(ctx context.Context, fn func(*queries.Queries) error) error {
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