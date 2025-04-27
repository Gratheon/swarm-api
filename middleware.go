package main

import (
	"context"
	"fmt"
	"github.com/Gratheon/swarm-api/logger"
	"github.com/dgrijalva/jwt-go"
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
			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				return []byte(jwtSecret), nil
			})

			unauthorizedBodyResponse := "{\"success\":false, \"errors\":[\"Unauthorized\"]}"

			if token == nil || !token.Valid {
				logger.Error(err.Error())
				http.Error(w, unauthorizedBodyResponse, http.StatusForbidden)
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)

			if !ok {
				logger.Error(err.Error())
				http.Error(w, unauthorizedBodyResponse, http.StatusForbidden)
				return
			}

			if claims["user_id"] == nil {
				http.Error(w, unauthorizedBodyResponse, http.StatusForbidden)
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
