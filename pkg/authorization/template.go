package authorization

import "github.com/gin-gonic/gin"

type Template interface {
	retrieveScopes(ctx *gin.Context) (scopeMap map[string]int, err error)
	hasScope(ctx *gin.Context, scopeMap map[string]int) (ok bool, err error)
}

type Tpl struct {
	Impl Template
}

func (t *Tpl) HasRole(ctx *gin.Context) (bool, error) {
	var (
		err      error
		scopeMap map[string]int
	)

	if scopeMap, err = t.Impl.retrieveScopes(ctx); err != nil {
		return false, err
	}
	if _, err = t.Impl.hasScope(ctx, scopeMap); err != nil {
		return false, err
	}

	return true, nil
}
