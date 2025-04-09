package services

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/hejmsdz/goslides/models"
	"gorm.io/gorm"
)

type AuthService struct {
	db                   *gorm.DB
	users                *UsersService
	idTokenValidator     IDTokenValidator
	jwtKey               []byte
	jwtSigningMethod     jwt.SigningMethod
	accessTokenDuration  time.Duration
	refreshTokenDuration time.Duration
}

func NewAuthService(db *gorm.DB, users *UsersService, idTokenValidator IDTokenValidator) *AuthService {
	jwtKey, err := hex.DecodeString(os.Getenv("JWT_KEY"))

	if err != nil {
		panic(fmt.Sprintf("failed to read JWT_KEY: %s", err.Error()))
	}

	if len(jwtKey) < 32 {
		panic("JWT_KEY is too weak: must be at least 32 bytes")
	}

	accessTokenDuration, err := time.ParseDuration(os.Getenv("ACCESS_TOKEN_DURATION"))
	if err != nil {
		panic(fmt.Sprintf("failed to read ACCESS_TOKEN_DURATION: %s", err.Error()))
	}

	refreshTokenDuration, err := time.ParseDuration(os.Getenv("REFRESH_TOKEN_DURATION"))
	if err != nil {
		panic(fmt.Sprintf("failed to read REFRESH_TOKEN_DURATION: %s", err.Error()))
	}

	return &AuthService{
		db:                   db,
		users:                users,
		jwtKey:               jwtKey,
		jwtSigningMethod:     jwt.SigningMethodHS256,
		idTokenValidator:     idTokenValidator,
		accessTokenDuration:  accessTokenDuration,
		refreshTokenDuration: refreshTokenDuration,
	}
}

func (s *AuthService) GetEmailFromGoogleIDToken(ctx context.Context, idToken string) (string, error) {
	payload, err := s.idTokenValidator.Validate(ctx, idToken)
	if err != nil {
		return "", err
	}

	email, ok := payload.Claims["email"].(string)
	if !ok {
		return "", err
	}

	return email, nil
}

func (s *AuthService) GenerateAccessTokenWithExpiration(user *models.User, expiresAt time.Time) (string, error) {
	token := jwt.NewWithClaims(s.jwtSigningMethod, jwt.MapClaims{
		"sub":   user.UUID,
		"admin": user.IsAdmin,
		"exp":   expiresAt.Unix(),
	})

	return token.SignedString(s.jwtKey)
}

func (s *AuthService) GenerateAccessToken(user *models.User) (string, error) {
	return s.GenerateAccessTokenWithExpiration(user, time.Now().Add(s.accessTokenDuration))
}

func (s *AuthService) ValidateAccessToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if token.Method != s.jwtSigningMethod || token.Method.Alg() != s.jwtSigningMethod.Alg() {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return s.jwtKey, nil
	})

	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		if sub, ok := claims["sub"].(string); ok {
			return sub, nil
		}
	}

	return "", errors.New("failed to read claims from token")
}

func (s *AuthService) GenerateRefreshToken(user *models.User) (string, error) {
	rt := models.NewRefreshToken(user.ID)
	result := s.db.Create(&rt)
	if result.Error != nil {
		return "", result.Error
	}

	return rt.Token, nil
}

func (s *AuthService) findRefreshToken(tokenString string) (*models.RefreshToken, error) {
	var rt *models.RefreshToken
	result := s.db.Preload("User").Where("token", tokenString).Where("expires_at > current_timestamp").Take(&rt)

	if result.Error != nil {
		return nil, errors.New("refresh token not found")
	}

	return rt, nil
}

func (s *AuthService) ValidateRefreshToken(tokenString string) (*models.RefreshToken, error) {
	rt, err := s.findRefreshToken(tokenString)
	if err != nil {
		return nil, err
	}

	rt.Regenerate()
	result := s.db.Save(rt)
	if result.Error != nil {
		return nil, errors.New("failed to regenerate the refresh token")
	}

	return rt, nil
}

func (s *AuthService) DeleteRefreshToken(tokenString string) error {
	rt, err := s.findRefreshToken(tokenString)
	if err != nil {
		return err
	}

	result := s.db.Delete(rt)
	return result.Error
}

func (s *AuthService) UserBelongsToTeam(user *models.User, teamID uint) bool {
	if user == nil {
		return false
	}

	var count int64
	s.db.Table("user_teams").Where("user_id = ? AND team_id = ?", user.ID, teamID).Count(&count)
	return count > 0
}

func (s *AuthService) Can(user *models.User, action string, resource *models.Song) bool {
	if user != nil && user.IsAdmin {
		return true
	}

	switch action {
	case "read":
		return resource.TeamID == nil || s.UserBelongsToTeam(user, *resource.TeamID)
	case "create", "update", "delete":
		return resource.TeamID != nil && s.UserBelongsToTeam(user, *resource.TeamID)
	case "override":
		return resource.TeamID == nil
	}
	return false
}

func getBearerToken(c *gin.Context) (string, bool) {
	authHeader := c.Request.Header.Get("Authorization")
	return strings.CutPrefix(authHeader, "Bearer ")
}

func (s *AuthService) handleTokenValidation(c *gin.Context, token string) error {
	userUUID, err := s.ValidateAccessToken(token)

	if err == nil {
		c.Set("userUUID", userUUID)
		return nil
	}

	message := "invalid token"
	if strings.Contains(err.Error(), "token is expired") {
		message = "token expired"
	}
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
		"error": message,
	})

	return err
}

func (s *AuthService) OptionalAuthMiddleware(c *gin.Context) {
	token, isBearer := getBearerToken(c)
	if !isBearer {
		c.Next()
	} else if s.handleTokenValidation(c, token) != nil {
		c.Next()
	}
}

func (s *AuthService) AuthMiddleware(c *gin.Context) {
	token, isBearer := getBearerToken(c)
	if isBearer && s.handleTokenValidation(c, token) == nil {
		c.Next()
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		c.Abort()
	}
}

func (s *AuthService) GetCurrentUser(c *gin.Context) *models.User {
	userUUID := c.GetString("userUUID")
	if userUUID == "" {
		return nil
	}

	user, err := s.users.GetUser(userUUID)
	if err != nil {
		return nil
	}

	return user
}
