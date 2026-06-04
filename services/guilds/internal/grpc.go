package internal

import (
	"context"

	pb "github.com/jakottelaar/relay-microservices/shared/proto/guilds"
	"go.uber.org/zap"
)

type GRPCServer struct {
	pb.UnimplementedGuildsServiceServer
	repo *GuildRepository
	log  *zap.Logger
}

func NewGRPCServer(repo *GuildRepository, log *zap.Logger) *GRPCServer {
	return &GRPCServer{
		repo: repo,
		log:  log,
	}
}

 
func (s *GRPCServer) GetChannelMembers(ctx context.Context, req *pb.GetChannelMembersRequest) (*pb.GetChannelMembersResponse, error) {
	s.log.Info("gRPC GetChannelMembers", zap.Int64("channel_id", req.ChannelId))
 
	userIDs, err := s.repo.GetMemberIDsByChannelID(ctx, req.ChannelId)
	if err != nil {
		s.log.Error("Failed to get channel members", zap.Error(err))
		return nil, err
	}
 
	return &pb.GetChannelMembersResponse{UserIds: userIDs}, nil
}