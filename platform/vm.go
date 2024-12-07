package platform

import (
	"context"
	"errors"
	"github.com/jjmaturino/pulleytakehome/pkg/bootstrapper/api"
	"log"
	"time"

	gzip "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func DefaultGinEngine(logger *zap.Logger) (*gin.Engine, error) {
	r := gin.Default()
	r.ContextWithFallback = true

	if logger == nil {
		logger = zap.L()
	}

	r.Use(
		gzip.Ginzap(logger, time.RFC3339, true),
		gzip.RecoveryWithZap(logger, true),
		corsMiddleware("*"),
	)
	r.NoRoute(func(c *gin.Context) {
		err := api.SendNotFoundResponse(c)
		if err != nil {
			zap.L().Error("failed to SendNotFoundResponse", zap.Error(err))
		}
	})
	return r, nil
}

// StartVM begins a gin server on a virtual machine
func StartVM(service ApiService, engine Engine, deps ...interface{}) error {
	ctx := context.TODO()

	if engine == nil {
		return errors.New("engine is nil")
	}

	err := service.ConstructService(ctx, deps)
	if err != nil {
		log.Printf("Error: %s", err.Error())
		return err
	}

	eng, err := service.SetupEngine(engine)
	if err != nil {
		log.Printf("Error: %s", err.Error())

		return err
	}

	// Start the Gin server on default port 8080
	return eng.Run() // Default listens on :8080
}

// api package private functions
func corsMiddleware(allowedOrigin string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set(api.CORSOriginHeader, allowedOrigin)
		c.Next()
	}
}
