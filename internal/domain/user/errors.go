package user

import "errors"

// ErrUserIDNotFound is returned when user ID is not found in cookies.
var ErrUserIDNotFound = errors.New("user_id not found")
