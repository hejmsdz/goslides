package tests

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/hejmsdz/goslides/dtos"
	"github.com/stretchr/testify/assert"
)

func ToJsonBody(data gin.H) io.Reader {
	jsonData, _ := json.Marshal(data)
	return strings.NewReader(string(jsonData))
}

type RequestOptions struct {
	Method string
	Path   string
	Body   *gin.H
	Token  string
}

func Request[R any](t *testing.T, testRouter *gin.Engine, opts RequestOptions) (*httptest.ResponseRecorder, *R, *dtos.ErrorResponse) {
	var bodyReader io.Reader
	if opts.Body != nil {
		bodyReader = ToJsonBody(*opts.Body)
	}

	w := httptest.NewRecorder()
	req, err := http.NewRequest(opts.Method, opts.Path, bodyReader)
	assert.NoError(t, err)

	req.Header.Set("Content-Type", "application/json")
	if opts.Token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", opts.Token))
	}

	testRouter.ServeHTTP(w, req)

	var errorResp dtos.ErrorResponse
	var resp R

	if w.Code >= 400 {
		assert.NoError(t, json.NewDecoder(w.Body).Decode(&errorResp))
		return w, nil, &errorResp
	} else if w.Code >= 200 && w.Code < 300 && w.Code != 204 {
		assert.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
		return w, &resp, nil
	} else {
		return w, nil, nil
	}
}
