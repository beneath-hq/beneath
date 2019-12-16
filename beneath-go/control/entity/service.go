package entity

import (
	"context"
	"regexp"
	"time"

	"github.com/beneath-core/beneath-go/core/timeutil"
	"github.com/beneath-core/beneath-go/db"
	"github.com/beneath-core/beneath-go/metrics"
	uuid "github.com/satori/go.uuid"
	"gopkg.in/go-playground/validator.v9"
)

// Service represents external service keys, models, and, in the future, charts, all of which need OrganizationIDs for billing
type Service struct {
	ServiceID      uuid.UUID   `sql:",pk,type:uuid,default:uuid_generate_v4()"`
	Name           string      `sql:",notnull",validate:"required,gte=1,lte=40"` // not unique because of (organization_id, service_id) index
	Kind           ServiceKind `sql:",notnull"`
	OrganizationID uuid.UUID   `sql:"on_delete:cascade,notnull,type:uuid"`
	Organization   *Organization
	ReadQuota      int64
	WriteQuota     int64
	CreatedOn      time.Time `sql:",notnull,default:now()"`
	UpdatedOn      time.Time `sql:",notnull,default:now()"`
	Secrets        []*ServiceSecret
}

// ServiceKind represents a external, model, etc. services
type ServiceKind string

const (
	// ServiceKindExternal is a service key that people can put on external facing products
	ServiceKindExternal ServiceKind = "external"

	// ServiceKindModel is a service key for models
	ServiceKindModel ServiceKind = "model"
)

var (
	// used for validation
	serviceNameRegex *regexp.Regexp
)

func init() {
	// configure validation
	serviceNameRegex = regexp.MustCompile("^[_a-z][_a-z0-9]*$")
	GetValidator().RegisterStructValidation(serviceValidation, Service{})
}

// custom service validation
func serviceValidation(sl validator.StructLevel) {
	s := sl.Current().Interface().(Service)

	if !serviceNameRegex.MatchString(s.Name) {
		sl.ReportError(s.Name, "Name", "", "alphanumericorunderscore", "")
	}
}

// FindService returns the matching service or nil
func FindService(ctx context.Context, serviceID uuid.UUID) *Service {
	service := &Service{
		ServiceID: serviceID,
	}
	err := db.DB.ModelContext(ctx, service).WherePK().Select()
	if !AssertFoundOne(err) {
		return nil
	}

	return service
}

// FindServiceByNameAndOrganization returns the matching service or nil
func FindServiceByNameAndOrganization(ctx context.Context, name, organizationName string) *Service {
	service := &Service{}
	err := db.DB.ModelContext(ctx, service).
		Column("service.*", "Organization").
		Where("lower(service.name) = lower(?)", name).
		Where("lower(organization.name) = lower(?)", organizationName).
		Select()
	if !AssertFoundOne(err) {
		return nil
	}
	return service
}

// CreateService consolidates and returns the service matching the args
func CreateService(ctx context.Context, name string, kind ServiceKind, organizationID uuid.UUID, readQuota int, writeQuota int) (*Service, error) {
	s := &Service{}

	// set service fields
	s.Name = name
	s.Kind = kind
	s.OrganizationID = organizationID
	s.ReadQuota = int64(readQuota)
	s.WriteQuota = int64(writeQuota)

	// validate
	err := GetValidator().Struct(s)
	if err != nil {
		return nil, err
	}

	// insert
	err = db.DB.Insert(s)
	if err != nil {
		return nil, err
	}

	return s, nil
}

// UpdateDetails consolidates and returns the service matching the args
func (s *Service) UpdateDetails(ctx context.Context, name *string, readQuota *int, writeQuota *int) (*Service, error) {
	// set fields
	if name != nil {
		s.Name = *name
	}
	if readQuota != nil {
		s.ReadQuota = int64(*readQuota)
	}
	if writeQuota != nil {
		s.WriteQuota = int64(*writeQuota)
	}

	// validate
	err := GetValidator().Struct(s)
	if err != nil {
		return nil, err
	}

	// update
	s.UpdatedOn = time.Now()
	_, err = db.DB.ModelContext(ctx, s).Column("name", "read_quota", "write_quota", "updated_on").WherePK().Update()
	return s, err
}

// Delete removes a service from the database
func (s *Service) Delete(ctx context.Context) error {
	isMidPeriod := true
	err := commitUsageToBill(ctx, s.OrganizationID, ServiceEntityKind, s.ServiceID, isMidPeriod)
	if err != nil {
		return err
	}
	_, err = db.DB.ModelContext(ctx, s).WherePK().Delete()
	return err
}

