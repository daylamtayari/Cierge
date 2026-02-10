package handler

import "github.com/daylamtayari/cierge/server/internal/service"

type User struct {
	userService *service.User
}

func NewUser(userService *service.User) *User {
	return &User{
		userService: userService,
	}
}
