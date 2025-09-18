package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/scaranin/go-svc-short-url/internal/gen"
)

// DeleteUserURLs processes short URLs for deletion from a channel. It acts as a worker
// that consumes short URL strings from the provided channel. For each URL received,
// it calls the `DeleteBulk` method of the storage layer, associating the deletion
// with the `UserID` from the handler's context. This function is designed to run
// in a separate goroutine to handle deletion tasks asynchronously.
func (s *GRPCServer) DeleteUserURLs(ctx context.Context, req *gen.DeleteUserURLsRequest) (*emptypb.Empty, error) {
	userID := s.getUserIDFromContext(ctx)
	if userID == "" {
		return nil, status.Error(codes.Unauthenticated, "authentication required")
	}

	if len(req.ShortUrls) == 0 {
		return nil, status.Error(codes.InvalidArgument, "short URLs are required")
	}

	var shortURLIDs []string
	for _, shortURL := range req.ShortUrls {
		shortURLIDs = append(shortURLIDs, extractShortURLID(shortURL))
	}

	err := s.storage.DeleteBulk(userID, shortURLIDs)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to delete URLs")
	}

	return &emptypb.Empty{}, nil
}
