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
	"google.golang.org/api/idtoken"
)

type AuthService struct {
	Users            *UsersService
	jwtKey           []byte
	jwtSigningMethod jwt.SigningMethod
	googleClientID   string
}

func NewAuthService(users *UsersService) *AuthService {
	jwtKey, err := hex.DecodeString(os.Getenv("JWT_KEY"))

	if err != nil {
		panic(fmt.Sprintf("failed to read JWT_KEY: %s", err.Error()))
	}

	if len(jwtKey) < 32 {
		panic("JWT_KEY is too weak: must be at least 32 bytes")
	}

	return &AuthService{
		Users:            users,
		jwtKey:           jwtKey,
		jwtSigningMethod: jwt.SigningMethodHS256,
		googleClientID:   os.Getenv("GOOGLE_CLIENT_ID"),
	}
}

func (s *AuthService) GetEmailFromGoogleIDToken(ctx context.Context, idToken string) (string, error) {
	payload, err := idtoken.Validate(ctx, idToken, s.googleClientID)
	if err != nil {
		return "", err
	}

	email, ok := payload.Claims["email"].(string)
	if !ok {
		return "", err
	}

	return email, nil
}

func (s *AuthService) GenerateToken(user *models.User) (string, error) {
	token := jwt.NewWithClaims(s.jwtSigningMethod, jwt.MapClaims{
		"sub": user.UUID,
		"exp": time.Now().Add(time.Hour * 3).Unix(),
	})

	return token.SignedString(s.jwtKey)
}

func (s *AuthService) ValidateToken(tokenString string) (string, error) {
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

func (s *AuthService) AuthMiddleware(c *gin.Context) {
	authHeader := c.Request.Header.Get("Authorization")
	token, isBearer := strings.CutPrefix(authHeader, "Bearer ")

	if !isBearer {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		c.Abort()
		return
	}

	userUuid, err := s.ValidateToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		c.Abort()
		return
	}

	user, err := s.Users.GetUser(userUuid)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		c.Abort()
		return
	}
	c.Set("user", user)
	c.Next()
}
