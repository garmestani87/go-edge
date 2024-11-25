package errors

type ServiceErrors struct {
	Errors []ServiceError `json:"errors"`
}

type ServiceError struct {
	ErrorCode        int    `json:"errorCode"`
	ErrorDescription string `json:"errorDescription"`
	ReferenceName    string `json:"referenceName"`
	OriginalValue    string `json:"originalValue"`
	ExtraData        string `json:"extraData"`
}

func (s *ServiceError) Error() string {
	return s.ErrorDescription
}
