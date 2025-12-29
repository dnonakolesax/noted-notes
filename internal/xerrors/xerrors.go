package xerrors

import "errors"

var ErrInvalidFileName = errors.New("bad file name")

var ErrCSRFTokenNotFound = errors.New("CSRF token not found")