// UpdatePermissions updates a service's permissions for a given stream
// UpdatePermissions sets permissions if they do not exist yet
func (s *Service) UpdatePermissions(ctx context.Context, streamID uuid.UUID, read *bool, write *bool) (*PermissionsServicesStreams, error) {
	// create perm
	pss := &PermissionsServicesStreams{
		ServiceID: s.ServiceID,
		StreamID:  streamID,
	}
	if read != nil {
		pss.Read = *read
	}
	if write != nil {
		pss.Write = *write
	}

	// if neither read nor write, delete permission (if exists) -- else update
	if !pss.Read && !pss.Write {
		_, err := db.DB.ModelContext(ctx, pss).WherePK().Delete()
		if err != nil {
			return nil, err
		}
	} else {
		// build upsert
		q := db.DB.ModelContext(ctx, pss).OnConflict("(service_id, stream_id) DO UPDATE")
		if read != nil {
			q = q.Set("read = EXCLUDED.read")
		}
		if write != nil {
			q = q.Set("write = EXCLUDED.write")
		}

		// run upsert
		_, err := q.Insert()
		if err != nil {
			return nil, err
		}
	}

	return pss, nil
}

func commitUsageToBill(ctx context.Context, organizationID uuid.UUID, entityKind Kind, entityID uuid.UUID, isMidPeriod bool) error {
	billingInfo := FindBillingInfo(ctx, organizationID)
	if billingInfo == nil {
		panic("organization not found")
	}

	billTimes := calculateBillTimes(time.Now(), billingInfo.BillingPlan.Period, isMidPeriod)

	var billedResources []*BilledResource

	_, monthlyMetrics, err := metrics.GetHistoricalUsage(ctx, entityID, billingInfo.BillingPlan.Period, billTimes.StartTime, billTimes.BillingTime) // when adding annual plans, remember this function only accepts hourly or monthly periods
	if err != nil {
		return err
	}

	if len(monthlyMetrics) == 1 {
		// add reads
		billedResources = append(billedResources, &BilledResource{
			OrganizationID:  organizationID,
			BillingTime:     billTimes.BillingTime,
			EntityID:        entityID,
			EntityKind:      entityKind,
			StartTime:       billTimes.StartTime,
			EndTime:         billTimes.EndTime,
			Product:         ReadProduct,
			Quantity:        monthlyMetrics[0].ReadBytes,
			TotalPriceCents: 0,
			Currency:        billingInfo.BillingPlan.Currency,
		})

		// add writes
		billedResources = append(billedResources, &BilledResource{
			OrganizationID:  organizationID,
			BillingTime:     billTimes.BillingTime,
			EntityID:        entityID,
			EntityKind:      entityKind,
			StartTime:       billTimes.StartTime,
			EndTime:         billTimes.EndTime,
			Product:         WriteProduct,
			Quantity:        monthlyMetrics[0].WriteBytes,
			TotalPriceCents: 0,
			Currency:        billingInfo.BillingPlan.Currency,
		})
	} else if len(monthlyMetrics) > 1 {
		panic("expected a maximum of one item in monthlyMetrics")
	}

	if len(billedResources) > 0 {
		err = CreateOrUpdateBilledResources(ctx, billedResources)
		if err != nil {
			panic("unable to write billed resources to table")
		}
	}

	// done
	return nil
}

type billTimes struct {
	BillingTime time.Time
	StartTime   time.Time
	EndTime     time.Time
}

func calculateBillTimes(ts time.Time, p timeutil.Period, isMidPeriod bool) *billTimes {
	now := time.Now()

	var billingTime time.Time
	var startTime time.Time
	var endTime time.Time

	if isMidPeriod {
		billingTime = BeginningOfNextPeriod(p)
		startTime = BeginningOfThisPeriod(p)
		endTime = now
	} else {
		billingTime = BeginningOfThisPeriod(p)
		startTime = BeginningOfLastPeriod(p)
		endTime = EndOfLastPeriod(p)
		// FOR TESTING (since I don't have usage data from last month):
		startTime = startTime.AddDate(0, 1, 0)
		endTime = endTime.AddDate(0, 1, 0)
		billingTime = billingTime.AddDate(0, 1, 0)
	}
	return &billTimes{
		BillingTime: billingTime,
		StartTime:   startTime,
		EndTime:     endTime,
	}
}

// BeginningOfThisPeriod gets the beginning of this period
func BeginningOfThisPeriod(p timeutil.Period) time.Time {
	return timeutil.Floor(time.Now(), p)
}

// BeginningOfNextPeriod gets the beginning of the next period
func BeginningOfNextPeriod(p timeutil.Period) time.Time {
	ts := time.Now().UTC()
	return timeutil.Next(ts, p)
}

// BeginningOfLastPeriod gets the beginning of the last period
func BeginningOfLastPeriod(p timeutil.Period) time.Time {
	ts := time.Now().UTC()
	return timeutil.Last(ts, p)
}

// EndOfLastPeriod gets the end of the last period
func EndOfLastPeriod(p timeutil.Period) time.Time {
	ts := time.Now().UTC()
	return timeutil.Floor(ts, p)
}
