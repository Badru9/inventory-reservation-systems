package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/indico/flashsale/internal/service"
)

func NewRouter(svc *service.Service) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery(), requestLogger())

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
