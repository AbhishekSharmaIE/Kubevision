package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const sessionPrefix = "kv:sess:"

func sessionKey(userID uuid.UUID, jti string) string {
	return fmt.Sprintf("%s%s:%s", sessionPrefix, userID.String(), jti)
}

func sessionIndexPattern(userID uuid.UUID) string {
	return fmt.Sprintf("%s%s:*", sessionPrefix, userID.String())
}

// StoreSession records an active access token session in Redis.
func StoreSession(ctx context.Context, rdb *redis.Client, userID uuid.UUID, jti string, ttl time.Duration) error {
	return rdb.Set(ctx, sessionKey(userID, jti), "1", ttl).Err()
}

// ValidateSession returns true if the jti is still active for the user.
func ValidateSession(ctx context.Context, rdb *redis.Client, userID uuid.UUID, jti string) bool {
	if jti == "" {
		return false
	}
	v, err := rdb.Get(ctx, sessionKey(userID, jti)).Result()
	return err == nil && v != ""
}

// RevokeSession removes one session.
func RevokeSession(ctx context.Context, rdb *redis.Client, userID uuid.UUID, jti string) error {
	return rdb.Del(ctx, sessionKey(userID, jti)).Err()
}

// RevokeAllSessions removes every session for a user (logout-all devices).
func RevokeAllSessions(ctx context.Context, rdb *redis.Client, userID uuid.UUID) error {
	iter := rdb.Scan(ctx, 0, sessionIndexPattern(userID), 0).Iterator()
	for iter.Next(ctx) {
		if err := rdb.Del(ctx, iter.Val()).Err(); err != nil {
			return err
		}
	}
	return iter.Err()
}
