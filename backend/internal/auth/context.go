package auth

import "context"

type contextKey string

const (
	userIDContextKey contextKey = "userID"
	isDemoContextKey contextKey = "isDemo"
)

func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDContextKey, userID)
}

func WithIdentity(ctx context.Context, userID string, isDemo bool) context.Context {
	ctx = context.WithValue(ctx, userIDContextKey, userID)
	return context.WithValue(ctx, isDemoContextKey, isDemo)
}

func UserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(userIDContextKey).(string)
	if !ok || userID == "" {
		return "", false
	}
	return userID, true
}

func IsDemoFromContext(ctx context.Context) bool {
	isDemo, _ := ctx.Value(isDemoContextKey).(bool)
	return isDemo
}
