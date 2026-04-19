// Package logger provides request-scoped structured logging for MSA.
// Each conversation gets a unique requestID injected into context via WithRequestID,
// and all log entries throughout the pipeline carry this ID via FromCtx.
package logger

import (
	"context"
	"fmt"
	"math/rand"

	log "github.com/sirupsen/logrus"
)

type contextKey string

const requestIDKey contextKey = "req_id"

// WithRequestID injects a new requestID into ctx and returns the new ctx and the ID.
// ID format: 6-character hex string, easy to grep.
func WithRequestID(ctx context.Context) (context.Context, string) {
	id := fmt.Sprintf("%06x", rand.Int63n(0xFFFFFF))
	return context.WithValue(ctx, requestIDKey, id), id
}

// FromCtx retrieves the requestID from ctx and returns a logrus.Entry with the "req" field set.
// If no requestID is found, uses "------" as placeholder.
func FromCtx(ctx context.Context) *log.Entry {
	id, ok := ctx.Value(requestIDKey).(string)
	if !ok || id == "" {
		id = "------"
	}
	return log.WithField("req", id)
}
