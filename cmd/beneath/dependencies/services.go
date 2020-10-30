package dependencies

import (
	"time"

	"gitlab.com/beneath-hq/beneath/cmd/beneath/cli"
	"gitlab.com/beneath-hq/beneath/services/data"
	"gitlab.com/beneath-hq/beneath/services/metrics"
	"gitlab.com/beneath-hq/beneath/services/middleware"
	"gitlab.com/beneath-hq/beneath/services/organization"
	"gitlab.com/beneath-hq/beneath/services/permissions"
	"gitlab.com/beneath-hq/beneath/services/project"
	"gitlab.com/beneath-hq/beneath/services/secret"
	"gitlab.com/beneath-hq/beneath/services/service"
	"gitlab.com/beneath-hq/beneath/services/stream"
	"gitlab.com/beneath-hq/beneath/services/user"
)

// TO ADD A NEW SERVICE:
// 1. Add it as a member to AllServices
// 2. Add it to NewAllServices (two places!)
// 3. Add it as a dependency in init()

// AllServices is a convenience wrapper that initializes all services
type AllServices struct {
	Data         *data.Service
	Metrics      *metrics.Service
	Middleware   *middleware.Service
	Organization *organization.Service
	Permissions  *permissions.Service
	Project      *project.Service
	Secret       *secret.Service
	Service      *service.Service
	Stream       *stream.Service
	User         *user.Service
}

// NewAllServices creates a new AllServices
func NewAllServices(
	data *data.Service,
	metrics *metrics.Service,
	middleware *middleware.Service,
	organization *organization.Service,
	permissions *permissions.Service,
	project *project.Service,
	secret *secret.Service,
	service *service.Service,
	stream *stream.Service,
	user *user.Service,
) *AllServices {
	return &AllServices{
		Data:         data,
		Metrics:      metrics,
		Middleware:   middleware,
		Organization: organization,
		Permissions:  permissions,
		Project:      project,
		Secret:       secret,
		Service:      service,
		Stream:       stream,
		User:         user,
	}
}

func init() {
	cli.AddDependency(NewAllServices)
	cli.AddDependency(data.New)
	cli.AddDependency(metrics.New)
	cli.AddDependency(middleware.New)
	cli.AddDependency(organization.New)
	cli.AddDependency(permissions.New)
	cli.AddDependency(project.New)
	cli.AddDependency(secret.New)
	cli.AddDependency(service.New)
	cli.AddDependency(stream.New)
	cli.AddDependency(user.New)

	// the metrics service takes some extra options
	cli.AddDependency(func() *metrics.Options {
		return &metrics.Options{
			CacheSize:      2500,
			CommitInterval: 30 * time.Second,
		}
	})
}
