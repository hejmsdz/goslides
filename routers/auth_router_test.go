package routers_test

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/hejmsdz/goslides/dtos"
	"github.com/hejmsdz/goslides/models"
	"github.com/hejmsdz/goslides/tests"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func getAuthMe(t *testing.T, router *gin.Engine, token string) (*httptest.ResponseRecorder, *dtos.UserMeResponse, *dtos.ErrorResponse) {
	return tests.Request[dtos.UserMeResponse](t, router, tests.RequestOptions{
		Method: "GET",
		Path:   "/v2/users/me",
		Token:  token,
	})
}

func postAuthGoogle(t *testing.T, router *gin.Engine, idToken string) (*httptest.ResponseRecorder, *dtos.AuthResponse, *dtos.ErrorResponse) {
	return tests.Request[dtos.AuthResponse](t, router, tests.RequestOptions{
		Method: "POST",
		Path:   "/v2/auth/google",
		Body:   &gin.H{"idToken": idToken},
	})
}

func postAuthRefresh(t *testing.T, router *gin.Engine, refreshToken string) (*httptest.ResponseRecorder, *dtos.AuthResponse, *dtos.ErrorResponse) {
	return tests.Request[dtos.AuthResponse](t, router, tests.RequestOptions{
		Method: "POST",
		Path:   "/v2/auth/refresh",
		Body:   &gin.H{"refreshToken": refreshToken},
	})
}

func deleteAuthRefresh(t *testing.T, router *gin.Engine, refreshToken string) (*httptest.ResponseRecorder, interface{}, *dtos.ErrorResponse) {
	return tests.Request[interface{}](t, router, tests.RequestOptions{
		Method: "DELETE",
		Path:   "/v2/auth/refresh",
		Body:   &gin.H{"refreshToken": refreshToken},
	})
}

func postAuthNonce(t *testing.T, router *gin.Engine, token string) (*httptest.ResponseRecorder, *dtos.AuthNonceResponse, *dtos.ErrorResponse) {
	return tests.Request[dtos.AuthNonceResponse](t, router, tests.RequestOptions{
		Method: "POST",
		Path:   "/v2/auth/nonce",
		Token:  token,
	})
}

func postAuthNonceVerify(t *testing.T, router *gin.Engine, nonce string) (*httptest.ResponseRecorder, *dtos.AuthResponse, *dtos.ErrorResponse) {
	return tests.Request[dtos.AuthResponse](t, router, tests.RequestOptions{
		Method: "POST",
		Path:   "/v2/auth/nonce/verify",
		Body:   &gin.H{"nonce": nonce},
	})
}

func testAuthMe(t *testing.T, router *gin.Engine, token string, shouldBeValid bool, expectedUserEmail string) {
	w, resp, errResp := getAuthMe(t, router, token)
	if shouldBeValid {
		assert.Equal(t, 200, w.Code)
		assert.Nil(t, errResp)
		assert.Equal(t, expectedUserEmail, resp.Email)
	} else {
		assert.Equal(t, 401, w.Code)
		assert.Equal(t, "invalid token", errResp.Error)
	}
}

func createTestUser(t *testing.T, db *gorm.DB) *models.User {
	user := &models.User{Email: "john.doe@gmail.com", DisplayName: "John Doe"}
	err := db.Create(user).Error
	assert.NoError(t, err)
	return user
}

