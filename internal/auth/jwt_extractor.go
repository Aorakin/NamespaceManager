package auth

import (
	"fmt"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
)

func ExtractDataFromToken(accessToken string) (uuid.UUID, string, string, string, error) {

	token, err := jwt.Parse(accessToken, func(t *jwt.Token) (interface{}, error) {
		// Ensure token is RS512
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("invalid signing method")
		}
		return PublicKey, nil
	})

	if err != nil || !token.Valid {
		return uuid.Nil, "", "", "", fmt.Errorf("invalid or expired access token")
	}

	// 4. Extract claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return uuid.Nil, "", "", "", fmt.Errorf("invalid token claims")
	}

	// 5. Extract user info
	userIDStr, ok := claims["user_id"].(string)
	if !ok {
		return uuid.Nil, "", "", "", fmt.Errorf("invalid user_id in token")
	}

	userID, err := uuid.Parse(userIDStr)

	firstname, _ := claims["first_name"].(string)
	lastname, _ := claims["last_name"].(string)
	email, _ := claims["email"].(string)

	return userID, firstname, lastname, email, nil
}
