package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/scaranin/go-svc-short-url/internal/gen"
)

// GetOriginal handles GET requests for short URLs, redirecting clients to the original URL.
// It extracts the `shortURL` from the path parameter.
func (s *GRPCServer) GetOriginal(ctx context.Context, req *gen.GetOriginalRequest) (*gen.GetOriginalResponse, error) {
	if req.ShortUrl == "" {
		return nil, status.Error(codes.InvalidArgument, "short URL is required")
	}

	// Извлекаем только ID из полного URL если нужно
	shortURLID := extractShortURLID(req.ShortUrl)

	originalURL, err := s.storage.Load(shortURLID)
	if err != nil {
		if err.Error() == "ROW_IS_DELETED" {
			return nil, status.Error(codes.NotFound, "URL was deleted")
		}
		return nil, status.Error(codes.Internal, "failed to load URL")
	}

	return &gen.GetOriginalResponse{OriginalUrl: originalURL}, nil
}

// GetUserURLs is an HTTP handler that retrieves all URLs created by the currently authenticated user.
func (s *GRPCServer) GetUserURLs(ctx context.Context, req *emptypb.Empty) (*gen.GetUserURLsResponse, error) {
	userID := s.getUserIDFromContext(ctx)
	if userID == "" {
		return nil, status.Error(codes.Unauthenticated, "authentication required")
	}

	urlList, err := s.storage.GetUserURLList(userID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get user URLs")
	}

	if len(urlList) == 0 {
		return nil, status.Error(codes.NotFound, "no URLs found for user")
	}

	response := &gen.GetUserURLsResponse{}
	for _, item := range urlList {
		response.Items = append(response.Items, &gen.UserURLItem{
			ShortUrl:    s.baseURL + "/" + item.ShortURL,
			OriginalUrl: item.OriginalURL,
		})
	}

	return response, nil
}
