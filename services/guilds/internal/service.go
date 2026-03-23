package internal

import (
	"context"
	"fmt"
	"mime/multipart"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jakottelaar/relay-microservices/services/guilds/internal/queries"
	"github.com/jakottelaar/relay-microservices/shared/errors"
	"github.com/jakottelaar/relay-microservices/shared/sonyflake"
	"go.uber.org/zap"
)

type GuildService interface {
    CreateGuild(ctx context.Context, ownerID int64, req *CreateGuildRequest, icon *multipart.FileHeader) (*GuildResponse, error)
}

type guildService struct {
    repo    *GuildRepository
    storage *GuildStorage
    log     *zap.Logger
}

func NewGuildService(repo *GuildRepository, storage *GuildStorage, log *zap.Logger) *guildService {
    return &guildService{
        repo: repo, 
        storage: storage, 
        log: log,
    }
}

func (s *guildService) CreateGuild(ctx context.Context, ownerID int64, req *CreateGuildRequest, icon *multipart.FileHeader) (*GuildResponse, error) {
    guildID, err := sonyflake.GenerateSonyFlakeID()
    if err != nil {
        s.log.Error("Failed to generate guild ID", zap.Error(err))
        return nil, errors.NewInternalServerError("Failed to create guild")
    }

    var iconPath pgtype.Text
    if icon != nil {
        path, err := s.storage.UploadGuildIcon(ctx, guildID, icon)
        if err != nil {
            s.log.Error("Failed to upload guild icon", zap.Error(err))
            return nil, errors.NewInternalServerError("Failed to upload guild icon")
        }
        iconPath = pgtype.Text{String: path, Valid: true}
    }

    params := queries.CreateGuildParams{
        ID:          guildID,
        Name:        req.Name,
        OwnerID:     ownerID,
        Icon:        iconPath,
        Description: pgtype.Text{String: req.Description, Valid: req.Description != ""},
    }

    guild, err := s.repo.CreateGuild(ctx, params)
    if err != nil {
        s.log.Error("Failed to create guild in repository", zap.Error(err))
        return nil, errors.NewInternalServerError("Failed to create guild")
    }

    return &GuildResponse{
        ID:          strconv.FormatInt(guildID, 10),
        Name:        guild.Name,
        Description: guild.Description.String,
        CreatedAt:   guild.CreatedAt.Time.Format(time.RFC3339),
        Icon:        s.buildIconURL(guild.Icon),
        OwnerID:     strconv.FormatInt(ownerID, 10),
    }, nil
}

func (s *guildService) buildIconURL(icon pgtype.Text) *string {
    if !icon.Valid {
        return nil
    }
    url := fmt.Sprintf("%s/%s", 
        s.storage.baseURL,
        icon.String,
    )
    return &url
}