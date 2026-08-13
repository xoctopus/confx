package kv

import (
	"context"
	"time"
)

// Store is a key-value store with optional TTL.
type Store interface {
	// Get returns the value for key.
	// The bool is false when the key does not exist.
	Get(ctx context.Context, key string) (string, bool, error)
	// Set writes key to val, overwriting any existing value.
	// ttl <= 0 means the key has no expiration.
	Set(ctx context.Context, key, val string, ttl time.Duration) error
	// SetNX writes key to val only if the key does not already exist.
	// ok is true when the key was set.
	// ttl <= 0 means the key has no expiration.
	SetNX(ctx context.Context, key, val string, ttl time.Duration) (ok bool, err error)
	// Del deletes key.
	// deleted is true when the key existed and was removed.
	// It is not an error if the key does not exist.
	Del(ctx context.Context, key string) (deleted bool, err error)
	// TTL returns the remaining time to live of key.
	//
	// Return cases (when err is nil):
	//  1. key does not exist:           ttl == 0, exists == false
	//  2. key exists, no expiration:    ttl == 0, exists == true
	//  3. key exists with expiration:   ttl > 0,  exists == true
	TTL(ctx context.Context, key string) (ttl time.Duration, exists bool, err error)
}
