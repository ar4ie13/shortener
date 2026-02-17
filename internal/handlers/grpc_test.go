package handlers

import (
	"context"
	"testing"

	pb "github.com/ar4ie13/shortener/api/proto"
	"github.com/ar4ie13/shortener/internal/model"
	"github.com/ar4ie13/shortener/internal/myerrors"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// Mock implementations for testing
type mockGRPCService struct {
	urls map[string]string
}

func newMockGRPCService() *mockGRPCService {
	return &mockGRPCService{
		urls: map[string]string{
			"abc123": "https://example.com",
			"xyz789": "https://test.com",
		},
	}
}

func (m *mockGRPCService) GetURL(_ context.Context, _ uuid.UUID, id string) (string, error) {
	if url, ok := m.urls[id]; ok {
		return url, nil
	}
	return "", myerrors.ErrNotFound
}

func (m *mockGRPCService) SaveURL(_ context.Context, _ uuid.UUID, url string) (string, error) {
	if url == "https://example.com" {
		return "abc123", nil
	}
	if url == "https://duplicate.com" {
		return "dup456", myerrors.ErrURLExist
	}
	return "new789", nil
}

func (m *mockGRPCService) SaveBatch(_ context.Context, _ uuid.UUID, batch []model.URL) ([]model.URL, error) {
	return batch, nil
}

func (m *mockGRPCService) GetUserShortURLs(_ context.Context, _ uuid.UUID) (map[string]string, error) {
	return m.urls, nil
}

func (m *mockGRPCService) SendShortURLForDelete(_ context.Context, _ uuid.UUID, _ []string) {}

func (m *mockGRPCService) GetStats(_ context.Context) (int, int, error) {
	return 10, 5, nil
}

type mockGRPCAuth struct {
	shouldFailValidation bool
}

func (m *mockGRPCAuth) GenerateUserUUID() uuid.UUID {
	return uuid.MustParse("12345678-1234-1234-1234-123456789abc")
}

func (m *mockGRPCAuth) BuildJWTString(_ uuid.UUID) (string, error) {
	return "test-jwt-token", nil
}

func (m *mockGRPCAuth) ValidateUserUUID(token string) (uuid.UUID, error) {
	if m.shouldFailValidation {
		return uuid.Nil, myerrors.ErrInvalidUserUUID
	}
	return uuid.MustParse("87654321-4321-4321-4321-cba987654321"), nil
}

type mockGRPCConfig struct{}

func (m mockGRPCConfig) GetLocalServerAddr() string      { return ":8080" }
func (m mockGRPCConfig) GetShortURLTemplate() string     { return "http://localhost:8080" }
func (m mockGRPCConfig) GetLogLevel() zerolog.Level      { return zerolog.InfoLevel }
func (m mockGRPCConfig) CheckPostgresConnection(_ context.Context) error { return nil }
func (m mockGRPCConfig) GetHTTPS() bool                  { return false }
func (m mockGRPCConfig) GetTLSCertPath() string          { return "" }
func (m mockGRPCConfig) GetTLSKeyPath() string           { return "" }
func (m mockGRPCConfig) GetTrustedSubnet() string        { return "192.168.31.0/24" }
func (m mockGRPCConfig) GetGRPCServerAddr() string       { return "localhost:8081" }
func (m mockGRPCConfig) GetGRPCEnabled() bool            { return true }

type mockGRPCAuditor struct{}

func (m mockGRPCAuditor) Notify(_ string, _ uuid.UUID, _ string) {}

// Helper to create test handler
func newTestGRPCHandler() *Handler {
	return &Handler{
		service:  newMockGRPCService(),
		cfg:      mockGRPCConfig{},
		auth:     &mockGRPCAuth{},
		observer: mockGRPCAuditor{},
		zlog:     zerolog.Nop(),
	}
}

// Test authorizationInterceptor with no token
func TestAuthorizationInterceptor_NoToken(t *testing.T) {
	handler := newTestGRPCHandler()
	ctx := context.Background()

	called := false
	mockHandler := func(ctx context.Context, req interface{}) (interface{}, error) {
		called = true
		// Verify user_id was set in context
		userID := ctx.Value("user_id")
		assert.NotNil(t, userID)
		return "success", nil
	}

	resp, err := handler.authorizationInterceptor(ctx, nil, &grpc.UnaryServerInfo{}, mockHandler)

	require.NoError(t, err)
	assert.True(t, called)
	assert.Equal(t, "success", resp)
}

// Test authorizationInterceptor with valid token
func TestAuthorizationInterceptor_ValidToken(t *testing.T) {
	handler := newTestGRPCHandler()

	// Create context with metadata containing user_id
	md := metadata.New(map[string]string{"user_id": "valid-token"})
	ctx := metadata.NewIncomingContext(context.Background(), md)

	called := false
	mockHandler := func(ctx context.Context, req interface{}) (interface{}, error) {
		called = true
		userID := ctx.Value("user_id")
		assert.NotNil(t, userID)
		return "success", nil
	}

	resp, err := handler.authorizationInterceptor(ctx, nil, &grpc.UnaryServerInfo{}, mockHandler)

	require.NoError(t, err)
	assert.True(t, called)
	assert.Equal(t, "success", resp)
}

// Test authorizationInterceptor with invalid token
func TestAuthorizationInterceptor_InvalidToken(t *testing.T) {
	handler := newTestGRPCHandler()
	handler.auth = &mockGRPCAuth{shouldFailValidation: true}

	md := metadata.New(map[string]string{"user_id": "invalid-token"})
	ctx := metadata.NewIncomingContext(context.Background(), md)

	mockHandler := func(ctx context.Context, req interface{}) (interface{}, error) {
		t.Fatal("handler should not be called with invalid token")
		return nil, nil
	}

	resp, err := handler.authorizationInterceptor(ctx, nil, &grpc.UnaryServerInfo{}, mockHandler)

	require.Error(t, err)
	assert.Nil(t, resp)

	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Unauthenticated, st.Code())
}

