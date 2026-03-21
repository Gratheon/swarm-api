//go:build !integration
// +build !integration

package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthMiddleware_AllowsInternalHeaders(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("internal-userId", "internal-user")
	req.Header.Set("internal-billing-plan", "starter")

	rr := httptest.NewRecorder()
	var gotUserID string
	var gotBillingPlan string

	handler := authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUserID, _ = r.Context().Value("userID").(string)
		gotBillingPlan, _ = r.Context().Value("billingPlan").(string)
		w.WriteHeader(http.StatusNoContent)
	}))

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNoContent, rr.Code)
	assert.Equal(t, "internal-user", gotUserID)
	assert.Equal(t, "starter", gotBillingPlan)
}

func TestAuthMiddleware_RejectsInvalidToken(t *testing.T) {
	viper.Set("jwt_key", "test-secret")
	t.Cleanup(viper.Reset)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("token", "not-a-token")
	rr := httptest.NewRecorder()

	handler := authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
	assert.Contains(t, rr.Body.String(), "Unauthorized")
}

func TestAuthMiddleware_RejectsUnexpectedSigningMethod(t *testing.T) {
	viper.Set("jwt_key", "test-secret")
	t.Cleanup(viper.Reset)

	token := jwt.New(jwt.SigningMethodNone)
	tokenString, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("token", tokenString)
	rr := httptest.NewRecorder()

	handler := authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
	assert.Contains(t, rr.Body.String(), "Unauthorized")
}

func TestAuthMiddleware_AcceptsValidJWTWithBillingPlan(t *testing.T) {
	viper.Set("jwt_key", "test-secret")
	t.Cleanup(viper.Reset)

	tokenString := signedJWT(t, map[string]interface{}{
		"user_id":      123,
		"billing_plan": "professional",
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("token", tokenString)
	rr := httptest.NewRecorder()

	var gotUserID string
	var gotBillingPlan string
	handler := authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUserID, _ = r.Context().Value("userID").(string)
		gotBillingPlan, _ = r.Context().Value("billingPlan").(string)
		w.WriteHeader(http.StatusNoContent)
	}))

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNoContent, rr.Code)
	assert.Equal(t, "123", gotUserID)
	assert.Equal(t, "professional", gotBillingPlan)
}

func TestAuthMiddleware_AcceptsValidJWTWithLegacyBillingPlanField(t *testing.T) {
	viper.Set("jwt_key", "test-secret")
	t.Cleanup(viper.Reset)

	tokenString := signedJWT(t, map[string]interface{}{
		"user_id":     "u-1",
		"billingPlan": "hobbyist",
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("token", tokenString)
	rr := httptest.NewRecorder()

	var gotBillingPlan string
	handler := authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotBillingPlan, _ = r.Context().Value("billingPlan").(string)
		w.WriteHeader(http.StatusNoContent)
	}))

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNoContent, rr.Code)
	assert.Equal(t, "hobbyist", gotBillingPlan)
}

func TestAuthMiddleware_RejectsTokenWithoutUserID(t *testing.T) {
	viper.Set("jwt_key", "test-secret")
	t.Cleanup(viper.Reset)

	tokenString := signedJWT(t, map[string]interface{}{
		"billing_plan": "starter",
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("token", tokenString)
	rr := httptest.NewRecorder()

	handler := authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
	assert.Contains(t, rr.Body.String(), "Unauthorized")
}

func signedJWT(t *testing.T, claims map[string]interface{}) string {
	t.Helper()

	viperSecret := viper.GetString("jwt_key")
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims(claims))
	tokenString, err := token.SignedString([]byte(viperSecret))
	require.NoError(t, err, fmt.Sprintf("failed to sign JWT with claims: %v", claims))
	return tokenString
}
