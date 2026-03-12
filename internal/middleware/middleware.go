// package middleware

// import (
// 	"net/http"

// 	"github.com/NamespaceManager/internal/utils"
// 	"github.com/gin-gonic/gin"
// )

//	func AuthMiddleware() gin.HandlerFunc {
//		return func(c *gin.Context) {
//			id := utils.GetSession(c, "userID")
//			if id == nil {
//				c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
//				c.Abort()
//				return
//			}
//			c.Next()
//		}
//	}
package middleware

import (
	"fmt"
	"strings"

	"github.com/NamespaceManager/internal/auth"
	apiError "github.com/NamespaceManager/pkg/api_error"
	"github.com/NamespaceManager/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		var accessToken string

		// 1. Try to get token from Authorization header first
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			// Parse "Bearer <token>" format
			parts := strings.Split(authHeader, " ")
			if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
				accessToken = parts[1]
			}
		}

		// 2. If no token in header, try to get from cookie
		if accessToken == "" {
			var err error
			accessToken, err = c.Cookie("access_token")
			if err != nil {
				c.JSON(response.ErrorResponseBuilder(apiError.NewUnauthorizedError("Missing access token")))
				c.Abort()
				return
			}
		}

		// 3. Parse JWT using your RSA public key
		token, err := jwt.Parse(accessToken, func(t *jwt.Token) (interface{}, error) {
			// Ensure token is RS512
			if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("invalid signing method")
			}
			return auth.PublicKey, nil
		})

		if err != nil || !token.Valid {
			c.JSON(response.ErrorResponseBuilder(apiError.NewUnauthorizedError("Invalid or expired access token")))
			c.Abort()
			return
		}

		// 4. Extract claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(response.ErrorResponseBuilder(apiError.NewUnauthorizedError("Invalid token")))
			c.Abort()
			return
		}

		// 5. Extract user info
		userIDStr, ok := claims["user_id"].(string)
		if !ok {
			c.JSON(response.ErrorResponseBuilder(apiError.NewUnauthorizedError("Invalid token")))
			c.Abort()
			return
		}

		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			c.JSON(response.ErrorResponseBuilder(apiError.NewUnauthorizedError("Invalid token")))
			c.Abort()
			return
		}

		firstname, _ := claims["first_name"].(string)
		lastname, _ := claims["last_name"].(string)
		email, _ := claims["email"].(string)

		if userID == uuid.Nil {
			c.JSON(response.ErrorResponseBuilder(apiError.NewUnauthorizedError("Invalid token")))
			c.Abort()
			return
		}

		// 6. Store into Gin context
		c.Set("userID", userID)
		c.Set("firstname", firstname)
		c.Set("lastname", lastname)
		c.Set("email", email)
		c.Set("accessToken", accessToken)

		c.Next()
	}
}
