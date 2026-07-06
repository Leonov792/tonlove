package middleware

import (
	"context"
	"net/http"
	"strconv"
)

type ctxKey string

const UserIDKey ctxKey = "user_id"

// Auth validates Telegram initData (simplified for demo)
// Production: verify hash with bot token
func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Header.Get("X-Telegram-User-ID")
		if userID == "" {
			userID = r.URL.Query().Get("user_id")
		}
		if userID == "" {
			http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
			return
		}
		id, err := strconv.ParseInt(userID, 10, 64)
		if err != nil {
			http.Error(w, `{"error":"Invalid user ID"}`, http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), UserIDKey, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetUserID(r *http.Request) int64 {
	if id, ok := r.Context().Value(UserIDKey).(int64); ok {
		return id
	}
	return 0
}
