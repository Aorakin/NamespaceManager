package interfaces

import "github.com/gin-gonic/gin"

type UsersHandlers interface {
	LoginWithGoogle() gin.HandlerFunc
	Login() gin.HandlerFunc
	Callback() gin.HandlerFunc
	Register() gin.HandlerFunc
	Logout() gin.HandlerFunc
	Me() gin.HandlerFunc
	GetAccessTokenFromCode() gin.HandlerFunc
	RefreshAccessToken() gin.HandlerFunc
	CheckAuthStatus() gin.HandlerFunc
}
