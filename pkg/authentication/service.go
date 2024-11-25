package authentication

import (
	"edge-app/configs"
	"edge-app/pkg/constant"
	"edge-app/pkg/errors"
	"edge-app/pkg/logging"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
)

type Service struct {
	logger logging.Logger
	cfg    *configs.Config
	Tpl
}

func NewAuthenticationService(cfg *configs.Config) *Service {
	logger := logging.NewLogger(cfg)
	return &Service{
		cfg:    cfg,
		logger: logger,
	}
}

func (s *Service) extractToken(jwtToken string) (token string, err error) {
	tokens := strings.Split(jwtToken, " ")
	if jwtToken == "" || len(tokens) < 2 {
		return "", &errors.ServiceError{ErrorDescription: errors.ErrMissingJwtToken}
	}
	return tokens[1], nil
}

func (s *Service) getClaims(token string) (claimMap map[string]interface{}, err error) {
	claimMap = map[string]interface{}{}
	accessToken, _, err := new(jwt.Parser).ParseUnverified(token, jwt.MapClaims{})
	if err != nil {
		return nil, err
	}
	if claims, ok := accessToken.Claims.(jwt.MapClaims); ok {
		for k, v := range claims {
			claimMap[k] = v
		}
		return claimMap, nil
	}
	return nil, &errors.ServiceError{ErrorDescription: errors.ErrClaimNotFound}
}

func (s *Service) isSignatureValid(claimMap map[string]interface{}, token string) (bool, error) {
	clientId, exists := claimMap[constant.Aud]
	if !exists {
		return false, &errors.ServiceError{ErrorDescription: errors.ErrClientIdNotFound}
	}

	publicKey, exist := s.cfg.PublicKeys[clientId.(string)]
	if !exist {
		return false, &errors.ServiceError{ErrorDescription: errors.ErrPublicKeyNotFound}
	}
	pem := constant.BeginPublicKey + publicKey + constant.EndPublicKey
	key, err := jwt.ParseRSAPublicKeyFromPEM([]byte(pem))
	if err != nil {
		return false, &errors.ServiceError{ErrorDescription: errors.ErrPublicKeyIsInvalid}
	}

	parts := strings.Split(token, ".")
	err = jwt.SigningMethodRS256.Verify(strings.Join(parts[0:2], "."), parts[2], key)
	if err != nil {
		return false, &errors.ServiceError{ErrorDescription: errors.ErrSignatureIsInvalid}
	}

	return true, nil
}

func (s *Service) isExpirationTimeValid(claims map[string]interface{}) (bool, error) {
	if expireAt, exists := claims[constant.Exp]; exists {
		if expireAt.(float64) != 0 && expireAt.(float64) < (float64(time.Now().Unix())) {
			return false, &errors.ServiceError{ErrorDescription: errors.ErrTokenExpired}
		}
	}
	return true, nil
}

func (s *Service) isIssuerValid(claims map[string]interface{}) (bool, error) {
	if iss, exists := claims[constant.Iss]; exists {
		if strings.Contains(iss.(string), constant.Issuer) {
			return true, nil
		}
	}
	return false, &errors.ServiceError{ErrorDescription: errors.ErrIssuerIsInvalid}
}
