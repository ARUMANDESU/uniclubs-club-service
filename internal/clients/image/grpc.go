package image

import (
	"context"
	"fmt"
	"github.com/ARUMANDESU/uniclubs-club-service/internal/config"
	imagev1 "github.com/ARUMANDESU/uniclubs-protos/gen/go/filestorage"
	grpclog "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	grpcretry "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/retry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"log/slog"
)

type Client struct {
	imagev1.ImageStorageClient
	log *slog.Logger
}

func New(
	ctx context.Context,
	log *slog.Logger,
	cfg config.ClientsConfig,
) (*Client, error) {
	const op = "grpc.New"

	retryOpts := []grpcretry.CallOption{
		grpcretry.WithCodes(codes.NotFound, codes.Aborted, codes.DeadlineExceeded),
		grpcretry.WithMax(uint(cfg.Image.RetriesCount)),
		grpcretry.WithPerRetryTimeout(cfg.Image.Timeout),
	}

	logOpts := []grpclog.Option{
		grpclog.WithLogOnEvents(grpclog.StartCall, grpclog.FinishCall),
	}

	cc, err := grpc.DialContext(ctx, cfg.Image.Address,
		grpc.WithTransportCredentials(insecure.NewCredentials()), // in the future, we can use tls/ssl cert if we want
		grpc.WithChainUnaryInterceptor(
			grpclog.UnaryClientInterceptor(InterceptorLogger(log), logOpts...),
			grpcretry.UnaryClientInterceptor(retryOpts...),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Client{
		ImageStorageClient: imagev1.NewImageStorageClient(cc),
		log:                log,
	}, nil
}

// InterceptorLogger adapts slog logger to interceptor logger
func InterceptorLogger(l *slog.Logger) grpclog.Logger {
	return grpclog.LoggerFunc(func(ctx context.Context, lvl grpclog.Level, msg string, fields ...any) {
		l.Log(ctx, slog.Level(lvl), msg, fields...)
	})
}
