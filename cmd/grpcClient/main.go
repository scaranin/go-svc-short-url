package main

import (
	"context"
	"log"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/scaranin/go-svc-short-url/internal/gen"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("Error: %v", err)
		os.Exit(1)
	}
}

type Claims struct {
	jwt.RegisteredClaims
	UserID string
}

func BuildJWTString() (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
		UserID: uuid.New().String(),
	})

	authToken, err := token.SignedString([]byte("TsoyZhiv"))
	if err != nil {
		return "", err
	}

	return authToken, nil
}

func run() error {
	conn, err := grpc.NewClient("localhost:9090", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	defer conn.Close()

	client := gen.NewShortenerServiceClient(conn)
	token, err := BuildJWTString()
	if err != nil {
		log.Println("Error: ", err)
	}
	token = "auth_token=" + token
	md := metadata.Pairs("cookie", token)
	ctx := metadata.NewOutgoingContext(context.Background(), md)

	log.Println("Testing gRPC server connection...")

	if err := testPing(client, ctx); err != nil {
		return err
	}

	shortURL, err := testShorten(client, ctx)
	if err != nil {
		return err
	}

	if err := testGetOriginal(client, ctx, shortURL); err != nil {
		return err
	}

	if err := testStats(client, ctx); err != nil {
		return err
	}

	log.Println("All tests passed successfully!")
	return nil
}

func testPing(client gen.ShortenerServiceClient, ctx context.Context) error {
	resp, err := client.Ping(ctx, &emptypb.Empty{})
	if err != nil {
		return err
	}
	log.Printf("Ping successful: %v", resp.Success)
	return nil
}

func testShorten(client gen.ShortenerServiceClient, ctx context.Context) (string, error) {
	resp, err := client.ShortenText(ctx, &gen.ShortenTextRequest{
		Url: "https://yango.com",
	})
	if err != nil {
		return "", err
	}
	log.Printf("URL shortened: %s", resp.ShortUrl)
	return resp.ShortUrl, nil
}

func testGetOriginal(client gen.ShortenerServiceClient, ctx context.Context, shortURL string) error {
	resp, err := client.GetOriginal(ctx, &gen.GetOriginalRequest{
		ShortUrl: shortURL,
	})
	if err != nil {
		return err
	}
	log.Printf("Original URL retrieved: %s", resp.OriginalUrl)
	return nil
}

func testStats(client gen.ShortenerServiceClient, ctx context.Context) error {
	resp, err := client.GetStats(ctx, &emptypb.Empty{})
	if err != nil {
		return err
	}
	log.Printf("Statistics: URLs=%d, Users=%d", resp.Urls, resp.Users)
	return nil
}
