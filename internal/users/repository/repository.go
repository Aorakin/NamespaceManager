package repository

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/NamespaceManager/internal/models"
	"github.com/NamespaceManager/internal/users/interfaces"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"gorm.io/gorm"
)

type UsersRepository struct {
	db *gorm.DB
}

func NewUsersRepository(db *gorm.DB) interfaces.UsersRepository {
	return &UsersRepository{db: db}
}

func (r *UsersRepository) GetUserGoogle(token *oauth2.Token) (map[string]interface{}, error) {
	client := oauth2.NewClient(context.TODO(), oauth2.StaticTokenSource(token))
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("failed to fetch user info")
	}

	var userInfo map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&userInfo)
	if err != nil {
		return nil, err
	}
	
	// Handle user creation/login with OAuth
	googleID := userInfo["id"].(string)
	email := userInfo["email"].(string)
	name := userInfo["name"].(string)
	
	user, err := r.CreateOrGetOAuthUser(googleID, email, name, models.ProviderGoogle, token)
	if err != nil {
		return nil, err
	}
	
	userInfo["user"] = user
	return userInfo, nil
}

func (r *UsersRepository) CreateOrGetOAuthUser(providerID, email, username string, provider models.ProviderType, token *oauth2.Token) (*models.User, error) {
	// First, try to find existing user provider record
	var userProvider models.UserProvider
	err := r.db.Preload("User").First(&userProvider, "provider = ? AND provider_id = ?", provider, providerID).Error
	
	if err == nil {
		// User exists, update tokens
		userProvider.AccessToken = token.AccessToken
		userProvider.RefreshToken = token.RefreshToken
		userProvider.ExpiresAt = &token.Expiry
		r.db.Save(&userProvider)
		return &userProvider.User, nil
	}
	
	// Check if user exists by email
	var existingUser models.User
	err = r.db.Preload("UserProviders").First(&existingUser, "email = ?", email).Error
	
	if err == nil {
		// User exists, add new provider
		userProvider = models.UserProvider{
			UserID:       existingUser.ID,
			Provider:     provider,
			ProviderID:   providerID,
			AccessToken:  token.AccessToken,
			RefreshToken: token.RefreshToken,
			ExpiresAt:    &token.Expiry,
		}
		if err := r.db.Create(&userProvider).Error; err != nil {
			return nil, err
		}
		existingUser.UserProviders = append(existingUser.UserProviders, userProvider)
		return &existingUser, nil
	}
	
	// Create new user with provider
	newUser := models.User{
		Username: username,
		Email:    email,
		Role:     models.UserRoleUser,
	}
	
	if err := r.db.Create(&newUser).Error; err != nil {
		return nil, err
	}
	
	// Create provider record
	userProvider = models.UserProvider{
		UserID:       newUser.ID,
		Provider:     provider,
		ProviderID:   providerID,
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		ExpiresAt:    &token.Expiry,
	}
	
	if err := r.db.Create(&userProvider).Error; err != nil {
		return nil, err
	}
	
	newUser.UserProviders = []models.UserProvider{userProvider}
	return &newUser, nil
}

func (r *UsersRepository) GetByEmail(email string) (*models.User, error) {
	var user *models.User
	if err := r.db.Preload("UserProviders").First(&user, "email = ?", email).Error; err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UsersRepository) Create(user *models.User) error {
	if err := r.db.Create(&user).Error; err != nil {
		return err
	}
	return nil
}

func (r *UsersRepository) Delete(id uuid.UUID) error {
	if err := r.db.Delete(&models.User{}, "id = ?", id).Error; err != nil {
		return err
	}
	return nil
}

func (r *UsersRepository) GetUser(id uuid.UUID) (*models.User, error) {
	var user *models.User
	if err := r.db.First(&user, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UsersRepository) GetByUsername(username string) (*models.User, error) {
	var user *models.User
	if err := r.db.First(&user, "username = ?", username).Error; err != nil {
		return nil, err
	}
	return user, nil
}
