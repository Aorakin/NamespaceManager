package http

import (
	"log"
	"net/http"

	"github.com/NamespaceManager/internal/users/dtos"
	"github.com/NamespaceManager/internal/users/interfaces"
	"github.com/NamespaceManager/internal/utils"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type UsersHandlers struct {
	usersUsecase interfaces.UsersUsecase
}

func NewUsersHandler(usersUsecase interfaces.UsersUsecase) interfaces.UsersHandlers {
	return &UsersHandlers{
		usersUsecase: usersUsecase,
	}
}

func (h *UsersHandlers) LoginWithGoogle() gin.HandlerFunc {
	return func(c *gin.Context) {
		url := h.usersUsecase.GenerateLoginURL("state-token")
		c.Redirect(http.StatusTemporaryRedirect, url)
	}
}

func (h *UsersHandlers) Callback() gin.HandlerFunc {
	return func(c *gin.Context) {
		code := c.Query("code")
		if code == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No code provided"})
			return
		}

		userInfo, err := h.usersUsecase.HandleGoogleCallback(code, c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to handle callback"})
			return
		}

		c.JSON(http.StatusOK, userInfo)
	}
}

// Register godoc
// @Summary User registration
// @Description Register a new user
// @Tags users
// @Accept json
// @Produce json
// @Param user body dtos.RegisterInput true "User registration data"
// @Success 201 {object} map[string]string "User registered successfully"
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 409 {object} map[string]string "User already exists"
// @Router /users/register [get]
func (h *UsersHandlers) Register() gin.HandlerFunc {
	return func(c *gin.Context) {
		var registerInput dtos.RegisterInput
		if err := c.ShouldBindJSON(&registerInput); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
			return
		}
		if err := utils.CheckValidater(registerInput); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Bad request"})
			return
		}

		if err := h.usersUsecase.Register(registerInput); err != nil {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"message": "registered successfully"})
	}
}

// Login godoc
// @Summary User login
// @Description Login with username and password
// @Tags users
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param username formData string true "Username"
// @Param password formData string true "Password"
// @Success 200 {object} map[string]string "Login successful"
// @Failure 401 {object} map[string]string "Invalid credentials"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /users/login [post]
func (h *UsersHandlers) Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		username := c.PostForm("username")
		password := c.PostForm("password")
		user, err := h.usersUsecase.Login(username, password)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid credentials"})
			return
		}

		session := sessions.Default(c)
		session.Set("userID", user.ID)
		err = session.Save()
		if err != nil {
			log.Printf("Session save failed: %+v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"Session save failed": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Login successful"})

	}
}

// Logout godoc
// @Summary User logout
// @Description Logout user and clear session
// @Tags users
// @Produce json
// @Success 200 {object} map[string]string "Logout successful"
// @Security ApiKeyAuth
// @Router /users/logout [post]
func (h *UsersHandlers) Logout() gin.HandlerFunc {
	return func(c *gin.Context) {
		utils.ClearSession(c)
		c.JSON(http.StatusOK, gin.H{"message": "Logout successful"})
	}
}
