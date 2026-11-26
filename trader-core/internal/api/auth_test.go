package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func setupAuthRouter(key string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(APIKeyAuth(key))
	r.GET("/api/ping", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) })
	return r
}

func doAuth(t *testing.T, r *gin.Engine, target, header string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, target, nil)
	if header != "" {
		req.Header.Set("Authorization", header)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestAuthValidHeaderAndQuery(t *testing.T) {
	r := setupAuthRouter("s3cret")

	if w := doAuth(t, r, "/api/ping", "ApiKey s3cret"); w.Code != http.StatusOK {
		t.Fatalf("valid header = %d, want 200; body: %s", w.Code, w.Body.String())
	}
	w := doAuth(t, r, "/api/ping?api_key=s3cret", "")
	if w.Code != http.StatusOK {
		t.Fatalf("valid query = %d, want 200; body: %s", w.Code, w.Body.String())
	}
}

func TestAuthRejectsBadKeys(t *testing.T) {
	r := setupAuthRouter("s3cret")
	cases := []struct{ target, header string }{
		{"/api/ping", ""},                          // no header
		{"/api/ping", "ApiKey wrong"},              // wrong key
		{"/api/ping", "ApiKey s3cret "},            // trailing space
		{"/api/ping", "ApiKey "},                   // empty key
		{"/api/ping", "apikey s3cret"},             // wrong scheme case
		{"/api/ping", "Bearer s3cret"},             // wrong scheme
		{"/api/ping", "ApiKey s3cret-longer"},      // longer (length check)
		{"/api/ping?api_key=wrong", ""},            // wrong query
		{"/api/ping?api_key=", ""},                 // empty query
		{"/api/ping?api_key=s3cretx", ""},          // longer query
	}
	for _, c := range cases {
		if w := doAuth(t, r, c.target, c.header); w.Code != http.StatusUnauthorized {
			t.Errorf("target=%q header=%q -> %d, want 401; body: %s", c.target, c.header, w.Code, w.Body.String())
		}
	}
}

func TestAuthFailsClosedOnEmptyConfiguredKey(t *testing.T) {
	// SERVER_API_KEY unset must NOT open the API: every request is rejected
	r := setupAuthRouter("")
	for _, c := range []struct{ target, header string }{
		{"/api/ping", ""},
		{"/api/ping", "ApiKey "},
		{"/api/ping?api_key=", ""},
		{"/api/ping?api_key=anything", ""},
	} {
		w := doAuth(t, r, c.target, c.header)
		if w.Code != http.StatusInternalServerError {
			t.Errorf("empty configured key: target=%q header=%q -> %d, want 500 (fail closed)", c.target, c.header, w.Code)
		}
	}
}
