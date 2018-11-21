package v1alpha1

import (
	"context"

	"github.com/sirupsen/logrus"
)

var (
	// L is an alias for the the standard logger.
	L = logrus.NewEntry(logrus.StandardLogger())
)

type (
	loggerKey struct{}
)

// WithLogger returns a new context with the provided logger.
func WithLogger(ctx context.Context, logger *logrus.Entry) context.Context {
	return context.WithValue(ctx, loggerKey{}, logger)
}

// GetLogger returns the logger from the context
func GetLogger(ctx context.Context) *logrus.Entry {
	logger := ctx.Value(loggerKey{})

	if logger == nil {
		return L
	}

	return logger.(*logrus.Entry)
}
