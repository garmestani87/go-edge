package middlewares

import (
	"edge-app/api/helpers"
	"edge-app/configs"
	"edge-app/pkg/authorization"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Authorization(cfg *configs.Config) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authorizationService := authorization.NewAuthorizationService(cfg)
		template := authorization.Tpl{
			Impl: authorizationService,
		}
		if _, err := template.HasRole(ctx); err != nil {
			ctx.AbortWithStatusJSON(http.StatusForbidden, helpers.CreateBaseResponseWithError(nil, false, helpers.ForbiddenError, err))
			return
		}
		ctx.Next()
	}
}
