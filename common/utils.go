package common

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rainycape/unidecode"
)

var nonAlpha = regexp.MustCompile(`[^a-zA-Z0-9\. ]+`)

func Slugify(text string) string {
	text = strings.ToLower(text)
	text = unidecode.Unidecode(text)
	text = nonAlpha.ReplaceAllString(text, "")
	text = strings.Trim(text, " ")

	return text
}

func GetRandomString(length int) string {
	buffer := make([]byte, length)
	rand.Read(buffer)
	return fmt.Sprintf("%x", buffer)
}

func GetSecureRandomString(length int) (string, error) {
	buffer := make([]byte, length)
	_, err := rand.Read(buffer)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buffer), nil
}

func GetURL(c *gin.Context, path string) string {
	scheme := "https"
	return fmt.Sprintf("%s://%s/%s", scheme, c.Request.Host, path)
}

func GetPublicURL(c *gin.Context, fileName string) string {
	return GetURL(c, fmt.Sprintf("public/%s", fileName))
}
