package models

type RequestError struct {
	error
	msg string
}

func NewRequestError(msg string) RequestError {
	return RequestError{
		msg: msg,
	}
}

func (err RequestError) Error() string {
	return err.msg
}

func IsRequestError(err error) bool {
	if _, ok := err.(RequestError); ok {
		return true
	} else {
		return false
	}
}
