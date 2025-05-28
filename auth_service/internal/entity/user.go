package entity

import "errors"

type User struct {
	ID       string
	Username string
	Email    string
	Password string
	Role     string
}

type TokenDetails struct {
	AccessToken  string
	RefreshToken string
	AccessUuid   string
	RefreshUuid  string
	AtExpires    int64
	RtExpires    int64
}

var (
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrInvalidEmail      = errors.New("invalid email")
	ErrWeakPassword      = errors.New("weak password")
	ErrEmptyUsername     = errors.New("empty username")
)
