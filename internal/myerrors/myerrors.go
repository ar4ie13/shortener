package myerrors

import "errors"

var (
	ErrNotFound           = errors.New("not found")
	ErrURLExist           = errors.New("URL already exist")
	ErrEmptyShortURLorURL = errors.New("shortURL or URL cannot be empty")
	ErrShortURLExist      = errors.New("shortURL already exist")
	ErrInvalidUserUUID    = errors.New("invalid user UUID")
	ErrShortURLIsDeleted  = errors.New("short URL is deleted")

	ErrEmptyURL         = errors.New("URL template cannot be empty")
	ErrWrongHTTPScheme  = errors.New("URL template must use http or https scheme")
	ErrMustIncludeHost  = errors.New("URL template must include a host")
	ErrInvalidURLFormat = errors.New("invalid URL format")

	ErrEmptyID        = errors.New("short url cannot be empty")
	ErrShortURLLength = errors.New("short url length is too small")
)
