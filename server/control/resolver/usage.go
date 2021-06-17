package resolver

import (
	"context"
	"time"

	uuid "github.com/satori/go.uuid"
	"github.com/vektah/gqlparser/v2/gqlerror"

	"github.com/beneath-hq/beneath/infra/engine/driver"
	"github.com/beneath-hq/beneath/server/control/gql"
	"github.com/beneath-hq/beneath/services/middleware"
)

func (r *queryResolver) GetUsage(ctx context.Context, input gql.GetUsageInput) ([]*gql.Usage, error) {
	entityInput := gql.GetEntityUsageInput{
		EntityID: input.EntityID,
		Label:    input.Label,
		From:     input.From,
		Until:    input.Until,
	}
	switch input.EntityKind {
	case gql.EntityKindOrganization:
		return r.GetOrganizationUsage(ctx, entityInput)
	case gql.EntityKindService:
		return r.GetServiceUsage(ctx, entityInput)
	case gql.EntityKindTableInstance:
		return r.GetTableInstanceUsage(ctx, entityInput)
	case gql.EntityKindTable:
		return r.GetTableUsage(ctx, entityInput)
	case gql.EntityKindUser:
		return r.GetUserUsage(ctx, entityInput)
	}
	return nil, gqlerror.Errorf("Unrecognized entity kind %s", input.EntityKind)
}

func (r *queryResolver) GetOrganizationUsage(ctx context.Context, input gql.GetEntityUsageInput) ([]*gql.Usage, error) {
	secret := middleware.GetSecret(ctx)
	perms := r.Permissions.OrganizationPermissionsForSecret(ctx, secret, input.EntityID)
	if !perms.View {
		return nil, gqlerror.Errorf("you do not have permission to view this organization's usage")
	}

	return r.getUsage(ctx, input.EntityID, input.Label, input.From, input.Until)
}

func (r *queryResolver) GetServiceUsage(ctx context.Context, input gql.GetEntityUsageInput) ([]*gql.Usage, error) {
	service := r.Services.FindService(ctx, input.EntityID)
	if service == nil {
		return nil, gqlerror.Errorf("service not found")
	}

	secret := middleware.GetSecret(ctx)
	perms := r.Permissions.ProjectPermissionsForSecret(ctx, secret, service.ProjectID, service.Project.Public)
	if !perms.View {
		return nil, gqlerror.Errorf("you do not have permission to view this service's usage")
	}

	return r.getUsage(ctx, input.EntityID, input.Label, input.From, input.Until)
}

func (r *queryResolver) GetTableInstanceUsage(ctx context.Context, input gql.GetEntityUsageInput) ([]*gql.Usage, error) {
	table := r.Tables.FindCachedInstance(ctx, input.EntityID)
	if table == nil {
		return nil, gqlerror.Errorf("Table for instance %s not found", input.EntityID.String())
	}

	secret := middleware.GetSecret(ctx)
	perms := r.Permissions.TablePermissionsForSecret(ctx, secret, table.TableID, table.ProjectID, table.Public)
	if !perms.Read {
		return nil, gqlerror.Errorf("you do not have permission to view this table's usage")
	}

	return r.getUsage(ctx, input.EntityID, input.Label, input.From, input.Until)
}

func (r *queryResolver) GetTableUsage(ctx context.Context, input gql.GetEntityUsageInput) ([]*gql.Usage, error) {
	table := r.Tables.FindTable(ctx, input.EntityID)
	if table == nil {
		return nil, gqlerror.Errorf("Table %s not found", input.EntityID.String())
	}

	secret := middleware.GetSecret(ctx)
	perms := r.Permissions.TablePermissionsForSecret(ctx, secret, table.TableID, table.ProjectID, table.Project.Public)
	if !perms.Read {
		return nil, gqlerror.Errorf("you do not have permission to view this table's usage")
	}

	return r.getUsage(ctx, input.EntityID, input.Label, input.From, input.Until)
}

func (r *queryResolver) GetUserUsage(ctx context.Context, input gql.GetEntityUsageInput) ([]*gql.Usage, error) {
	secret := middleware.GetSecret(ctx)
	if secret.GetOwnerID() != input.EntityID {
		user := r.Users.FindUser(ctx, input.EntityID)
		if user == nil {
			return nil, gqlerror.Errorf("user not found")
		}

		perms := r.Permissions.OrganizationPermissionsForSecret(ctx, secret, user.BillingOrganizationID)
		if !perms.Admin {
			return nil, gqlerror.Errorf("you do not have permission to view this user's usage")
		}
	}

	return r.getUsage(ctx, input.EntityID, input.Label, input.From, input.Until)
}

func (r *queryResolver) getUsage(ctx context.Context, entityID uuid.UUID, label gql.UsageLabel, from *time.Time, until *time.Time) ([]*gql.Usage, error) {
	// parse label from period from string
	var driverLabel driver.UsageLabel
	switch label {
	case gql.UsageLabelTotal:
		driverLabel = driver.UsageLabelMonthly
	case gql.UsageLabelQuotaMonth:
		driverLabel = driver.UsageLabelQuotaMonth
	case gql.UsageLabelMonthly:
		driverLabel = driver.UsageLabelMonthly
	case gql.UsageLabelHourly:
		driverLabel = driver.UsageLabelHourly
	}

	times, usages, err := r.Usage.GetHistoricalUsageRange(ctx, entityID, driverLabel, from, until)
	if err != nil {
		return nil, gqlerror.Errorf("couldn't get usage: %v", err)
	}

	// If label is gql.UsageLabelTotal, sum up usages,
	// else, map to gql.Usage.
	var gqlUsages []*gql.Usage
	if label == gql.UsageLabelTotal {
		res := &gql.Usage{
			EntityID: entityID,
			Label:    label,
			Time:     time.Now(),
		}
		if len(times) > 0 {
			res.Time = times[0]
		}
		for _, usage := range usages {
			res.ReadOps += int(usage.ReadOps)
			res.ReadBytes += int(usage.ReadBytes)
			res.ReadRecords += int(usage.ReadRecords)
			res.WriteOps += int(usage.WriteOps)
			res.WriteBytes += int(usage.WriteBytes)
			res.WriteRecords += int(usage.WriteRecords)
			res.ScanOps += int(usage.ScanOps)
			res.ScanBytes += int(usage.ScanBytes)
		}
		gqlUsages = []*gql.Usage{res}
	} else {
		gqlUsages = make([]*gql.Usage, len(usages))
		for i, usage := range usages {
			gqlUsages[i] = &gql.Usage{
				EntityID:     entityID,
				Label:        label,
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
	}

	return gqlUsages, nil
}
