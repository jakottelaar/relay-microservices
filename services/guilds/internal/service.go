package internal

import (
	"context"
	"fmt"
	"mime/multipart"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jakottelaar/relay-microservices/services/guilds/internal/queries"
	"github.com/jakottelaar/relay-microservices/shared/errors"
	"github.com/jakottelaar/relay-microservices/shared/sonyflake"
	"go.uber.org/zap"
)

type GuildService interface {
    CreateGuild(ctx context.Context, ownerID int64, req *CreateGuildRequest, icon *multipart.FileHeader) (*GuildResponse, error)
    GetGuild(ctx context.Context, guildID int64) (*GuildResponse, error)
    CreateGuildChannel(ctx context.Context, guildID int64, req *CreateGuildChannelRequest) (*GuildChannelResponse, error)
    CreateGuildMember(ctx context.Context, guildID int64, userID int64, nick *string) (*GuildMemberResponse, error)
    GetGuildChannels(ctx context.Context, guildID int64) ([]*GuildChannelResponse, error)
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

    description := ""
    if req.Description != nil {
        description = *req.Description
    }
    params := queries.CreateGuildParams{
        ID:          guildID,
        Name:        req.Name,
        OwnerID:     ownerID,
        Icon:        iconPath,
        Description: pgtype.Text{String: description, Valid: description != ""},
    }

    var guild queries.Guild
    err = s.repo.WithTx(ctx, func(q *queries.Queries) error {
        createdGuild, err := q.CreateGuild(ctx, params)
        if err != nil {
            s.log.Error("Failed to create guild in repository", zap.Error(err))
            return err
        }
        guild = createdGuild

        _, err = q.CreateGuildMember(ctx, queries.CreateGuildMemberParams{
            GuildID: guildID,
            UserID:  guild.OwnerID,
        })
        if err != nil {
            s.log.Error("Failed to create guild member for owner", zap.Error(err))
            return err
        }

        return nil
    })
    
    if err != nil {
        return nil, errors.NewInternalServerError("Failed to create guild")
    }

    var desc *string
    if guild.Description.Valid {
        desc = &guild.Description.String
    }

    return &GuildResponse{
        ID:          strconv.FormatInt(guildID, 10),
        Name:        guild.Name,
        Description: desc,
        Icon:        s.buildIconURL(guild.Icon),
        OwnerID:     strconv.FormatInt(ownerID, 10),
        CreatedAt:   guild.CreatedAt.Time.Format(time.RFC3339),
    }, nil
}

func (s *guildService) GetGuild(ctx context.Context, guildID int64) (*GuildResponse, error) {
    guild, err := s.repo.GetGuild(ctx, guildID)
    if err != nil {
        if err == pgx.ErrNoRows {
            return nil, errors.NewNotFoundError("Guild not found")
        }
        s.log.Error("Failed to get guild from repository", zap.Error(err))
        return nil, errors.NewInternalServerError("Failed to get guild")
    }
    
    var desc *string
    if guild.Description.Valid {
        desc = &guild.Description.String
    }
    
    return &GuildResponse{
        ID:          strconv.FormatInt(guild.ID, 10),
        Name:        guild.Name,
        Description: desc,
        Icon:        s.buildIconURL(guild.Icon),
        OwnerID:     strconv.FormatInt(guild.OwnerID, 10),
        CreatedAt:   guild.CreatedAt.Time.Format(time.RFC3339),
        UpdatedAt:   guild.UpdatedAt.Time.Format(time.RFC3339),
    }, nil
}

