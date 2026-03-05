package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hupham/x-gatekeeper/internal/config"
	"github.com/hupham/x-gatekeeper/internal/word"
	"google.golang.org/grpc"
)

// Server wraps the Gin engine and config. No shared mutable request/session state.
type Server struct {
	engine   *gin.Engine
	cfg      *config.Config
	logger   *slog.Logger
	wordConn *grpc.ClientConn
}

// New builds a Server with the given config and logger. Router and handlers are set up here.
func New(cfg *config.Config, logger *slog.Logger) (*Server, error) {
	wordClient, wordConn, err := word.NewClient(cfg.XWordAddr)
	if err != nil {
		return nil, fmt.Errorf("word client: %w", err)
	}
	wordHandler := word.NewHandler(wordClient)

	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.Use(gin.Recovery())
	engine.Use(corsMiddleware(cfg.CORSAllowedOrigins))
	engine.Use(requestLogger(logger))

	registerRoutes(engine, wordHandler)

	return &Server{engine: engine, cfg: cfg, logger: logger, wordConn: wordConn}, nil
}

// Run starts the HTTP server and blocks until ctx is cancelled or the server fails.
func (s *Server) Run(ctx context.Context) error {
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", s.cfg.HTTPPort),
		Handler: s.engine,
	}
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
		_ = s.wordConn.Close()
	}()
	s.logger.Info("http server listening", "port", s.cfg.HTTPPort)
	return srv.ListenAndServe()
}

// registerRoutes wires all API paths to their handlers.
// The nginx ingress rewrites /api/* → /* so routes here use the base paths.
func registerRoutes(r *gin.Engine, wh *word.Handler) {
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"service": "x-gatekeeper", "message": "hello from x-gatekeeper", "time": time.Now().Format(time.RFC3339)})
	})
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Word API — proxied from x-word gRPC
	r.GET("/v1/word/:word", wh.GetWord)
	r.POST("/v1/words", wh.GetWords)
}

// corsMiddleware sets CORS headers on every response and short-circuits OPTIONS preflight requests.
func corsMiddleware(allowedOrigins []string) gin.HandlerFunc {
	originSet := make(map[string]struct{}, len(allowedOrigins))
	allowAll := false
	for _, o := range allowedOrigins {
		if o == "*" {
			allowAll = true
			break
		}
		originSet[o] = struct{}{}
	}

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		allowed := ""
		if allowAll {
			allowed = "*"
		} else if _, ok := originSet[origin]; ok {
			allowed = origin
		}

		if allowed != "" {
			c.Header("Access-Control-Allow-Origin", allowed)
			c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
			c.Header("Access-Control-Max-Age", "86400")
		}

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

// requestLogger returns a Gin middleware that logs request method, path, and status (structured).
func requestLogger(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		method := c.Request.Method
		c.Next()
		status := c.Writer.Status()
		logger.Info("request", "method", method, "path", path, "status", status)
	}
}