func TestAuthRouter(t *testing.T) {
	te := tests.NewTestEnvironment(t)

	t.Run("/auth/google", func(t *testing.T) {
		te.Run("existing user", func(t *testing.T, tce *tests.TestCaseEnvironment) {
			createTestUser(t, tce.DB)

			w, resp, errResp := postAuthGoogle(t, tce.App, "valid-token:john.doe@gmail.com")
			assert.Equal(t, 200, w.Code)
			assert.Nil(t, errResp)
			assert.Equal(t, "John Doe", resp.User.DisplayName)
			assert.False(t, resp.IsNewUser)
		})

		te.Run("non-existent user", func(t *testing.T, tce *tests.TestCaseEnvironment) {
			w, resp, errResp := postAuthGoogle(t, tce.App, "valid-token:jane.doe@gmail.com")
			assert.Equal(t, 201, w.Code)
			assert.Nil(t, errResp)
			assert.Equal(t, "jane.doe@gmail.com", resp.User.Email)
			assert.True(t, resp.IsNewUser)
		})
	})

	t.Run("/auth/me", func(t *testing.T) {
		te.Run("unauthenticated", func(t *testing.T, tce *tests.TestCaseEnvironment) {
			testAuthMe(t, tce.App, "", false, "")
		})

		te.Run("authenticated", func(t *testing.T, tce *tests.TestCaseEnvironment) {
			user := createTestUser(t, tce.DB)
			token, err := tce.Container.Auth.GenerateAccessToken(user)
			assert.NoError(t, err)
			testAuthMe(t, tce.App, token, true, "john.doe@gmail.com")
		})
	})

	t.Run("/auth/refresh", func(t *testing.T) {
		te.Run("refresh token flow", func(t *testing.T, tce *tests.TestCaseEnvironment) {
			createTestUser(t, tce.DB)

			w, authResp, _ := postAuthGoogle(t, tce.App, "valid-token:john.doe@gmail.com")
			assert.Equal(t, 200, w.Code, "refresh token returned from /auth/google should be valid")

			w, refreshResp, errResp := postAuthRefresh(t, tce.App, authResp.RefreshToken)
			assert.Equal(t, 200, w.Code, "access token returned from /auth/refresh should be valid")
			assert.Nil(t, errResp)
			assert.NotNil(t, refreshResp)

			w, _, _ = getAuthMe(t, tce.App, refreshResp.AccessToken)
			assert.Equal(t, 200, w.Code, "access token returned from /auth/refresh should be valid")
			assert.NotEqual(t, refreshResp.RefreshToken, authResp.RefreshToken)

			w, replayedRefreshResp, errResp := postAuthRefresh(t, tce.App, authResp.RefreshToken)
			assert.Equal(t, 401, w.Code, "should not accept a reused refresh token")
			assert.Equal(t, "refresh token rejected", errResp.Error)
			assert.Nil(t, replayedRefreshResp)
		})

		te.Run("delete refresh token", func(t *testing.T, tce *tests.TestCaseEnvironment) {
			createTestUser(t, tce.DB)

			w, authResp, _ := postAuthGoogle(t, tce.App, "valid-token:john.doe@gmail.com")
			assert.Equal(t, 200, w.Code)

			w, _, _ = deleteAuthRefresh(t, tce.App, authResp.RefreshToken)
			assert.Equal(t, 204, w.Code)

			w, _, _ = postAuthRefresh(t, tce.App, authResp.RefreshToken)
			assert.Equal(t, 401, w.Code, "deleted refresh token should not be usable")
		})
	})

	t.Run("/auth/nonce", func(t *testing.T) {
		te.Run("unauthenticated", func(t *testing.T, tce *tests.TestCaseEnvironment) {
			w, _, errResp := postAuthNonce(t, tce.App, "")
			assert.Equal(t, 401, w.Code)
			assert.Equal(t, "invalid token", errResp.Error)
		})

		te.Run("authenticated", func(t *testing.T, tce *tests.TestCaseEnvironment) {
			user := createTestUser(t, tce.DB)
			token, err := tce.Container.Auth.GenerateAccessToken(user)
			assert.NoError(t, err)

			w, resp, errResp := postAuthNonce(t, tce.App, token)
			assert.Equal(t, 200, w.Code)
			assert.Nil(t, errResp)
			assert.NotEmpty(t, resp.Nonce)
		})
	})

	t.Run("/auth/nonce/verify", func(t *testing.T) {
		te.Run("invalid nonce", func(t *testing.T, tce *tests.TestCaseEnvironment) {
			w, _, errResp := postAuthNonceVerify(t, tce.App, "invalid-nonce")
			assert.Equal(t, 401, w.Code)
			assert.Equal(t, "invalid nonce", errResp.Error)
		})

		te.Run("valid nonce flow", func(t *testing.T, tce *tests.TestCaseEnvironment) {
			// Create user and get auth token
			user := createTestUser(t, tce.DB)
			token, err := tce.Container.Auth.GenerateAccessToken(user)
			assert.NoError(t, err)

			// Get a nonce
			w, nonceResp, errResp := postAuthNonce(t, tce.App, token)
			assert.Equal(t, 200, w.Code)
			assert.Nil(t, errResp)
			assert.NotEmpty(t, nonceResp.Nonce)

			// Verify the nonce
			w, authResp, errResp := postAuthNonceVerify(t, tce.App, nonceResp.Nonce)
			assert.Equal(t, 200, w.Code)
			assert.Nil(t, errResp)
			assert.NotEmpty(t, authResp.AccessToken)
			assert.NotEmpty(t, authResp.RefreshToken)
			assert.Equal(t, user.Email, authResp.User.Email)

			// Verify the new access token works
			w, meResp, errResp := getAuthMe(t, tce.App, authResp.AccessToken)
			assert.Equal(t, 200, w.Code)
			assert.Nil(t, errResp)
			assert.Equal(t, user.Email, meResp.Email)

			// Verify the nonce can't be reused
			w, _, errResp = postAuthNonceVerify(t, tce.App, nonceResp.Nonce)
			assert.Equal(t, 401, w.Code)
			assert.Equal(t, "invalid nonce", errResp.Error)
		})
	})
}
