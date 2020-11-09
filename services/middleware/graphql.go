package middleware

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.uber.org/zap"
)

type gqlLog struct {
	Op    string                 `json:"op"`
	Name  string                 `json:"name,omitempty"`
	Field string                 `json:"field"`
	Error error                  `json:"error,omitempty"`
	Vars  map[string]interface{} `json:"vars,omitempty"`
}

// DefaultGQLErrorPresenter formats graphql errors
func (s *Service) DefaultGQLErrorPresenter(ctx context.Context, err error) *gqlerror.Error {
	// Uncomment this line to print resolver error details in the console
	// fmt.Printf("Error in GraphQL Resolver: %s", err.Error())
	return graphql.DefaultErrorPresenter(ctx, err)
}

// DefaultGQLRecoverFunc is called when a graphql request panics
func (s *Service) DefaultGQLRecoverFunc(logger *zap.Logger) graphql.RecoverFunc {
	return func(ctx context.Context, errT interface{}) error {
		if err, ok := errT.(error); ok {
			logger.Error("http recovered panic", zap.Error(err))
		} else {
			logger.Error("http recovered panic", zap.Reflect("error", errT))
		}
		return gqlerror.Errorf("Internal server error")
	}
}

// QueryLoggingGQLMiddleware logs sensible info about every GraphQL request
func (s *Service) QueryLoggingGQLMiddleware(ctx context.Context, next graphql.ResponseHandler) *graphql.Response {
	reqCtx := graphql.GetRequestContext(ctx)
	SetTagsPayload(ctx, s.logInfoFromGQLRequestContext(reqCtx))
	return next(ctx)
}

func (s *Service) logInfoFromGQLRequestContext(ctx *graphql.RequestContext) interface{} {
	var queries []gqlLog
	if ctx.Doc == nil {
		return queries
	}

	for _, op := range ctx.Doc.Operations {
		for _, sel := range op.SelectionSet {
			if field, ok := sel.(*ast.Field); ok {
				name := op.Name
				if name == "" {
					name = "Unnamed"
				}
				queries = append(queries, gqlLog{
					Op:    string(op.Operation),
					Name:  name,
					Field: field.Name,
					Vars:  ctx.Variables,
				})
			}
		}
	}
	return queries
}
