package constant

const (
	AuthorizationKey string = "Authorization"
	Scope            string = "scope"
	Aud              string = "aud"
	Exp              string = "exp"
	Iss              string = "iss"
	Issuer           string = "http://keycloak-ip:8080"
	Metrics          string = "/metrics"
	BeginPublicKey   string = "-----BEGIN PUBLIC KEY-----\n"
	EndPublicKey     string = "\n-----END PUBLIC KEY-----"
)

type PaymentStatus int

const (
	NotPaid PaymentStatus = 0
	Paid    PaymentStatus = 1
)

type RequestStatus int

const (
	Acceptance    RequestStatus = 1
	NonAcceptance RequestStatus = 2
)
