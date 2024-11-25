package authorization

import (
	"edge-app/configs"
	"edge-app/pkg/constant"
	"edge-app/pkg/errors"
	"edge-app/pkg/logging"
	"strings"

	"github.com/gin-gonic/gin"
)

type Service struct {
	logger logging.Logger
	cfg    *configs.Config
	Template
}

func NewAuthorizationService(cfg *configs.Config) *Service {
	logger := logging.NewLogger(cfg)
	return &Service{
		logger: logger,
		cfg:    cfg,
	}
}

func (s *Service) retrieveScopes(ctx *gin.Context) (scopeMap map[string]int, err error) {
	scope, exists := ctx.Get(constant.Scope)
	if !exists {
		return nil, &errors.ServiceError{ErrorDescription: errors.ErrScopeNotFound}
	}

	scopes := strings.Split(scope.(string), " ")
	scopeMap = map[string]int{}
	for _, item := range scopes {
		scopeMap[item] = 0
	}
	return scopeMap, nil
}

func (s *Service) hasScope(ctx *gin.Context, scopeMap map[string]int) (ok bool, err error) {
	aud, exists := ctx.Get(constant.Aud)
	if !exists {
		return false, &errors.ServiceError{ErrorDescription: errors.ErrAudNotFound}
	}
	validScope, exists := s.cfg.ValidScopes[aud.(string)]
	if !exists {
		return false, &errors.ServiceError{ErrorDescription: errors.ErrValidScopeNotDefined}
	}

	validScopes := strings.Split(validScope, ",")
	for _, item := range validScopes {
		if _, ok := scopeMap[item]; ok {
			return true, nil
		}
	}
	return false, &errors.ServiceError{ErrorDescription: errors.ErrAccessForbidden}
}
