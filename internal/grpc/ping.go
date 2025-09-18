package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/scaranin/go-svc-short-url/internal/gen"
)

// Ping serves as an gRPC to check the health of the database connection.
// It attempts to create a new connection pool using the DSN from the handler
// and then pings the database to verify connectivity.
//
// On success, creating the connection pool and return success: true
// or the ping fails, it responds success: false.
func (s *GRPCServer) Ping(ctx context.Context, req *emptypb.Empty) (*gen.PingResponse, error) {
	err := s.storage.Ping(ctx)
	if err != nil {
		return &gen.PingResponse{
			Success: false,
			Message: "Database connection failed: " + err.Error(),
		}, status.Error(codes.Internal, "database connection failed")
	}

	return &gen.PingResponse{
		Success: true,
		Message: "Database connection successful",
	}, nil
}
