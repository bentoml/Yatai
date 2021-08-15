package scookie

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

const (
	UserNameKey = "username"
)

func SetUsernameToCookie(ctx *gin.Context, username string) error {
	session := sessions.Default(ctx)
	session.Set(UserNameKey, username)
	return session.Save()
}

func GetUsernameFromCookie(ctx *gin.Context) string {
	session := sessions.Default(ctx)
	username, ok := session.Get(UserNameKey).(string)
	if !ok {
		return ""
	}
	return username
}
