package cmd

import (
	"context"
	"errors"
)

func appFromContext(ctx context.Context) (*VervetApp, error) {
	v, ok := ctx.Value(vervetKey).(*VervetApp)
	if !ok {
		return nil, errors.New("could not retrieve vervet app from context")
	}
	return v, nil
}
