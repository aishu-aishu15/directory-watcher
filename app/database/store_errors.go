package database

type NoSuchRecordError struct {
	Message string
	Cause   error
}

func (e NoSuchRecordError) Error() string {
	if e.Cause != nil {
		return e.Message + ": " + e.Cause.Error()
	}

	return e.Message
}

type RepoError struct {
	Message string
	Cause   error
}

func (e RepoError) Error() string {
	if e.Cause != nil {
		return e.Message + ": " + e.Cause.Error()
	}

	return e.Message
}
