package http

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/NamespaceManager/internal/auth"
	"github.com/NamespaceManager/internal/models"
	"github.com/NamespaceManager/internal/users/dtos"
	"github.com/NamespaceManager/internal/users/interfaces"
	"github.com/NamespaceManager/internal/utils"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UsersHandlers struct {
	usersUsecase interfaces.UsersUsecase
}

func NewUsersHandler(usersUsecase interfaces.UsersUsecase) interfaces.UsersHandlers {
	return &UsersHandlers{
		usersUsecase: usersUsecase,
	}
}

func (h *UsersHandlers) Me() gin.HandlerFunc {
	return func(c *gin.Context) {
		userData := map[string]interface{}{}
		userID := c.MustGet("userID").(uuid.UUID)
		firstName := c.MustGet("firstname").(string)
		lastName := c.MustGet("lastname").(string)
		email := c.MustGet("email").(string)
		log.Printf("Me Handler - UserID: %s, FirstName: %s, LastName: %s, Email: %s", userID, firstName, lastName, email)

		userData["id"] = userID
		userData["username"] = firstName + " " + lastName
		userData["email"] = email
		if userID == uuid.Nil {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
			return
		}
		c.JSON(http.StatusOK, userData)
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

		// Extract user from userInfo
		if user, exists := userInfo["user"]; exists {
			if userModel, ok := user.(*models.User); ok {
				// Set session for the logged-in user
				session := sessions.Default(c)
				session.Set("userID", userModel.ID)
				err = session.Save()
				if err != nil {
					log.Printf("Session save failed: %+v", err)
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Session save failed"})
					return
				}

				// c.JSON(http.StatusOK, gin.H{
				// 	"message": "Login successful",
				// 	"user": gin.H{
				// 		"id":       userModel.ID,
				// 		"username": userModel.Username,
				// 		"email":    userModel.Email,
				// 		"role":     userModel.Role,
				// 	},
				// })
				c.Redirect(http.StatusFound, "http://localhost:3000/projects")
				return
			}
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
// @Router /users/register [post]
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
			c.JSON(http.StatusUnauthorized, gin.H{"message": "invalid username or password"})
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
		// Get refresh token from cookie
		refreshToken, err := c.Cookie("refresh_token")
		if err == nil && refreshToken != "" {
			// Send logout request to clearing house to blacklist the token
			url := os.Getenv("CLEARINGHOUSE_URL") + "/auth/logout"
			status, body, err := utils.SendRequestWithAccessToken(url, nil, "POST", refreshToken)
			if err != nil {
				// Log error but don't fail logout
				log.Printf("Failed to blacklist token at clearing house: %v", err)
			} else if status != http.StatusOK {
				log.Printf("Clearing house logout failed with status %d: %s", status, string(body))
			}
		}

		// Clear the access token cookie
		c.SetCookie("access_token", "", -1, "/", ".onepointfive.life", true, true)
		// Clear the refresh token cookie
		c.SetCookie("refresh_token", "", -1, "/users/auth", ".onepointfive.life", true, true)

		c.JSON(http.StatusOK, gin.H{"message": "Successfully logged out"})
	}
}

func (h *UsersHandlers) GetAccessTokenFromCode() gin.HandlerFunc {
	return func(c *gin.Context) {
		rawQuery := c.Request.URL.RawQuery
		url := os.Getenv("CLEARINGHOUSE_URL") + "/auth/callback/google?" + rawQuery
		status, body, err := utils.SendRequest(url, nil, "GET")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if status != http.StatusOK {
			c.JSON(status, gin.H{"error": string(body)})
			return
		}
		var token map[string]interface{}
		if err := json.Unmarshal(body, &token); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse token"})
			return
		}

		accessToken, ok := token["access_token"].(string)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Access token not found"})
			return
		}
		refreshToken, ok := token["refresh_token"].(string)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Refresh token not found"})
			return
		}

		userID, firstName, lastName, email, err := auth.ExtractDataFromToken(accessToken)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to extract user data from token"})
			return
		}
		log.Printf("UserID: %s, FirstName: %s, LastName: %s, Email: %s", userID, firstName, lastName, email)

		_, err = h.usersUsecase.FindOrCreateUser(userID, email, firstName, lastName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find or create user"})
			return
		}

		c.SetCookie("access_token", accessToken, 7*24*3600, "/", ".onepointfive.life", true, true)
		c.SetCookie("refresh_token", refreshToken, 7*24*3600, "/users/auth", ".onepointfive.life", true, true)

		c.JSON(http.StatusOK, gin.H{"message": "Tokens set successfully"})
	}
}

func (h *UsersHandlers) RefreshAccessToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		refreshTokenCookie, err := c.Cookie("refresh_token")
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Refresh token not found"})
			return
		}
		url := os.Getenv("CLEARINGHOUSE_URL") + "/auth/refresh-token"
		status, body, err := utils.SendRequestWithAccessToken(url, nil, "GET", refreshTokenCookie)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if status != http.StatusOK {
			c.JSON(status, gin.H{"error": string(body)})
			return
		}
		var token map[string]interface{}
		if err := json.Unmarshal(body, &token); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse token"})
			return
		}

		accessToken, ok := token["access_token"].(string)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Access token not found"})
			return
		}
		// c.SetCookie("access_token", accessToken, 3600, "/", "localhost", true, true)
		c.SetCookie("access_token", accessToken, 7*24*3600, "/", ".onepointfive.life", true, true)

		c.JSON(http.StatusOK, gin.H{"message": "Tokens refreshed successfully"})
	}
}
