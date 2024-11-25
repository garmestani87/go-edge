package errors

const (
	ErrGeneral                       = 0
	ErrRequiredFieldMissing          = 1
	ErrInvalidFormatOrCheckDigit     = 2
	ErrDuplicateData                 = 3
	ErrDataMismatch                  = 4
	ErrDataContractMismatch          = 5
	ErrDataNotFound                  = 6
	ErrInsufficientResources         = 7
	ErrAccessDenied                  = 8
	ErrTransactionUnavailable        = 9
	ErrServiceUnavailable            = 10
	ErrInternalServerError           = 11
	ErrExternalServiceUnavailable    = 12
	ErrNoResponseFromExternalService = 13
	ErrResendRequest                 = 14
	ErrServiceNotAvailable           = 15
	ErrSecurityError                 = 16
	ErrDataOutOfRange                = 17
	ErrInactiveReference             = 18
	ErrReferenceExpiredOrInvalid     = 19
)

const (
	ErrMissingJwtToken      = "missing jwt token !"
	ErrUnexpectedError      = "unexpected error occurred !"
	ErrClaimNotFound        = "claim not found !"
	ErrTokenExpired         = "token expired !"
	ErrTokenInvalid         = "token is invalid !"
	ErrClientIdNotFound     = "client id not found !"
	ErrPublicKeyNotFound    = "public key not found !"
	ErrPublicKeyIsInvalid   = "public key is invalid !"
	ErrSignatureIsInvalid   = "signature is invalid !"
	ErrIssuerIsInvalid      = "issuer is invalid !"
	ErrScopeNotFound        = "scope not found !"
	ErrAudNotFound          = "aud not found !"
	ErrValidScopeNotDefined = "valid scope note defined in config file for this client !"
	ErrAccessForbidden      = "you can not consume this service !"
)
