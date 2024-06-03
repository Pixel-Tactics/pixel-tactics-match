package exceptions

import "errors"

func InvalidDataError() error {
	return errors.New("data is invalid")
}

func InvalidJsonError() error {
	return errors.New("json is invalid")
}

func SessionNotFound() error {
	return errors.New("session not found")
}

func ActionNotAllowed() error {
	return errors.New("action not allowed")
}

func ExceededDeadlineError() error {
	return errors.New("exceeded deadline")
}

func HeroPickupError() error {
	return errors.New("player didn't pickup hero")
}

func HeroIsDead() error {
	return errors.New("hero is dead")
}
