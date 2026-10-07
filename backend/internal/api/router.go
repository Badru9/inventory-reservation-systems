package api

import (
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/indico/flashsale/internal/service"
)

func NewRouter(svc *service.Service) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery(), requestLogger(), corsMiddleware())

	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "time": time.Now().UTC()})
	})

	v1 := r.Group("/api/v1")
	h := NewHandler(svc)
	inventory := v1.Group("/inventory")
	{
		inventory.POST("/reserve", h.Reserve)
		inventory.POST("/confirm", h.Confirm)
		inventory.GET("/stock", h.Stock)
	}
	return r
}

func requestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		// Keep logs tiny — only the fields that matter for ops.
		gin.DefaultWriter.Write([]byte(
			time.Now().Format(time.RFC3339) + " " +
				c.Request.Method + " " + c.Request.URL.Path + " " +
				http.StatusText(c.Writer.Status()) + " " +
				time.Since(start).String() + "\n",
		))
	}
}

// corsMiddleware lets the browser frontend talk to this API.
//
// Origins are read from CORS_ALLOWED_ORIGINS (comma-separated). When unset
// (e.g. local dev), http://localhost:5173 is allowed by default.
func corsMiddleware() gin.HandlerFunc {
	defaults := []string{"http://localhost:5173"}
	allowed := defaults
	if raw := os.Getenv("CORS_ALLOWED_ORIGINS"); raw != "" {
		allowed = nil
		for _, o := range strings.Split(raw, ",") {
			if o = strings.TrimSpace(o); o != "" {
				allowed = append(allowed, o)
			}
		}
	}
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin != "" {
			ok := false
			for _, a := range allowed {
				if a == "*" || a == origin {
					ok = true
					break
				}
			}
			if ok {
				c.Header("Access-Control-Allow-Origin", origin)
				c.Header("Vary", "Origin")
				c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
				c.Header("Access-Control-Allow-Headers", "Content-Type")
				c.Header("Access-Control-Max-Age", "600")
			}
		}
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
