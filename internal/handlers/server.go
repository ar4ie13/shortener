package handlers

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	pb "github.com/ar4ie13/shortener/api/proto"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

// StartServer starts http and gRPC (if enabled) server
func (h Handler) StartServer() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-sigChan
		cancel()
	}()

	g, ctx := errgroup.WithContext(ctx)

	// HTTP server
	srv := &http.Server{
		Addr:    h.cfg.GetLocalServerAddr(),
		Handler: h.RegisterRoutes(),
	}

	g.Go(func() error {
		h.zlog.Info().Msgf("listening on %v\nURL Template: %v\nLog Level: %v", h.cfg.GetLocalServerAddr(),
			h.cfg.GetShortURLTemplate(), h.cfg.GetLogLevel())
		switch h.cfg.GetHTTPS() {
		case true:
			if err := srv.ListenAndServeTLS(h.cfg.GetTLSCertPath(), h.cfg.GetTLSKeyPath()); !errors.Is(err, http.ErrServerClosed) {
				return err
			}
		default:
			if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
				return err
			}

		}
		return nil
	})

	// HTTP graceful shutdown
	g.Go(func() error {
		<-ctx.Done()
		h.zlog.Info().Msg("Shutting down HTTP server...")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		return srv.Shutdown(shutdownCtx)
	})

	// gRPC server
	if h.cfg.GetGRPCEnabled() {
		grpcServer := grpc.NewServer(grpc.UnaryInterceptor(h.authorizationInterceptor))
		pb.RegisterShortenerServiceServer(grpcServer, &h)

		g.Go(func() error {
			listen, err := net.Listen("tcp", h.cfg.GetGRPCServerAddr())
			if err != nil {
				h.zlog.Error().Err(err).Msg("gRPC listener initialization error")
				return err
			}
			h.zlog.Info().Msgf("gRPC server listening on %s", h.cfg.GetGRPCServerAddr())

			if err = grpcServer.Serve(listen); err != nil {
				return fmt.Errorf("gRPC server error: %w", err)
			}
			return nil
		})

		// Graceful shutdown for gRPC
		g.Go(func() error {
			<-ctx.Done()
			h.zlog.Info().Msg("Shutting down gRPC server...")
			grpcServer.GracefulStop()
			return nil
		})

	}

	// Waiting for all goroutines to stop
	if err := g.Wait(); err != nil {
		h.zlog.Error().Msgf("Server shutdown with error: %v", err)
	}

	h.zlog.Info().Msgf("shortener server shutdown gracefully")
	return nil
}
