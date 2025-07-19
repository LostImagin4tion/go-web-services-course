package explorer

type apiError struct {
	ResponseCode int
	Err          error
}

func (e apiError) Error() string {
	return e.Err.Error()
}
