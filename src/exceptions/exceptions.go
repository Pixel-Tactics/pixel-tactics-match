package exceptions

import "errors"

func InvalidDataError() error {
	return errors.New("data is invalid")
}

func InvalidJsonError() error {
	return errors.New("json is invalid")
}
