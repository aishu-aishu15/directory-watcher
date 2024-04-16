package core

type NotFoundError struct {
	Message string
	Cause   error
}

func (e NotFoundError) Error() string {
	if e.Cause != nil {
		return e.Message + ": " + e.Cause.Error()
	}

	return e.Message
}

type InvalidRequestError struct {
	Message string
	Cause   error
}

func (e InvalidRequestError) Error() string {
	if e.Cause != nil {
		return e.Message + ": " + e.Cause.Error()
	}

	return e.Message
}

type InternalServerError struct {
	Message string
	Cause   error
}

func (e InternalServerError) Error() string {
	if e.Cause != nil {
		return e.Message + ": " + e.Cause.Error()
	}

	return e.Message
}
