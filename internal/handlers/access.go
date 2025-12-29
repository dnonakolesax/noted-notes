package handlers

import (
	"context"
	"fmt"
	"log/slog"

	pbAccess "github.com/dnonakolesax/noted-notes/internal/handlers/proto"
	"google.golang.org/grpc/metadata"
)

type AccessUsecase interface {
	Get(fileID string, userID string, byBlock bool) (string, error)
}

type Server struct {
	pbAccess.UnimplementedAcessServiceServer

	logger        *slog.Logger
	accessService AccessUsecase
}

func NewAccessServer(accessService AccessUsecase) *Server {
	return &Server{
		accessService: accessService,
		logger:        slog.Default(),
	}
}

func (s *Server) FileAccessCtx(ctx context.Context, req *pbAccess.AccessRequest) (*pbAccess.AccessData, error) {
	md, ok := metadata.FromIncomingContext(ctx)

	if !ok {
		s.logger.ErrorContext(ctx, "Couldn't parse request metadata")
		return nil, fmt.Errorf("Couldn't parse request metadata")
	}
	traceID := md["trace_id"]
	s.logger.DebugContext(ctx, "got request", slog.String("trace", traceID[0]))
	access, err := s.accessService.Get(req.FileID, req.UserID, false)

	if err != nil {
		s.logger.ErrorContext(ctx, "error while getting access")
		return nil, err
	}

	return &pbAccess.AccessData{
		Access: access,
	}, nil
}
