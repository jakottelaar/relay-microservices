package internal

import (
	"context"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jakottelaar/relay-microservices/services/guilds/internal/queries"
	"github.com/jakottelaar/relay-microservices/shared/errors"
	"github.com/jakottelaar/relay-microservices/shared/sonyflake"
	"go.uber.org/zap"
)

type GuildService interface {
	CreateGuild(ctx context.Context, ownerID int64, req *CreateGuildRequest) (*GuildResponse, error)
}

type guildService struct {
	repo 	*GuildRepository
	log     *zap.Logger
}

func NewGuildService(repo *GuildRepository, log *zap.Logger) *guildService {
	return &guildService{repo: repo, log: log}
}

func (s *guildService) CreateGuild(ctx context.Context, ownerID int64, req *CreateGuildRequest) (*GuildResponse, error) {
	guildId, err := sonyflake.GenerateSonyFlakeID()
	if err != nil {
		s.log.Error("Failed to generate guild ID", zap.Error(err))
		return nil, errors.NewInternalServerError("Failed to create guild")
	}

	iconUrl := "" // TODO: Handle icon upload and get URL

	params := queries.CreateGuildParams{
		ID:      guildId,
		Name:    req.Name,
		OwnerID: ownerID,
		Icon: pgtype.Text{
			String: iconUrl,
			Valid:  iconUrl != "",
		},
		Description: pgtype.Text{String: req.Description, Valid: req.Description != ""},
	}

	guild, err := s.repo.CreateGuild(ctx, params)
	if err != nil {
		s.log.Error("Failed to create guild in repository", zap.Error(err))
		return nil, errors.NewInternalServerError("Failed to create guild")
	}

	return &GuildResponse{
		ID:          strconv.Itoa(int(guildId)),
		Name:        guild.Name,
		Description: guild.Description.String,
		CreatedAt:   guild.CreatedAt.Time.Format(time.RFC3339),
		Icon:        nil, // TODO: Return actual icon URL,
		OwnerID:     strconv.FormatInt(ownerID, 10),
	}, nil
}