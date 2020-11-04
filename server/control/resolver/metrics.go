package resolver

import (
	"context"
	"fmt"
	"time"

	"gitlab.com/beneath-hq/beneath/infrastructure/engine/driver"

	uuid "github.com/satori/go.uuid"
	"github.com/vektah/gqlparser/v2/gqlerror"

	"gitlab.com/beneath-hq/beneath/pkg/mathutil"
	"gitlab.com/beneath-hq/beneath/server/control/gql"
	"gitlab.com/beneath-hq/beneath/services/middleware"
)

func (r *queryResolver) GetUsage(ctx context.Context, entityKind gql.EntityKind, entityID uuid.UUID, period string, from time.Time, until *time.Time) ([]*gql.Usage, error) {
	switch entityKind {
	case gql.EntityKindOrganization:
		return r.GetOrganizationUsage(ctx, entityID, period, from, until)
	case gql.EntityKindService:
		return r.GetServiceUsage(ctx, entityID, period, from, until)
	case gql.EntityKindStreamInstance:
		return r.GetStreamInstanceUsage(ctx, entityID, period, from, until)
	case gql.EntityKindStream:
		return r.GetStreamUsage(ctx, entityID, period, from, until)
	case gql.EntityKindUser:
		return r.GetUserUsage(ctx, entityID, period, from, until)
	}
	return nil, gqlerror.Errorf("Unrecognized entity kind %s", entityKind)
}

func (r *queryResolver) GetOrganizationUsage(ctx context.Context, organizationID uuid.UUID, period string, from time.Time, until *time.Time) ([]*gql.Usage, error) {
	secret := middleware.GetSecret(ctx)
	perms := r.Permissions.OrganizationPermissionsForSecret(ctx, secret, organizationID)
	if !perms.View {
		return nil, gqlerror.Errorf("you do not have permission to view this organization's usage")
	}

	return r.getUsage(ctx, organizationID, period, from, until)
}

func (r *queryResolver) GetServiceUsage(ctx context.Context, serviceID uuid.UUID, period string, from time.Time, until *time.Time) ([]*gql.Usage, error) {
	service := r.Services.FindService(ctx, serviceID)
	if service == nil {
		return nil, gqlerror.Errorf("service not found")
	}

	secret := middleware.GetSecret(ctx)
	perms := r.Permissions.ProjectPermissionsForSecret(ctx, secret, service.ProjectID, service.Project.Public)
	if !perms.View {
		return nil, gqlerror.Errorf("you do not have permission to view this service's usage")
	}

	return r.getUsage(ctx, serviceID, period, from, until)
}

func (r *queryResolver) GetStreamInstanceUsage(ctx context.Context, streamInstanceID uuid.UUID, period string, from time.Time, until *time.Time) ([]*gql.Usage, error) {
	stream := r.Streams.FindCachedInstance(ctx, streamInstanceID)
	if stream == nil {
		return nil, gqlerror.Errorf("Stream for instance %s not found", streamInstanceID.String())
	}

	secret := middleware.GetSecret(ctx)
	perms := r.Permissions.StreamPermissionsForSecret(ctx, secret, stream.StreamID, stream.ProjectID, stream.Public)
	if !perms.Read {
		return nil, gqlerror.Errorf("you do not have permission to view this stream's usage")
	}

	return r.getUsage(ctx, streamInstanceID, period, from, until)
}

func (r *queryResolver) GetStreamUsage(ctx context.Context, streamID uuid.UUID, period string, from time.Time, until *time.Time) ([]*gql.Usage, error) {
	stream := r.Streams.FindStream(ctx, streamID)
	if stream == nil {
		return nil, gqlerror.Errorf("Stream %s not found", streamID.String())
	}

	secret := middleware.GetSecret(ctx)
	perms := r.Permissions.StreamPermissionsForSecret(ctx, secret, streamID, stream.ProjectID, stream.Project.Public)
	if !perms.Read {
		return nil, gqlerror.Errorf("you do not have permission to view this stream's usage")
	}

	return r.getUsage(ctx, stream.StreamID, period, from, until)
}

func (r *queryResolver) GetUserUsage(ctx context.Context, userID uuid.UUID, period string, from time.Time, until *time.Time) ([]*gql.Usage, error) {
	secret := middleware.GetSecret(ctx)
	if secret.GetOwnerID() != userID {
		user := r.Users.FindUser(ctx, userID)
		if user == nil {
			return nil, gqlerror.Errorf("user not found")
		}

		perms := r.Permissions.OrganizationPermissionsForSecret(ctx, secret, user.BillingOrganizationID)
		if !perms.View {
			return nil, gqlerror.Errorf("you do not have permission to view this user's usage")
		}
	}

	return r.getUsage(ctx, userID, period, from, until)
}

func (r *queryResolver) getUsage(ctx context.Context, entityID uuid.UUID, period string, from time.Time, until *time.Time) ([]*gql.Usage, error) {
	// if until is not provided, set to the empty time (usage.GetUsage then defaults to current time)
	if until == nil {
		until = &time.Time{}
	}

	// parse label from period from string
	var label driver.UsageLabel
	if period == "month" {
		label = driver.UsageLabelMonthly
	} else if period == "hour" {
		label = driver.UsageLabelHourly
	} else if period == "quota_month" {
		label = driver.UsageLabelQuotaMonth
	} else {
		return nil, gqlerror.Errorf("unsupported usage period '%s'", period)
	}

	times, usages, err := r.Usage.GetHistoricalUsageRange(ctx, entityID, label, from, *until)
	if err != nil {
		return nil, gqlerror.Errorf("couldn't get usage: %v", err)
	}

	gqlUsages := make([]*gql.Usage, len(usages))
	for i, usage := range usages {
		gqlUsages[i] = &gql.Usage{
			EntityID:     entityID,
			Period:       period,
			Time:         times[i],
			ReadOps:      int(usage.ReadOps),
			ReadBytes:    int(usage.ReadBytes),
			ReadRecords:  int(usage.ReadRecords),
			WriteOps:     int(usage.WriteOps),
			WriteBytes:   int(usage.WriteBytes),
			WriteRecords: int(usage.WriteRecords),
			ScanOps:      int(usage.ScanOps),
			ScanBytes:    int(usage.ScanBytes),
		}
	}

	return gqlUsages, nil
}

func mergeUsage(xs []*gql.Usage, ys []*gql.Usage) []*gql.Usage {
	if len(xs) == 0 {
		return ys
	}
	if len(ys) == 0 {
		return xs
	}

	n := mathutil.MaxInt(len(xs), len(ys))
	zs := make([]*gql.Usage, 0, n)

	var i, j int
	for i < len(xs) && j < len(ys) {
		diff := xs[i].Time.Sub(ys[j].Time)
		if diff == 0 {
			zs = append(zs, addUsage(xs[i], ys[j]))
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

func addUsage(target, other *gql.Usage) *gql.Usage {
	target.ReadOps += other.ReadOps
	target.ReadBytes += other.ReadBytes
	target.ReadRecords += other.ReadRecords
	target.WriteOps += other.WriteOps
	target.WriteBytes += other.WriteBytes
	target.WriteRecords += other.WriteRecords
	target.ScanOps += other.ScanOps
	target.ScanBytes += other.ScanBytes
	return target
}
