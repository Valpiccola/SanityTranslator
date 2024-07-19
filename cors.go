package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetCORSConfig() gin.HandlerFunc {
	env := os.Getenv("ENV")
	switch env {
	case "production":
		fmt.Println("CORS: Production")
		origins := os.Getenv("ALLOWED_ORIGINS")
		originsSlice := strings.Split(origins, ",")
		return cors.New(cors.Config{
			AllowOrigins: originsSlice,
			AllowMethods: []string{"POST", "OPTIONS", "GET"},
			AllowHeaders: []string{
				"Content-Type",
				"Content-Length",
				"Accept-Encoding",
				"X-CSRF-Token",
				"Authorization",
				"accept",
				"origin",
				"Cache-Control",
				"X-Requested-With",
				"sentry-trace",
				"baggage",
			},
			ExposeHeaders:    []string{"Content-Length"},
			AllowCredentials: true,
			MaxAge:           12 * time.Hour,
		})
	case "staging":
		fmt.Println("CORS: Staging")
		return cors.Default()
	default:
		fmt.Println("CORS: Development")
		return nil
	}
}
