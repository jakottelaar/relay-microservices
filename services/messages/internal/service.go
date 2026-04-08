package internal

import (
	"context"
	"strconv"
	"time"

	"github.com/jakottelaar/relay-microservices/services/messages/internal/queries"
	"github.com/jakottelaar/relay-microservices/shared/sonyflake"
	"go.uber.org/zap"
)

type MessageService interface {
	CreateMessage(ctx context.Context, userID int64, channelID int64, req *CreateMessageRequest) (*MessageResponse, error)
}

type messageService struct {
	repo *MessageRepository
	log  *zap.Logger
}

func NewMessageService(repo *MessageRepository, log *zap.Logger) *messageService {
	return &messageService{
		repo: repo,
		log:  log,
	}
}

func (s *messageService) CreateMessage(ctx context.Context, userID int64, channelID int64, req *CreateMessageRequest) (*MessageResponse, error) {
	s.log.Info("Creating message in service",
		zap.Int64("channel_id", channelID),
		zap.Int64("user_id", userID),
	)

	messageID, err := sonyflake.GenerateSonyFlakeID()
	if err != nil {
		s.log.Error("Failed to generate message ID", zap.Error(err))
		return nil, err
	}

	message, err := s.repo.CreateMessage(ctx, queries.CreateMessageParams{
		ID:        messageID,
		Content:   *req.Content,
		AuthorID:  userID,
		ChannelID: channelID,
	})
	if err != nil {
		s.log.Error("Failed to create message in repository", zap.Error(err))
		return nil, err
	}

	messageResp := &MessageResponse{
		ID:        strconv.FormatInt(message.ID, 10),
		Content:   message.Content,
		AuthorID:  strconv.FormatInt(message.AuthorID, 10),
		ChannelID: strconv.FormatInt(message.ChannelID, 10),
		CreatedAt: message.CreatedAt.Time.Format(time.RFC3339),
	}

	s.log.Info("Message created in service",
		zap.String("message_id", messageResp.ID),
		zap.Int64("channel_id", channelID),
	)

	return messageResp, nil
}