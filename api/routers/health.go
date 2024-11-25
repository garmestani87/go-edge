package routers

import (
	"edge-app/api/handlers"
	"github.com/gin-gonic/gin"
)

func Health(r *gin.RouterGroup) {
	r.GET("/health", handlers.Health)
}
