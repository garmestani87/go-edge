package logging

type (
	Category    string
	SubCategory string
	ExtraKey    string
)

const (
	General         Category = "General"
	Io              Category = "Io"
	Auth            Category = "Auth"
	Internal        Category = "Internal"
	Database        Category = "Database"
	Cache           Category = "Cache"
	Validation      Category = "Validation"
	RequestResponse Category = "RequestResponse"
	Prometheus      Category = "Prometheus"
	Kafka           Category = "Kafka"
)

const (
	PublicKey           SubCategory = "PUBLIC_KEY"
	Startup             SubCategory = "Startup"
	ExternalService     SubCategory = "ExternalService"
	Migration           SubCategory = "Migration"
	Select              SubCategory = "Select"
	Rollback            SubCategory = "Rollback"
	Update              SubCategory = "Update"
	Delete              SubCategory = "Delete"
	Insert              SubCategory = "Insert"
	Api                 SubCategory = "Api"
	HashPassword        SubCategory = "HashPassword"
	DefaultRoleNotFound SubCategory = "DefaultRoleNotFound"
	FailedToCreateUser  SubCategory = "FailedToCreateUser"
	MobileValidation    SubCategory = "MobileValidation"
	PasswordValidation  SubCategory = "PasswordValidation"
	RemoveFile          SubCategory = "RemoveFile"
	OpenFile            SubCategory = "OpenFile"
	Producer            SubCategory = "Producer"
	Consumer            SubCategory = "Consumer"
)

const (
	AppName      ExtraKey = "AppName"
	LoggerName   ExtraKey = "Logger"
	ClientIp     ExtraKey = "ClientIp"
	HostIp       ExtraKey = "HostIp"
	Method       ExtraKey = "Method"
	StatusCode   ExtraKey = "StatusCode"
	BodySize     ExtraKey = "BodySize"
	Path         ExtraKey = "Path"
	Latency      ExtraKey = "Latency"
	RequestBody  ExtraKey = "RequestBody"
	ResponseBody ExtraKey = "ResponseBody"
	ErrorMessage ExtraKey = "ErrorMessage"
)
