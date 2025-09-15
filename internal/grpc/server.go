package grpc

import (
	"context"
	"fmt"
	"strings"

	"github.com/scaranin/go-svc-short-url/internal/gen"
	"google.golang.org/grpc/metadata"

	"github.com/scaranin/go-svc-short-url/internal/models"
)

// GRPCServer implements the gRPC server for URL shortening service.
// Provides methods for shortening URLs, retrieving original URLs,
// managing user URLs, and retrieving statistics.
type GRPCServer struct {
	gen.UnimplementedShortenerServiceServer
	storage models.Storage
	baseURL string
	auth    AuthService
}

// AuthService defines the interface for authentication service.
// Used to extract user ID from cookies.
type AuthService interface {
	GetUserIDFromCookie(cookie string) (string, error)
}

// NewGRPCServer creates a new instance of GRPCServer.
func NewGRPCServer(storage models.Storage, baseURL string, auth AuthService) *GRPCServer {
	return &GRPCServer{
		storage: storage,
		baseURL: baseURL,
		auth:    auth,
	}
}

// getUserIDFromContext extracts user ID from gRPC request context.
// Analyzes request metadata to get cookies and extracts user ID via AuthService.
func (s *GRPCServer) getUserIDFromContext(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}

	cookies := md.Get("cookie")
	if len(cookies) == 0 {
		return ""
	}

	fmt.Println("cookies", cookies)

	userID, err := s.auth.GetUserIDFromCookie(cookies[0])
	if err != nil {
		return ""
	}

	return userID
}

// extractShortURLID extracts short URL identifier from full URL.
// Removes base URL and returns only the path suffix.
func extractShortURLID(fullURL string) string {
	parts := strings.Split(fullURL, "/")
	return parts[len(parts)-1]
}
