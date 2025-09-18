package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/scaranin/go-svc-short-url/internal/gen"
)

// GetStats retrieving and returning storage statistics.
// It performs authentication via cookies, verifies the client's IP against a trusted subnet
func (s *GRPCServer) GetStats(ctx context.Context, req *emptypb.Empty) (*gen.StatsResponse, error) {
	stats, err := s.storage.GetStats()
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get statistics")
	}

	return &gen.StatsResponse{
		Urls:  int32(stats.URLs),
		Users: int32(stats.Users),
	}, nil
}
