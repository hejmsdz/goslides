package common

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/rainycape/unidecode"
)

var nonAlpha = regexp.MustCompile(`[^a-zA-Z0-9\. ]+`)
var nonAlphaPipe = regexp.MustCompile(`[^a-zA-Z0-9\.\| ]+`)

func Slugify(text string, allowPipe bool) string {
	text = strings.ToLower(text)
	text = unidecode.Unidecode(text)
	if allowPipe {
		text = nonAlphaPipe.ReplaceAllString(text, "")
	} else {
		text = nonAlpha.ReplaceAllString(text, "")
	}
	text = strings.Trim(text, " ")

	return text
}

func GetSecureRandomString(length int) (string, error) {
	buffer := make([]byte, length)
	_, err := rand.Read(buffer)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buffer), nil
}

func GetFrontendURL(path string) string {
	return fmt.Sprintf("%s/%s", os.Getenv("FRONTEND_URL"), path)
}

func GetPublicURL(fileName string) string {
	return fmt.Sprintf("%s/%s", os.Getenv("PUBLIC_URL"), fileName)
}
