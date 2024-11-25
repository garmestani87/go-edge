package authentication

type Template interface {
	extractToken(jwtToken string) (token string, err error)
	getClaims(token string) (claimMap map[string]interface{}, err error)
	isIssuerValid(claims map[string]interface{}) (bool, error)
	isExpirationTimeValid(claims map[string]interface{}) (bool, error)
	isSignatureValid(claims map[string]interface{}, token string) (bool, error)
}

type Tpl struct {
	Impl Template
}

func (t *Tpl) VerifyTokenTP(jwtToken string) (map[string]interface{}, error) {
	var (
		err      error
		token    string
		claimMap map[string]interface{}
	)
	if token, err = t.Impl.extractToken(jwtToken); err != nil {
		return nil, err
	}
	if claimMap, err = t.Impl.getClaims(token); err != nil {
		return nil, err
	}
	if _, err = t.Impl.isIssuerValid(claimMap); err != nil {
		return nil, err
	}
	if _, err = t.Impl.isExpirationTimeValid(claimMap); err != nil {
		return nil, err
	}
	if _, err = t.Impl.isSignatureValid(claimMap, token); err != nil {
		return nil, err
	}

	return claimMap, nil
}
