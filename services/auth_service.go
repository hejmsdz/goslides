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
	"github.com/hejmsdz/goslides/common"
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

type UserInfo struct {
	Email string
	Name  string
}

func (s *AuthService) GetUserInfoFromGoogleIDToken(ctx context.Context, idToken string) (UserInfo, error) {
	var userInfo UserInfo
	payload, err := s.idTokenValidator.Validate(ctx, idToken)
	if err != nil {
		return userInfo, err
	}

	email, ok := payload.Claims["email"].(string)
	if !ok {
		return userInfo, errors.New("email not found in payload")
	}

	emailVerified, ok := payload.Claims["email_verified"].(bool)
	if !ok || !emailVerified {
		return userInfo, errors.New("email not verified")
	}

	name, ok := payload.Claims["name"].(string)
	if !ok || name == "" {
		name = email
	}

	userInfo.Email = email
	userInfo.Name = name

	return userInfo, nil
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

func (s *AuthService) GenerateNonce(user *models.User) (string, error) {
	nonce := models.NewNonce(user.ID)
	result := s.db.Create(&nonce)
	if result.Error != nil {
		fmt.Println(result.Error)
		return "", result.Error
	}

	return nonce.Token, nil
}

func (s *AuthService) GetUserFromNonce(token string) (*models.User, error) {
	var nonce models.Nonce
	result := s.db.Preload("User").Where("token = ? AND expires_at > current_timestamp", token).Take(&nonce)
	if result.Error != nil {
		return nil, common.NewAPIError(http.StatusUnauthorized, "invalid nonce", result.Error)
	}

	s.db.Delete(&nonce)

	return nonce.User, nil
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

	if strings.Contains(err.Error(), "token is expired") {
		return errors.New("token expired")
	}

	return errors.New("invalid token")
}

func (s *AuthService) OptionalAuthMiddleware(c *gin.Context) {
	token, isBearer := getBearerToken(c)
	if !isBearer {
		c.Next()
	} else if err := s.handleTokenValidation(c, token); err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
	} else {
		c.Next()
	}
}

func (s *AuthService) AuthMiddleware(c *gin.Context) {
	token, isBearer := getBearerToken(c)
	if !isBearer {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
	} else if err := s.handleTokenValidation(c, token); err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
	} else {
		c.Next()
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
