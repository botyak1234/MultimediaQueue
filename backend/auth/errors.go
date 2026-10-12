package auth

import "errors"

var (
	ErrUserExists       = errors.New("пользователь с таким именем уже существует")
	ErrInvalidCreds     = errors.New("неверное имя пользователя или пароль")
	ErrUsernameRequired = errors.New("имя пользователя не может быть пустым")
	ErrPasswordLength   = errors.New("пароль должен быть не менее 4 символов")
	ErrUnauthorized     = errors.New("необходима авторизация")
)
