package helpers

type ResultCode int

const (
	Success         ResultCode = 0
	ValidationError ResultCode = 400
	AuthError       ResultCode = 401
	ForbiddenError  ResultCode = 403
	NotFoundError   ResultCode = 404
	CustomRecovery  ResultCode = 500
	InternalError   ResultCode = 500
)
