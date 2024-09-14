package middlewares

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/Dnlbb/link-shortener/internal/config"
)

func SignData(data string) string {
	h := hmac.New(sha256.New, []byte(config.Conf.Key))
	h.Write([]byte(data))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func verifyData(data, signature string) bool {
	expectedSignature := SignData(data)
	return hmac.Equal([]byte(expectedSignature), []byte(signature))
}

func CreateCookie(w http.ResponseWriter) string {
	UserID := uuid.New().String()
	signature := SignData(UserID)
	cookieValue := fmt.Sprintf("%s|%s", UserID, signature)
	http.SetCookie(w, &http.Cookie{
		Name:    "session",
		Value:   cookieValue,
		Expires: time.Now().Add(24 * time.Hour),
	})
	return UserID
}

func ExtractUserIDFromCookie(r *http.Request) (string, error) {
	cookie, err := r.Cookie("session")
	if err != nil {
		return "", err
	}
	parts := string(cookie.Value)

	splitParts := strings.Split(parts, "|")
	if len(splitParts) != 2 {
		return "", fmt.Errorf("invalid cookie format")
	}

	userID, signature := splitParts[0], splitParts[1]
	if !verifyData(userID, signature) {
		return "", fmt.Errorf("invalid signature")
	}

	return userID, nil

}

type contextKey string

const UserIDKey contextKey = "userID"

func MiddlewareAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()
		userID, err := ExtractUserIDFromCookie(r)
		if r.URL.Path == "/api/user/urls" && err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		} else if err != nil {
			userID = CreateCookie(w)
		}
		ctx = context.WithValue(ctx, UserIDKey, userID)
		log.Printf("userID in context: %v", userID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
