package auth

import (
	"connectrpc.com/connect"
	"context"
	"errors"
	"github.com/google/uuid"
)

const TokenHeader = "X-Runner-Token"
const RunIdHeader = "X-Run-Id"

type Auth struct {
	store Store
}

func New(store Store) *Auth {
	return &Auth{store}
}

func (a *Auth) GenerateToken(runId string) (string, error) {
	token := uuid.New()
	if err := a.store.Set(runId, token.String()); err != nil {
		return "", err
	}
	return token.String(), nil
}

func (a *Auth) Validate(runId string, token string) (bool, error) {
	return a.store.Check(runId, token)
}

func (a *Auth) Interceptor() connect.UnaryInterceptorFunc {
	interceptor := func(next connect.UnaryFunc) connect.UnaryFunc {
		return connect.UnaryFunc(func(
			ctx context.Context,
			req connect.AnyRequest,
		) (connect.AnyResponse, error) {
			token := req.Header().Get(TokenHeader)
			runId := req.Header().Get(RunIdHeader)
			if token == "" || runId == "" {
				// Check token in handlers.
				return nil, connect.NewError(
					connect.CodeUnauthenticated,
					errors.New("no token provided"),
				)
			}
			check, err := a.Validate(runId, token)
			if err != nil || !check {
				return nil, connect.NewError(
					connect.CodeUnauthenticated,
					errors.New("no token provided"),
				)
			}
			return next(ctx, req)
		})
	}
	return connect.UnaryInterceptorFunc(interceptor)
}
