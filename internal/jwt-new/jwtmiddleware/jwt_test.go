package jwtmiddleware_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/linemk/avito-shop/internal/jwt-new/jwtmiddleware"
	"github.com/stretchr/testify/assert"
)

// createTestToken создаёт JWT-токен с заданным userID и секретом.
func createTestToken(userID int64, secret string) (string, error) {
	claims := jwt.MapClaims{
		"sub": fmt.Sprintf("%d", userID),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func TestJWTMiddleware_MissingAuthorization(t *testing.T) {
	os.Setenv("JWT_SECRET", "testsecret")
	defer os.Unsetenv("JWT_SECRET")

	middleware := jwtmiddleware.NewJWTMiddleware()
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code, "Expected unauthorized status when no token provided")
	assert.True(t, strings.Contains(rr.Body.String(), "missing token"))
}

func TestJWTMiddleware_InvalidAuthorizationFormat(t *testing.T) {
	os.Setenv("JWT_SECRET", "testsecret")
	defer os.Unsetenv("JWT_SECRET")

	middleware := jwtmiddleware.NewJWTMiddleware()
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "InvalidFormat")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code, "Expected unauthorized status for invalid token format")
	assert.True(t, strings.Contains(rr.Body.String(), "invalid token format"))
}

func TestJWTMiddleware_InvalidToken(t *testing.T) {
	os.Setenv("JWT_SECRET", "testsecret")
	defer os.Unsetenv("JWT_SECRET")

	middleware := jwtmiddleware.NewJWTMiddleware()
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.value")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code, "Expected unauthorized status for invalid token")
	assert.True(t, strings.Contains(rr.Body.String(), "invalid token"))
}

func TestJWTMiddleware_ValidToken(t *testing.T) {
	secret := "testsecret"
	os.Setenv("JWT_SECRET", secret)
	defer os.Unsetenv("JWT_SECRET")

	// Создаём токен для userID=123.
	tokenStr, err := createTestToken(123, secret)
	assert.NoError(t, err)

	middleware := jwtmiddleware.NewJWTMiddleware()
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := jwtmiddleware.FromContext(r.Context())
		if !ok {
			http.Error(w, "userID not found", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(string(rune(userID))))
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code, "Expected OK status for valid token")
}

func TestFromContext(t *testing.T) {
	ctx := context.WithValue(context.Background(), jwtmiddleware.UserIDKey, int64(456))
	userID, ok := jwtmiddleware.FromContext(ctx)
	assert.True(t, ok, "Expected to retrieve userID from context")
	assert.Equal(t, int64(456), userID, "Expected userID to match")
}
