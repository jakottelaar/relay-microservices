package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/jakottelaar/relay-microservices/services/messages/internal/queries"
	"github.com/jakottelaar/relay-microservices/shared/errors"
	"github.com/jakottelaar/relay-microservices/shared/events"
	"github.com/jakottelaar/relay-microservices/shared/sonyflake"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

type MessageService interface {
	CreateMessage(ctx context.Context, userID int64, channelID int64, req *CreateMessageRequest) (*MessageResponse, error)
	GetMessages(ctx context.Context, userID int64, channelID int64, beforeID int64, afterID int64, limit int32) ([]*MessageResponse, error)
}

type messageService struct {
	repo *MessageRepository
	nc  *nats.Conn
	log  *zap.Logger
}

func NewMessageService(repo *MessageRepository, nc *nats.Conn, log *zap.Logger) *messageService {
	return &messageService{
		repo: repo,
		nc: nc,
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

	event := &events.MessageCreatedEvent{
		MessageID: messageResp.ID,
		ChannelID: messageResp.ChannelID,
		AuthorID:  messageResp.AuthorID,
		Content:   messageResp.Content,
		CreatedAt: messageResp.CreatedAt,
	}

	data, err := json.Marshal(event)
	if err != nil {
		s.log.Error("Failed to marshal message created event", zap.Error(err))
		return nil, err
	}

	subject := fmt.Sprintf("messages.channels.%s.created", event.ChannelID)
	
	if err := s.nc.Publish(subject, data); err != nil {
		s.log.Error("Failed to publish message created event", zap.Error(err))
		return nil, err
	}

	s.log.Info("Message created event published",
		zap.String("subject", subject),
		zap.String("message_id", messageResp.ID),
	)

	return messageResp, nil
}

func (s *messageService) GetMessages(ctx context.Context, userID int64, channelID int64, beforeID int64, afterID int64, limit int32) ([]*MessageResponse, error) {
	if beforeID != 0 && afterID != 0 {
		return nil, errors.NewBadRequestError("Cannot use both before and after IDs")
	}

	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}

	s.log.Debug("Fetching messages",
		zap.Int64("channel_id", channelID),
		zap.Int64("user_id", userID),
	)

	messages, err := s.repo.GetChannelMessages(ctx, queries.GetChannelMessagesParams{
		ChannelID: channelID,
		BeforeID:  beforeID,
		AfterID:   afterID,
		LimitCount: limit,
	})
	if err != nil {
		s.log.Error("Failed to fetch messages", zap.Error(err))
		return nil, err
	}

	messageResponses := make([]*MessageResponse, 0, len(messages))
	for _, message := range messages {
		messageResponses = append(messageResponses, &MessageResponse{
			ID:        strconv.FormatInt(message.ID, 10),
			Content:   message.Content,
			AuthorID:  strconv.FormatInt(message.AuthorID, 10),
			ChannelID: strconv.FormatInt(message.ChannelID, 10),
			CreatedAt: message.CreatedAt.Time.Format(time.RFC3339),
		})
	}

	s.log.Debug("Messages fetched",
		zap.Int("count", len(messageResponses)),
	)

	return messageResponses, nil
}