// Test getUserUUIDFromGRPCRequest with valid context
func TestGetUserUUIDFromGRPCRequest_Valid(t *testing.T) {
	handler := newTestGRPCHandler()
	expectedUUID := uuid.MustParse("12345678-1234-1234-1234-123456789abc")

	ctx := context.WithValue(context.Background(), "user_id", expectedUUID.String())

	userUUID, err := handler.getUserUUIDFromGRPCRequest(ctx)

	require.NoError(t, err)
	assert.Equal(t, expectedUUID, userUUID)
}

// Test getUserUUIDFromGRPCRequest with missing user_id
func TestGetUserUUIDFromGRPCRequest_Missing(t *testing.T) {
	handler := newTestGRPCHandler()
	ctx := context.Background()

	userUUID, err := handler.getUserUUIDFromGRPCRequest(ctx)

	require.Error(t, err)
	assert.Equal(t, uuid.Nil, userUUID)

	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Unauthenticated, st.Code())
}

// Test getUserUUIDFromGRPCRequest with invalid type
func TestGetUserUUIDFromGRPCRequest_InvalidType(t *testing.T) {
	handler := newTestGRPCHandler()
	ctx := context.WithValue(context.Background(), "user_id", 12345) // wrong type

	userUUID, err := handler.getUserUUIDFromGRPCRequest(ctx)

	require.Error(t, err)
	assert.Equal(t, uuid.Nil, userUUID)

	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Internal, st.Code())
}

// Test getUserUUIDFromGRPCRequest with invalid UUID format
func TestGetUserUUIDFromGRPCRequest_InvalidFormat(t *testing.T) {
	handler := newTestGRPCHandler()
	ctx := context.WithValue(context.Background(), "user_id", "not-a-uuid")

	userUUID, err := handler.getUserUUIDFromGRPCRequest(ctx)

	require.Error(t, err)
	assert.Equal(t, uuid.Nil, userUUID)

	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, st.Code())
}

// Test ShortenURL with valid request
func TestShortenURL_Success(t *testing.T) {
	handler := newTestGRPCHandler()
	userUUID := uuid.MustParse("12345678-1234-1234-1234-123456789abc")
	ctx := context.WithValue(context.Background(), "user_id", userUUID.String())

	req := &pb.URLShortenRequest{}
	req.SetUrl("https://example.com")

	resp, err := handler.ShortenURL(ctx, req)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "abc123", resp.GetResult())
}

// Test ShortenURL without user_id in context
func TestShortenURL_NoUserID(t *testing.T) {
	handler := newTestGRPCHandler()
	ctx := context.Background()

	req := &pb.URLShortenRequest{}
	req.SetUrl("https://example.com")

	resp, err := handler.ShortenURL(ctx, req)

	require.Error(t, err)
	assert.Nil(t, resp)

	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Aborted, st.Code())
}

// Test ExpandURL with valid request
func TestExpandURL_Success(t *testing.T) {
	handler := newTestGRPCHandler()
	userUUID := uuid.MustParse("12345678-1234-1234-1234-123456789abc")
	ctx := context.WithValue(context.Background(), "user_id", userUUID.String())

	req := &pb.URLExpandRequest{}
	req.SetId("abc123")

	resp, err := handler.ExpandURL(ctx, req)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "https://example.com", resp.GetResult())
}

// Test ExpandURL with non-existent ID
func TestExpandURL_NotFound(t *testing.T) {
	handler := newTestGRPCHandler()
	userUUID := uuid.MustParse("12345678-1234-1234-1234-123456789abc")
	ctx := context.WithValue(context.Background(), "user_id", userUUID.String())

	req := &pb.URLExpandRequest{}
	req.SetId("nonexistent")

	resp, err := handler.ExpandURL(ctx, req)

	require.Error(t, err)
	assert.Nil(t, resp)

	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Aborted, st.Code())
}

// Test ExpandURL without user_id in context
func TestExpandURL_NoUserID(t *testing.T) {
	handler := newTestGRPCHandler()
	ctx := context.Background()

	req := &pb.URLExpandRequest{}
	req.SetId("abc123")

	resp, err := handler.ExpandURL(ctx, req)

	require.Error(t, err)
	assert.Nil(t, resp)

	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Aborted, st.Code())
}

// Test ListUserURLs with valid request
func TestListUserURLs_Success(t *testing.T) {
	handler := newTestGRPCHandler()
	userUUID := uuid.MustParse("12345678-1234-1234-1234-123456789abc")
	ctx := context.WithValue(context.Background(), "user_id", userUUID.String())

	resp, err := handler.ListUserURLs(ctx, &emptypb.Empty{})

	require.NoError(t, err)
	require.NotNil(t, resp)

	urls := resp.GetUrl()
	assert.Equal(t, 2, len(urls))

	// Verify URLs are returned
	foundURLs := make(map[string]string)
	for _, urlData := range urls {
		foundURLs[urlData.GetShortUrl()] = urlData.GetOriginalUrl()
	}

	assert.Equal(t, "https://example.com", foundURLs["abc123"])
	assert.Equal(t, "https://test.com", foundURLs["xyz789"])
}

// Test ListUserURLs without user_id in context
func TestListUserURLs_NoUserID(t *testing.T) {
	handler := newTestGRPCHandler()
	ctx := context.Background()

	resp, err := handler.ListUserURLs(ctx, &emptypb.Empty{})

	require.Error(t, err)
	assert.Nil(t, resp)

	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Aborted, st.Code())
}