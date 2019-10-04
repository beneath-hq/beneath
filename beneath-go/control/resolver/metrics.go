package resolver

import (
	"context"
	"time"

	"github.com/beneath-core/beneath-go/control/entity"
	"github.com/beneath-core/beneath-go/control/gql"
	"github.com/beneath-core/beneath-go/core/middleware"
	"github.com/beneath-core/beneath-go/metrics"
	uuid "github.com/satori/go.uuid"
	"github.com/vektah/gqlparser/gqlerror"
)

func (r *queryResolver) GetStreamMetrics(ctx context.Context, streamID uuid.UUID, period string, from time.Time, until *time.Time) ([]*gql.Metrics, error) {
	secret := middleware.GetSecret(ctx)
	stream := entity.FindStream(ctx, streamID)

	perms := secret.ProjectPermissions(ctx, stream.ProjectID) // does this return true for public projects?
	if !perms.View {
		return nil, gqlerror.Errorf("you do not have permission to view this stream's metrics")
	}

	// if until is not provided, set to 0 (metrics.GetUsage sets this to the current time)
	if until == nil {
		until := time.Time{}
	} else {
		until := &until
	}

	usagePackets := metrics.GetUsage(ctx, serviceID, period, from, until)
	metrics := make([]*gql.Metrics, len(usagePackets))

	// unpack usagePackets and craft output
	for i, usage := range usagePackets {
		metrics[i] = *gql.Metrics{
			EntityID:     streamID,
			Period:       period,
			Time:         from,
			ReadOps:      usage.ReadOps,
			ReadBytes:    usage.ReadBytes,
			ReadRecords:  usage.ReadRecords,
			WriteOps:     usage.WriteOps,
			WriteBytes:   usage.WriteBytes,
			WriteRecords: usage.WriteRecords,
		}
	}

	// done
	return metrics, nil
}

func (r *queryResolver) GetUserMetrics(ctx context.Context, userID uuid.UUID, period string, from time.Time, until *time.Time) ([]*gql.Metrics, error) {
	secret := middleware.GetSecret(ctx)
	user := entity.FindUser(ctx, userID)

	perms := secret.OrganizationPermissions(ctx, user.OrganizationID)
	if !perms.View {
		return nil, gqlerror.Errorf("you do not have permission to view this stream's metrics")
	}

	// if until is not provided, set to 0 (metrics.GetUsage sets this to the current time)
	if until == nil {
		until := time.Time{}
	} else {
		until := &until
	}

	metricsPackets := metrics.GetUsage(ctx, serviceID, period, from, until)
	metrics := make([]*gql.Metrics, len(metricsPackets))

	// unpack metricsPackets and craft output
	for i, usage := range metricsPackets {
		metrics[i] = *gql.Metrics{
			EntityID:     serviceID,
			Period:       period,
			Time:         from,
			ReadOps:      usage.ReadOps,
			ReadBytes:    usage.ReadBytes,
			ReadRecords:  usage.ReadRecords,
			WriteOps:     usage.WriteOps,
			WriteBytes:   usage.WriteBytes,
			WriteRecords: usage.WriteRecords,
		}
	}

	// done
	return metrics, nil
}

func (r *queryResolver) GetServiceMetrics(ctx context.Context, serviceID uuid.UUID, period string, from time.Time, until *time.Time) ([]*gql.Metrics, error) {
	secret := middleware.GetSecret(ctx)
	service := entity.FindService(ctx, serviceID)

	perms := secret.OrganizationPermissions(ctx, service.OrganizationID)
	if !perms.View {
		return nil, gqlerror.Errorf("you do not have permission to view this service's metrics")
	}

	// if until is not provided, set to 0 (metrics.GetUsage sets this to the current time)
	if until == nil {
		until := time.Time{}
	} else {
		until := &until
	}

	metricsPackets := metrics.GetUsage(ctx, serviceID, period, from, until)
	metrics := make([]*gql.Metrics, len(metricsPackets))

	// unpack metricsPackets and craft output
	for i, usage := range metricsPackets {
		metrics[i] = *gql.Metrics{
			EntityID:     serviceID,
			Period:       period,
			Time:         from,
			ReadOps:      usage.ReadOps,
			ReadBytes:    usage.ReadBytes,
			ReadRecords:  usage.ReadRecords,
			WriteOps:     usage.WriteOps,
			WriteBytes:   usage.WriteBytes,
			WriteRecords: usage.WriteRecords,
		}
	}

	// done
	return metrics, nil
}
