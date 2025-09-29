package place_bid

import "errors"

var (
	ErrUnauthorized    = errors.New("unauthorized")
	ErrVersionConflict = errors.New("version_conflict")
)
