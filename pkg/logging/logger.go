package logging

import (
	"log/slog"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

var L *slog.Logger

func Init(serviceName string) {
	opts := &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}

	handler := slog.NewJSONHandler(os.Stdout, opts).WithAttrs([]slog.Attr{
		slog.String("service", serviceName),
	})

	L = slog.New(handler)
	slog.SetDefault(L)
}

// GinMiddleware returns a gin.HandlerFunc that logs requests
func GinMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		end := time.Now()
		latency := end.Sub(start)

		L.InfoContext(c.Request.Context(), "request handled",
			slog.Int("status", c.Writer.Status()),
			slog.String("method", c.Request.Method),
			slog.String("path", path),
			slog.String("query", query),
			slog.String("ip", c.ClientIP()),
			slog.Duration("latency", latency),
			slog.String("user_agent", c.Request.UserAgent()),
		)
	}
}

// GetLogger returns the global logger
func GetLogger() *slog.Logger {
	if L == nil {
		Init("default")
	}
	return L
}
