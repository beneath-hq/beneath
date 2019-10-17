package resolver

import (
	"context"
	"time"

	"github.com/beneath-core/beneath-go/core/timeutil"

	uuid "github.com/satori/go.uuid"
	"github.com/vektah/gqlparser/gqlerror"

	"github.com/beneath-core/beneath-go/control/entity"
	"github.com/beneath-core/beneath-go/control/gql"
	"github.com/beneath-core/beneath-go/core/middleware"
	"github.com/beneath-core/beneath-go/metrics"
)

func (r *queryResolver) GetStreamMetrics(ctx context.Context, streamID uuid.UUID, period string, from time.Time, until *time.Time) ([]*gql.Metrics, error) {
	stream := entity.FindStream(ctx, streamID)
	if stream == nil {
		return nil, gqlerror.Errorf("Stream %s not found", streamID.String())
	}

	if !stream.Project.Public {
		secret := middleware.GetSecret(ctx)
		perms := secret.StreamPermissions(ctx, streamID, stream.ProjectID, stream.External)
		if !perms.Read {
			return nil, gqlerror.Errorf("you do not have permission to view this stream's metrics")
		}
	}

	// lookup stream ID for batch, instance ID for streaming
	if stream.Batch {
		return getUsage(ctx, stream.StreamID, period, from, until)
	} else if stream.CurrentStreamInstanceID != nil {
		return getUsage(ctx, *stream.CurrentStreamInstanceID, period, from, until)
	}

	return nil, nil
}

func (r *queryResolver) GetUserMetrics(ctx context.Context, userID uuid.UUID, period string, from time.Time, until *time.Time) ([]*gql.Metrics, error) {
	user := entity.FindUser(ctx, userID)
	if user == nil {
		return nil, gqlerror.Errorf("user not found")
	}

	secret := middleware.GetSecret(ctx)
	if !secret.IsUserID(userID) {
		perms := entity.OrganizationPermissions{}
		if user.MainOrganizationID != nil {
			perms = secret.OrganizationPermissions(ctx, *user.MainOrganizationID)
		}
		if !perms.View {
			return nil, gqlerror.Errorf("you do not have permission to view this stream's metrics")
		}
	}

	return getUsage(ctx, userID, period, from, until)
}

func (r *queryResolver) GetServiceMetrics(ctx context.Context, serviceID uuid.UUID, period string, from time.Time, until *time.Time) ([]*gql.Metrics, error) {
	service := entity.FindService(ctx, serviceID)
	if service == nil {
		return nil, gqlerror.Errorf("service not found")
	}

	secret := middleware.GetSecret(ctx)
	perms := secret.OrganizationPermissions(ctx, service.OrganizationID)
	if !perms.View {
		return nil, gqlerror.Errorf("you do not have permission to view this service's metrics")
	}

	return getUsage(ctx, serviceID, period, from, until)
}

func getUsage(ctx context.Context, entityID uuid.UUID, period string, from time.Time, until *time.Time) ([]*gql.Metrics, error) {
	// if until is not provided, set to the empty time (metrics.GetUsage then defaults to current time)
	if until == nil {
		until = &time.Time{}
	}

	// parse period from string
	p, err := timeutil.PeriodFromString(period)
	if err != nil {
		return nil, gqlerror.Errorf("%v", err.Error())
	}

	times, usages, err := metrics.GetHistoricalUsage(ctx, entityID, p, from, *until)
	if err != nil {
		return nil, gqlerror.Errorf("couldn't get usage: %v", err)
	}

	metrics := make([]*gql.Metrics, len(usages))
	for i, usage := range usages {
		metrics[i] = &gql.Metrics{
			EntityID:     entityID,
			Period:       period,
			Time:         times[i],
			ReadOps:      int(usage.ReadOps),
			ReadBytes:    int(usage.ReadBytes),
			ReadRecords:  int(usage.ReadRecords),
			WriteOps:     int(usage.WriteOps),
			WriteBytes:   int(usage.WriteBytes),
			WriteRecords: int(usage.WriteRecords),
		}
	}

	return metrics, nil
}
