package errors

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

func IsUniqueViolation(err error) bool {
    if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
        return pgErr.Code == "23505"
    }
    return false
}