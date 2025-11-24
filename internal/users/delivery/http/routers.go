package http

import (
	"github.com/NamespaceManager/internal/middleware"
	"github.com/NamespaceManager/internal/users/interfaces"
	"github.com/NamespaceManager/internal/utils"
	"github.com/gin-gonic/gin"
)

func MapUsersRoutes(usersGroup *gin.RouterGroup, usersHandler interfaces.UsersHandlers) {
	usersGroup.GET("/access-token", usersHandler.GetAccessTokenFromCode())
	usersGroup.POST("/auth/refresh-token", usersHandler.RefreshAccessToken())
	usersGroup.GET("/auth/status", usersHandler.CheckAuthStatus())
	usersGroup.GET("/auth/google", usersHandler.LoginWithGoogle())
	usersGroup.GET("/auth/callback/google", usersHandler.Callback())
	usersGroup.POST("/register", usersHandler.Register())
	usersGroup.POST("/login", usersHandler.Login())
	usersGroup.POST("/podStatus", utils.TicketStatus())
	usersGroup.GET("/logout", usersHandler.Logout()).Use(middleware.AuthMiddleware())
	usersGroup.GET("/me", usersHandler.Me()).Use(middleware.AuthMiddleware())

}
