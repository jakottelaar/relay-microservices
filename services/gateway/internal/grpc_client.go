package internal

import (
	"context"
	"fmt"
	"time"

	pb "github.com/jakottelaar/relay-microservices/shared/proto/guilds"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GuildsClient struct {
	client pb.GuildsServiceClient
	log    *zap.Logger
}

func NewGuildsClient(addr string, log *zap.Logger) (*GuildsClient, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("connect to guilds grpc: %w", err)
	}
	return &GuildsClient{
		client: pb.NewGuildsServiceClient(conn),
		log:    log,
	}, nil
}

func (g *GuildsClient) GetChannelMembers(channelID int64) ([]int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := g.client.GetChannelMembers(ctx, &pb.GetChannelMembersRequest{
		ChannelId: channelID,
	})
	if err != nil {
		g.log.Error("gRPC GetChannelMembers failed",
			zap.Int64("channel_id", channelID),
			zap.Error(err),
		)
		return nil, err
	}

	return resp.UserIds, nil
}