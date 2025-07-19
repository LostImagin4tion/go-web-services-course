package utils

type ApiError struct {
	ResponseCode int
	Err          string
}

func (e ApiError) Error() string {
	return e.Err
}
