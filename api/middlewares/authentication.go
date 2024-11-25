package middlewares

import (
	"edge-app/api/helpers"
	"edge-app/configs"
	"edge-app/pkg/authentication"
	"edge-app/pkg/constant"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Authentication(cfg *configs.Config) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		template := authentication.Tpl{
			Impl: authentication.NewAuthenticationService(cfg),
		}
		claimMap, err := template.VerifyTokenTP(ctx.GetHeader(constant.AuthorizationKey))
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, helpers.CreateBaseResponseWithError(
				nil, false, helpers.AuthError, err,
			))
			return
		}
		ctx.Set(constant.Scope, claimMap[constant.Scope])
		ctx.Set(constant.Aud, claimMap[constant.Aud])
		ctx.Next()
	}
}
