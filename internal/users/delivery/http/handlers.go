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
	apiError "github.com/NamespaceManager/pkg/api_error"
	"github.com/NamespaceManager/pkg/response"
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

// Me godoc
// @Summary      Get current user profile
// @Description  Retrieve authenticated user profile from access token
// @Tags         users
// @Produce      json
// @Success      200  {object}  map[string]interface{}  "Current user profile"
// @Failure      401  {object}  response.ErrorResponse  "Unauthorized"
// @Security     ApiKeyAuth
// @Router       /users/me [get]
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
			c.JSON(response.ErrorResponseBuilder(apiError.NewUnauthorizedError("Please log in to continue")))
			return
		}
		c.JSON(http.StatusOK, userData)
	}
}

// LoginWithGoogle godoc
// @Summary      Start Google OAuth login
// @Description  Redirect user to Google OAuth authorization page
// @Tags         users
// @Produce      json
// @Success      307  {string}  string  "Temporary redirect"
// @Router       /users/auth/google [get]
func (h *UsersHandlers) LoginWithGoogle() gin.HandlerFunc {
	return func(c *gin.Context) {
		url := h.usersUsecase.GenerateLoginURL("state-token")
		c.Redirect(http.StatusTemporaryRedirect, url)
	}
}

// Callback godoc
// @Summary      Google OAuth callback
// @Description  Handle Google OAuth callback with authorization code
// @Tags         users
// @Produce      json
// @Param        code  query     string  true  "Authorization code"
// @Success      200   {object}  map[string]interface{}  "Authentication response"
// @Success      302   {string}  string                  "Redirect after successful login"
// @Failure      400   {object}  response.ErrorResponse  "Bad request"
// @Failure      500   {object}  response.ErrorResponse  "Internal server error"
// @Router       /users/auth/callback/google [get]
func (h *UsersHandlers) Callback() gin.HandlerFunc {
	return func(c *gin.Context) {
		code := c.Query("code")
		if code == "" {
			c.JSON(response.ErrorResponseBuilder(apiError.NewBadRequestError("Authorization code is required")))
			return
		}

		userInfo, err := h.usersUsecase.HandleGoogleCallback(code, c)
		if err != nil {
			log.Printf("failed to handle Google callback: %v", err)
			c.JSON(response.ErrorResponseBuilder(apiError.NewInternalServerError("Failed to complete sign-in")))
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
					c.JSON(response.ErrorResponseBuilder(apiError.NewInternalServerError("Failed to save session")))
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
// @Success 201 {object} map[string]string      "User registered successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid input"
// @Failure 409 {object} response.ErrorResponse "User already exists"
// @Router /users/register [post]
func (h *UsersHandlers) Register() gin.HandlerFunc {
	return func(c *gin.Context) {
		var registerInput dtos.RegisterInput
		if err := c.ShouldBindJSON(&registerInput); err != nil {
			c.JSON(response.ErrorResponseBuilder(apiError.NewBadRequestError("Invalid input")))
			return
		}
		if err := utils.CheckValidater(registerInput); err != nil {
			c.JSON(response.ErrorResponseBuilder(apiError.NewBadRequestError("Invalid input format")))
			return
		}

		if err := h.usersUsecase.Register(registerInput); err != nil {
			c.JSON(response.ErrorResponseBuilder(apiError.NewConflictError("Username is already taken")))
			return
		}
		c.JSON(http.StatusCreated, gin.H{"message": "Registered successfully"})
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
// @Success 200 {object} map[string]string      "Login successful"
// @Failure 401 {object} response.ErrorResponse "Invalid credentials"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /users/login [post]
func (h *UsersHandlers) Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		username := c.PostForm("username")
		password := c.PostForm("password")
		user, err := h.usersUsecase.Login(username, password)
		if err != nil {
			c.JSON(response.ErrorResponseBuilder(apiError.NewUnauthorizedError("Invalid username or password")))
			return
		}

		session := sessions.Default(c)
		session.Set("userID", user.ID)
		err = session.Save()
		if err != nil {
			log.Printf("Session save failed: %+v", err)
			c.JSON(response.ErrorResponseBuilder(apiError.NewInternalServerError("Failed to save session")))
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
// @Success 200 {object} map[string]string      "Logout successful"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Security ApiKeyAuth
// @Router /users/auth/logout [get]
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

// GetAccessTokenFromCode godoc
// @Summary      Exchange Google callback code for tokens
// @Description  Exchange OAuth code via clearing house and set access/refresh token cookies
// @Tags         users
// @Produce      json
// @Param        code   query     string  true  "Authorization code"
// @Param        state  query     string  false "OAuth state"
// @Success      200    {object}  map[string]string      "Tokens set successfully"
// @Failure      500    {object}  response.ErrorResponse "Internal server error"
// @Router       /users/access-token [get]
func (h *UsersHandlers) GetAccessTokenFromCode() gin.HandlerFunc {
	return func(c *gin.Context) {
		rawQuery := c.Request.URL.RawQuery
		url := os.Getenv("CLEARINGHOUSE_URL") + "/auth/callback/google?" + rawQuery
		status, body, err := utils.SendRequest(url, nil, "GET")
		if err != nil {
			log.Printf("failed to get access token from code: %v", err)
			c.JSON(response.ErrorResponseBuilder(apiError.NewInternalServerError("Failed to complete authentication")))
			return
		}
		if status != http.StatusOK {
			log.Printf("auth service returned status %d: %s", status, string(body))
			c.JSON(response.ErrorResponseBuilder(apiError.NewApiError(status, "Authentication failed", "Failed to authenticate, please try again")))
			return
		}
		var token map[string]interface{}
		if err := json.Unmarshal(body, &token); err != nil {
			log.Printf("failed to parse token response: %v", err)
			c.JSON(response.ErrorResponseBuilder(apiError.NewInternalServerError("Failed to process authentication response")))
			return
		}

		accessToken, ok := token["access_token"].(string)
		if !ok {
			c.JSON(response.ErrorResponseBuilder(apiError.NewInternalServerError("Failed to retrieve access token")))
			return
		}
		refreshToken, ok := token["refresh_token"].(string)
		if !ok {
			c.JSON(response.ErrorResponseBuilder(apiError.NewInternalServerError("Failed to retrieve refresh token")))
			return
		}

		userID, firstName, lastName, email, err := auth.ExtractDataFromToken(accessToken)
		if err != nil {
			log.Printf("failed to extract user data from token: %v", err)
			c.JSON(response.ErrorResponseBuilder(apiError.NewInternalServerError("Failed to process user information")))
			return
		}
		log.Printf("UserID: %s, FirstName: %s, LastName: %s, Email: %s", userID, firstName, lastName, email)

		_, err = h.usersUsecase.FindOrCreateUser(userID, email, firstName, lastName)
		if err != nil {
			log.Printf("failed to find or create user: %v", err)
			c.JSON(response.ErrorResponseBuilder(apiError.NewInternalServerError("Failed to set up user account")))
			return
		}

		c.SetCookie("access_token", accessToken, 7*24*3600, "/", ".onepointfive.life", true, true)
		c.SetCookie("refresh_token", refreshToken, 7*24*3600, "/users/auth", ".onepointfive.life", true, true)

		c.JSON(http.StatusOK, gin.H{"message": "Tokens set successfully"})
	}
}

// RefreshAccessToken godoc
// @Summary      Refresh access token
// @Description  Refresh access token using refresh token cookie
// @Tags         users
// @Produce      json
// @Success      200  {object}  map[string]string      "Tokens refreshed successfully"
// @Failure      401  {object}  response.ErrorResponse "Unauthorized"
// @Failure      500  {object}  response.ErrorResponse "Internal server error"
// @Router       /users/auth/refresh-token [get]
func (h *UsersHandlers) RefreshAccessToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		refreshTokenCookie, err := c.Cookie("refresh_token")
		if err != nil {
			c.JSON(response.ErrorResponseBuilder(apiError.NewUnauthorizedError("Session expired, please log in again")))
			return
		}
		url := os.Getenv("CLEARINGHOUSE_URL") + "/auth/refresh-token"
		status, body, err := utils.SendRequestWithAccessToken(url, nil, "GET", refreshTokenCookie)
		if err != nil {
			log.Printf("failed to refresh access token: %v", err)
			c.JSON(response.ErrorResponseBuilder(apiError.NewInternalServerError("Failed to refresh session")))
			return
		}
		if status != http.StatusOK {
			log.Printf("auth service returned status %d: %s", status, string(body))
			c.JSON(response.ErrorResponseBuilder(apiError.NewApiError(status, "Session refresh failed", "Failed to refresh session, please log in again")))
			return
		}
		var token map[string]interface{}
		if err := json.Unmarshal(body, &token); err != nil {
			log.Printf("failed to parse token response: %v", err)
			c.JSON(response.ErrorResponseBuilder(apiError.NewInternalServerError("Failed to process session data")))
			return
		}

		accessToken, ok := token["access_token"].(string)
		if !ok {
			c.JSON(response.ErrorResponseBuilder(apiError.NewInternalServerError("Failed to retrieve new access token")))
			return
		}
		// c.SetCookie("access_token", accessToken, 3600, "/", "localhost", true, true)
		c.SetCookie("access_token", accessToken, 7*24*3600, "/", ".onepointfive.life", true, true)

		c.JSON(http.StatusOK, gin.H{"message": "Tokens refreshed successfully"})
	}
}
