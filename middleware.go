package main

import (
	"context"
	"fmt"
	"github.com/Gratheon/swarm-api/logger"
	"github.com/golang-jwt/jwt/v5"
	"github.com/spf13/viper"
	"net/http"
)

//func logToBugsnag(next http.Handler) http.Handler {
//	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//		ctx := r.Context()
//		next.ServeHTTP(w, r.WithContext(ctx))
//	})
//}

type Input struct {
	OperationName string `json:"operationName"` //IntrospectionQuery
}

func authMiddleware(next http.Handler) http.Handler {
	jwtSecret := viper.GetString("jwt_key")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uid := r.Header.Get("internal-userId")

		if uid == "" {
			tokenString := r.Header.Get("token")

			// SECURITY FIX: Parse with explicit algorithm validation
			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				// Validate the signing algorithm to prevent algorithm confusion attacks
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				return []byte(jwtSecret), nil
			})

			unauthorizedBodyResponse := "{\"success\":false, \"errors\":[\"Unauthorized\"]}"

			// IMPROVED: Better error handling
			if err != nil {
				logger.Error(fmt.Sprintf("JWT parse error: %v", err))
				http.Error(w, unauthorizedBodyResponse, http.StatusForbidden)
				return
			}

			if token == nil || !token.Valid {
				logger.Error("Invalid or nil token")
				http.Error(w, unauthorizedBodyResponse, http.StatusForbidden)
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)

			if !ok {
				logger.Error("Failed to parse JWT claims")
				http.Error(w, unauthorizedBodyResponse, http.StatusForbidden)
				return
			}

			if claims["user_id"] == nil {
				logger.Error("Missing user_id in JWT claims")
				http.Error(w, unauthorizedBodyResponse, http.StatusForbidden)
				return
			}

			// Get context from an HTTP request
			ctx := context.WithValue(r.Context(), "userID", fmt.Sprintf("%v", claims["user_id"]))
			next.ServeHTTP(w, r.WithContext(ctx))
		} else {
			ctx := context.WithValue(r.Context(), "userID", uid)
			next.ServeHTTP(w, r.WithContext(ctx))
		}
	})
}
