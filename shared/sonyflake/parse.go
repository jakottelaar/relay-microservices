package sonyflake

import (
	"strconv"

	"github.com/jakottelaar/relay-microservices/shared/errors"
)

func ParseID(id string) (int64, error) {
	if id == "" {
		return 0, errors.NewBadRequestError("id is required")
	}

	parsed, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return 0, errors.NewBadRequestError("invalid id format")
	}

	if parsed <= 0 {
		return 0, errors.NewBadRequestError("id must be a positive number")
	}

	return parsed, nil
}