package handlers

import (
	"context"

	pb "github.com/ar4ie13/shortener/api/proto"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// authorizationInterceptor is authentication interceptor for gRPC requests
func (h Handler) authorizationInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (interface{}, error) {
	var (
		tokenString string
		userUUID    uuid.UUID
		err         error
	)
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		values := md.Get("user_id")
		if len(values) > 0 {
			tokenString = values[0]
		}
	}

	if tokenString == "" {
		// If no token - creating new userUUID and JWT token
		userUUID = h.auth.GenerateUserUUID()
		tokenString, err = h.auth.BuildJWTString(userUUID)
		if err != nil {
			h.zlog.Error().Msgf("Error building JWT string: %v", err)
			return nil, status.Errorf(codes.Internal, "Error building JWT string: %v", err)
		}
	} else {
		// Checking existing cookie
		userUUID, err = h.auth.ValidateUserUUID(tokenString)
		if err != nil {
			h.zlog.Error().Msgf("Error validating user UUID: %v", err)
			return nil, status.Errorf(codes.Unauthenticated, "Error validating user UUID: %v", err)
		}
	}

	// Store user_id in context for handlers to use
	ctx = context.WithValue(ctx, UserUUIDKey, userUUID.String())

	// Send user_id back to client in response headers for session persistence
	header := metadata.Pairs("user_id", userUUID.String())
	if err := grpc.SendHeader(ctx, header); err != nil {
		h.zlog.Error().Msgf("Error sending header: %v", err)
	}
	return handler(ctx, req)
}

// getUserUUIDFromGRPCRequest gets userUUID from gRPC request
func (h Handler) getUserUUIDFromGRPCRequest(ctx context.Context) (uuid.UUID, error) {
	userIDValue := ctx.Value(UserUUIDKey)
	if userIDValue == nil {
		h.zlog.Debug().Msg("user_id not found in context")
		return uuid.Nil, status.Error(codes.Unauthenticated, "user_id not found in context")
	}

	userIDStr, ok := userIDValue.(string)
	if !ok {
		h.zlog.Debug().Msg("user_id is not a string")
		return uuid.Nil, status.Error(codes.Internal, "user_id is not a string")
	}

	userUUID, err := uuid.Parse(userIDStr)
	if err != nil {
		h.zlog.Debug().Msgf("cannot parse user UUID: %v", err)
		return uuid.Nil, status.Error(codes.InvalidArgument, "invalid user UUID format")
	}

	return userUUID, nil
}

// ShortenURL receives URL from gPRC request to store it in the Repository via Service
func (h Handler) ShortenURL(ctx context.Context, in *pb.URLShortenRequest) (*pb.URLShortenResponse, error) {
	userUUID, err := h.getUserUUIDFromGRPCRequest(ctx)
	if err != nil {
		h.zlog.Error().Msgf("Error getting user UUID: %v", err)
		return nil, status.Errorf(codes.Aborted, "Error getting user UUID: %v", err)
	}
	var response pb.URLShortenResponse
	slug, err := h.service.SaveURL(ctx, userUUID, in.GetUrl())
	if err != nil {
		h.zlog.Error().Msgf("Error saving URL: %v", err)
		return nil, status.Errorf(codes.Aborted, "Error saving URL: %v", err)
	}
	response.SetResult(slug)

	return &response, nil
}

// ExpandURL handles gRPC request with provided shortURL and redirects to the URL if it is found in Repository
func (h Handler) ExpandURL(ctx context.Context, in *pb.URLExpandRequest) (*pb.URLExpandResponse, error) {
	userUUID, err := h.getUserUUIDFromGRPCRequest(ctx)
	if err != nil {
		h.zlog.Error().Msgf("Error getting user UUID: %v", err)
		return nil, status.Errorf(codes.Aborted, "Error getting user UUID: %v", err)
	}
	var response pb.URLExpandResponse
	url, err := h.service.GetURL(ctx, userUUID, in.GetId())
	if err != nil {
		h.zlog.Error().Msgf("Error getting URL: %v", err)
		return nil, status.Errorf(codes.Aborted, "Error getting URL: %v", err)
	}
	response.SetResult(url)

	return &response, nil
}

func (h Handler) ListUserURLs(ctx context.Context, _ *emptypb.Empty) (*pb.UserURLsResponse, error) {
	userUUID, err := h.getUserUUIDFromGRPCRequest(ctx)
	if err != nil {
		h.zlog.Error().Msgf("Error getting user UUID: %v", err)
		return nil, status.Errorf(codes.Aborted, "Error getting user UUID: %v", err)
	}
	userSlugsMap, err := h.service.GetUserShortURLs(ctx, userUUID)
	if err != nil {
		h.zlog.Error().Msgf("Error getting user URLs: %v", err)
		return nil, status.Errorf(codes.Aborted, "Error getting user URLs: %v", err)
	}

	userSlugs := make([]*pb.URLData, 0, len(userSlugsMap))
	for shortURL, originalURL := range userSlugsMap {
		urlData := &pb.URLData{}
		urlData.SetShortUrl(shortURL)
		urlData.SetOriginalUrl(originalURL)
		userSlugs = append(userSlugs, urlData)
	}

	response := pb.UserURLsResponse_builder{
		Url: userSlugs,
	}

	return response.Build(), nil
}
