package middleware

import (
	"context"
	"net/http"

	"github.com/99designs/gqlgen/graphql"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.uber.org/zap"

	"gitlab.com/beneath-hq/beneath/pkg/httputil"
	"gitlab.com/beneath-hq/beneath/pkg/log"
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
func (s *Service) DefaultGQLRecoverFunc(ctx context.Context, errT interface{}) error {
	if err, ok := errT.(error); ok {
		log.L.Error("http recovered panic", zap.Error(err))
	} else {
		log.L.Error("http recovered panic", zap.Reflect("error", errT))
	}
	return httputil.NewError(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
}

// QueryLoggingGQLMiddleware logs sensible info about every GraphQL request
func (s *Service) QueryLoggingGQLMiddleware(ctx context.Context, next graphql.ResponseHandler) *graphql.Response {
	reqCtx := graphql.GetRequestContext(ctx)
	SetTagsPayload(ctx, s.logInfoFromGQLRequestContext(reqCtx))
	return next(ctx)
}

func (s *Service) logInfoFromGQLRequestContext(ctx *graphql.RequestContext) interface{} {
	var queries []gqlLog
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
