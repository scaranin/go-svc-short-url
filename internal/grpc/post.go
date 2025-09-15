package grpc

import (
	"context"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/scaranin/go-svc-short-url/internal/gen"
	"github.com/scaranin/go-svc-short-url/internal/handlers"
	"github.com/scaranin/go-svc-short-url/internal/models"
)

// ShortenText handles requests to create a short URL from a plain text body.
// It delegates the core logic to the `post` helper, specifying `Text`
func (s *GRPCServer) ShortenText(ctx context.Context, req *gen.ShortenTextRequest) (*gen.ShortenTextResponse, error) {
	if req.Url == "" {
		return nil, status.Error(codes.InvalidArgument, "URL is required")
	}

	userID := s.getUserIDFromContext(ctx)

	shortURL := handlers.ShortURLCalc(req.Url)
	url := &models.URL{
		OriginalURL: req.Url,
		ShortURL:    shortURL,
		UserID:      userID,
	}

	shortURL, err := s.storage.Save(url)
	fmt.Println("shortURL", shortURL)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to save URL")
	}

	return &gen.ShortenTextResponse{ShortUrl: s.baseURL + "/" + shortURL}, nil
}

// ShortenText handles requests to create a short URL from a plain text body.
// It delegates the core logic to the `post` helper, specifying `JSON`
func (s *GRPCServer) ShortenJSON(ctx context.Context, req *gen.ShortenJSONRequest) (*gen.ShortenJSONResponse, error) {
	if req.Url == "" {
		return nil, status.Error(codes.InvalidArgument, "URL is required")
	}

	userID := s.getUserIDFromContext(ctx)
	url := &models.URL{
		OriginalURL: req.Url,
		UserID:      userID,
	}

	shortURL, err := s.storage.Save(url)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to save URL")
	}

	return &gen.ShortenJSONResponse{Result: s.baseURL + "/" + shortURL}, nil
}

// ShortenBatch requests to shorten multiple URLs in a single batch operation.
func (s *GRPCServer) ShortenBatch(ctx context.Context, req *gen.BatchShortenRequest) (*gen.BatchShortenResponse, error) {
	if len(req.Items) == 0 {
		return nil, status.Error(codes.InvalidArgument, "batch items are required")
	}

	userID := s.getUserIDFromContext(ctx)
	response := &gen.BatchShortenResponse{}

	for _, item := range req.Items {
		url := &models.URL{
			CorrelationID: item.CorrelationId,
			OriginalURL:   item.OriginalUrl,
			UserID:        userID,
		}

		shortURL, err := s.storage.Save(url)
		if err != nil {
			return nil, status.Error(codes.Internal, "failed to save URL in batch")
		}

		response.Items = append(response.Items, &gen.BatchItemResponse{
			CorrelationId: item.CorrelationId,
			ShortUrl:      s.baseURL + "/" + shortURL,
		})
	}

	return response, nil
}
