package routers

import (
	"edge-app/api/handlers"
	"github.com/gin-gonic/gin"
)

func BaseRouter(r *gin.RouterGroup) {
	r.GET("/", handlers.BaseHandler)
}
