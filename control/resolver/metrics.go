package resolver

import (
	"context"
	"fmt"
	"time"

	uuid "github.com/satori/go.uuid"
	"github.com/vektah/gqlparser/gqlerror"

	"github.com/beneath-core/control/entity"
	"github.com/beneath-core/control/gql"
	"github.com/beneath-core/pkg/mathutil"
	"github.com/beneath-core/internal/middleware"
	"github.com/beneath-core/pkg/timeutil"
	"github.com/beneath-core/metrics"
)

func (r *queryResolver) GetStreamMetrics(ctx context.Context, streamID uuid.UUID, period string, from time.Time, until *time.Time) ([]*gql.Metrics, error) {
	stream := entity.FindStream(ctx, streamID)
	if stream == nil {
		return nil, gqlerror.Errorf("Stream %s not found", streamID.String())
	}

	secret := middleware.GetSecret(ctx)
	perms := secret.StreamPermissions(ctx, streamID, stream.ProjectID, stream.Project.Public, stream.External)
	if !perms.Read {
		return nil, gqlerror.Errorf("you do not have permission to view this stream's metrics")
	}

	// lookup stream ID for batch, instance ID for streaming
	if stream.Batch {
		reads, err := getUsage(ctx, stream.StreamID, period, from, until)
		if err != nil {
			return nil, err
		}

		if stream.CurrentStreamInstanceID == nil {
			return reads, nil
		}

		writes, err := getUsage(ctx, *stream.CurrentStreamInstanceID, period, from, until)
		if err != nil {
			return nil, err
		}

		return mergeUsage(reads, writes), nil
	}

	if stream.CurrentStreamInstanceID == nil {
		return nil, nil
	}

	return getUsage(ctx, *stream.CurrentStreamInstanceID, period, from, until)
}

func (r *queryResolver) GetUserMetrics(ctx context.Context, userID uuid.UUID, period string, from time.Time, until *time.Time) ([]*gql.Metrics, error) {
	user := entity.FindUser(ctx, userID)
	if user == nil {
		return nil, gqlerror.Errorf("user not found")
	}

	secret := middleware.GetSecret(ctx)
	if secret.GetOwnerID() != userID {
		perms := secret.OrganizationPermissions(ctx, user.OrganizationID)
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

func mergeUsage(xs []*gql.Metrics, ys []*gql.Metrics) []*gql.Metrics {
	if len(xs) == 0 {
		return ys
	}
	if len(ys) == 0 {
		return xs
	}

	n := mathutil.MaxInt(len(xs), len(ys))
	zs := make([]*gql.Metrics, 0, n)

	var i, j int
	for i < len(xs) && j < len(ys) {
		diff := xs[i].Time.Sub(ys[j].Time)
		if diff == 0 {
			zs = append(zs, addMetrics(xs[i], ys[j]))
			i++
			j++
		} else if diff < 0 {
			zs = append(zs, xs[i])
			i++
		} else if diff > 0 {
			zs = append(zs, ys[j])
			j++
		} else {
			panic(fmt.Errorf("impossible state in mergeUsage"))
		}
	}

	for i < len(xs) {
		zs = append(zs, xs[i])
		i++
	}

	for j < len(ys) {
		zs = append(zs, ys[j])
		j++
	}

	return zs
}

func addMetrics(target, other *gql.Metrics) *gql.Metrics {
	target.ReadOps += other.ReadOps
	target.ReadBytes += other.ReadBytes
	target.ReadRecords += other.ReadRecords
	target.WriteOps += other.WriteOps
	target.WriteBytes += other.WriteBytes
	target.WriteRecords += other.WriteRecords
	return target
}