func (s *guildService) CreateGuildChannel(ctx context.Context, guildID int64, req *CreateGuildChannelRequest) (*GuildChannelResponse, error) {
    existingGuild, err := s.repo.GetGuild(ctx, guildID)
    if err != nil {
        if err == pgx.ErrNoRows {
            return nil, errors.NewNotFoundError("Guild not found")
        }
        s.log.Error("Failed to get guild from repository", zap.Error(err))
        return nil, errors.NewInternalServerError("Failed to get guild")
    }


    channelID, err := sonyflake.GenerateSonyFlakeID()
    if err != nil {
        s.log.Error("Failed to generate channel ID", zap.Error(err))
        return nil, errors.NewInternalServerError("Failed to create channel")
    }

    channelType := req.Type
    if channelType == 0 {
        channelType = TextChannel
    }

    reqTopic := ""
    if req.Topic != nil {
        reqTopic = *req.Topic
    }

    params := queries.CreateGuildChannelParams{
        ID:      channelID,
        GuildID: existingGuild.ID,
        Name:    req.Name,
        Type:    int16(channelType),
        Topic:   pgtype.Text{String: reqTopic, Valid: reqTopic != ""},
    }

    channel, err := s.repo.CreateGuildChannel(ctx, params)
    if err != nil {
        s.log.Error("Failed to create guild channel in repository", zap.Error(err))
        return nil, errors.NewInternalServerError("Failed to create channel")
    }

    var topic *string
    if channel.Topic.Valid {
        topic = &channel.Topic.String
    }

    return &GuildChannelResponse{
        ID:        strconv.FormatInt(channel.ID, 10),
        Name:      channel.Name,
        Type:      ChannelType(channel.Type),
        Topic:     topic,
        GuildID:   strconv.FormatInt(existingGuild.ID, 10),
        CreatedAt: channel.CreatedAt.Time.Format(time.RFC3339),
        UpdatedAt: channel.UpdatedAt.Time.Format(time.RFC3339),
    }, nil
}

func (s *guildService) CreateGuildMember(ctx context.Context, guildID int64, userID int64, nick *string) (*GuildMemberResponse, error) {
    guild, err := s.repo.GetGuild(ctx, guildID)
    if err != nil {
        if err == pgx.ErrNoRows {
            return nil, errors.NewNotFoundError("Guild not found")
        }
        s.log.Error("Failed to get guild", zap.Error(err))
        return nil, errors.NewInternalServerError("Failed to create guild member")
    }

    if guild.OwnerID == userID {
        return nil, errors.NewBadRequestError("Guild owner cannot be added as a member")
    }

    reqNick := ""
    if nick != nil {
        reqNick = *nick
    }

    member, err := s.repo.CreateGuildMember(ctx, queries.CreateGuildMemberParams{
        GuildID: guildID,
        UserID:  userID,
        Nick:    pgtype.Text{String: reqNick, Valid: reqNick != ""},
    })
    if err != nil {
        if errors.IsUniqueViolation(err) {
            return nil, errors.NewDuplicateError("User is already a member of this guild")
        }
        s.log.Error("Failed to create guild member", zap.Error(err))
        return nil, errors.NewInternalServerError("Failed to create guild member")
    }

    var memberNick *string
    if member.Nick.Valid {
        memberNick = &member.Nick.String
    }

    return &GuildMemberResponse{
        GuildID:  strconv.FormatInt(guildID, 10),
        UserID:   strconv.FormatInt(userID, 10),
        Nick:     memberNick,
        JoinedAt: member.JoinedAt.Time.Format(time.RFC3339),
    }, nil
}

func (s *guildService) GetGuildChannels(ctx context.Context, guildID int64) ([]*GuildChannelResponse, error) {
    channels, err := s.repo.GetGuildChannels(ctx, guildID)
    if err != nil {
        if err == pgx.ErrNoRows {
            return nil, errors.NewNotFoundError("Guild not found")
        }
        s.log.Error("Failed to get guild channels from repository", zap.Error(err))
        return nil, errors.NewInternalServerError("Failed to get guild channels")
    }

    var channelResponses []*GuildChannelResponse
    for _, channel := range channels {
        var topic *string
        if channel.Topic.Valid {
            topic = &channel.Topic.String
        }

        channelResponses = append(channelResponses, &GuildChannelResponse{
            ID:        strconv.FormatInt(channel.ID, 10),
            Name:      channel.Name,
            Type:      ChannelType(channel.Type),
            Topic:     topic,
            GuildID:   strconv.FormatInt(guildID, 10),
            CreatedAt: channel.CreatedAt.Time.Format(time.RFC3339),
            UpdatedAt: channel.UpdatedAt.Time.Format(time.RFC3339),
        })
    }

    return channelResponses, nil
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