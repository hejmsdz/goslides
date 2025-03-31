package routers

import (
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/hejmsdz/goslides/dtos"
	"github.com/hejmsdz/goslides/models"
	"github.com/hejmsdz/goslides/tests"
	"github.com/stretchr/testify/assert"
)

var testRouter *gin.Engine

func TestMain(m *testing.M) {
	testContainer := tests.SetupTestContainer()
	testRouter = tests.SetupTestRouter(testContainer, RegisterAuthRoutes)
	tests.ClearDatabase(testContainer.DB)
	user := &models.User{Email: "john.doe@gmail.com", DisplayName: "John Doe"}
	testContainer.DB.Create(user)

	exitVal := m.Run()
	os.Exit(exitVal)
}

func getAuthMe(t *testing.T, token string) (*httptest.ResponseRecorder, *dtos.AuthMeResponse, *dtos.ErrorResponse) {
	return tests.Request[dtos.AuthMeResponse](t, testRouter, tests.RequestOptions{
		Method: "GET",
		Path:   "/auth/me",
		Token:  token,
	})
}

func postAuthGoogle(t *testing.T, idToken string) (*httptest.ResponseRecorder, *dtos.AuthResponse, *dtos.ErrorResponse) {
	return tests.Request[dtos.AuthResponse](t, testRouter, tests.RequestOptions{
		Method: "POST",
		Path:   "/auth/google",
		Body:   &gin.H{"idToken": idToken},
	})
}

func postAuthRefresh(t *testing.T, refreshToken string) (*httptest.ResponseRecorder, *dtos.AuthResponse, *dtos.ErrorResponse) {
	return tests.Request[dtos.AuthResponse](t, testRouter, tests.RequestOptions{
		Method: "POST",
		Path:   "/auth/refresh",
		Body:   &gin.H{"refreshToken": refreshToken},
	})
}

func deleteAuthRefresh(t *testing.T, refreshToken string) (*httptest.ResponseRecorder, interface{}, *dtos.ErrorResponse) {
	return tests.Request[interface{}](t, testRouter, tests.RequestOptions{
		Method: "DELETE",
		Path:   "/auth/refresh",
		Body:   &gin.H{"refreshToken": refreshToken},
	})
}

func testAuthMe(t *testing.T, token string, shouldBeValid bool, expectedUserEmail string) {
	w, resp, errResp := getAuthMe(t, token)
	if shouldBeValid {
		assert.Equal(t, 200, w.Code)
		assert.Nil(t, errResp)
		assert.Equal(t, expectedUserEmail, resp.Email)
	} else {
		assert.Equal(t, 401, w.Code)
		assert.Equal(t, "invalid token", errResp.Error)
	}
}

func TestGoogle(t *testing.T) {
	w, resp, errResp := postAuthGoogle(t, "valid-token:john.doe@gmail.com")

	assert.Equal(t, 200, w.Code)
	assert.Nil(t, errResp)
	assert.Equal(t, "John Doe", resp.Name)
}

func TestGoogleNonExistentUser(t *testing.T) {
	w, resp, errResp := postAuthGoogle(t, "valid-token:jane.doe@gmail.com")

	assert.Equal(t, 401, w.Code)
	assert.Nil(t, resp)
	assert.Equal(t, "invalid credentials", errResp.Error)
}

func TestMeUnauthenticated(t *testing.T) {
	testAuthMe(t, "", false, "")
}

func TestMeAuthenticated(t *testing.T) {
	_, authResp, _ := postAuthGoogle(t, "valid-token:john.doe@gmail.com")

	testAuthMe(t, authResp.AccessToken, true, "john.doe@gmail.com")
}

func TestRefreshToken(t *testing.T) {
	_, authResp, _ := postAuthGoogle(t, "valid-token:john.doe@gmail.com")
	w, refreshResp, errResp := postAuthRefresh(t, authResp.RefreshToken)

	assert.Equal(t, 200, w.Code, "refresh token returned from /auth/google should be valid")
	assert.Nil(t, errResp)
	assert.NotNil(t, refreshResp)

	w, _, _ = getAuthMe(t, refreshResp.AccessToken)
	assert.Equal(t, 200, w.Code, "access token returned from /auth/refresh should be valid")

	assert.NotEqual(t, refreshResp.RefreshToken, authResp.RefreshToken)

	w, replayedRefreshResp, errResp := postAuthRefresh(t, authResp.RefreshToken)
	assert.Equal(t, 401, w.Code, "should not accept a reused refresh token")
	assert.Equal(t, "refresh token rejected", errResp.Error)
	assert.Nil(t, replayedRefreshResp)
}

func TestDeleteRefreshToken(t *testing.T) {
	_, authResp, _ := postAuthGoogle(t, "valid-token:john.doe@gmail.com")

	w, _, _ := deleteAuthRefresh(t, authResp.RefreshToken)
	assert.Equal(t, 204, w.Code)

	w, _, _ = postAuthRefresh(t, authResp.RefreshToken)
	assert.Equal(t, 401, w.Code, "deleted refresh token should not be usable")
}
