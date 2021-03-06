// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package gql

import (
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/beneath-hq/beneath/models"
	"github.com/satori/go.uuid"
)

type Organization interface {
	IsOrganization()
}

type CompileSchemaInput struct {
	SchemaKind models.TableSchemaKind `json:"schemaKind"`
	Schema     string                 `json:"schema"`
	Indexes    *string                `json:"indexes"`
}

type CompileSchemaOutput struct {
	CanonicalAvroSchema string `json:"canonicalAvroSchema"`
	CanonicalIndexes    string `json:"canonicalIndexes"`
}

type CreateAuthTicketInput struct {
	RequesterName string `json:"requesterName"`
}

type CreateProjectInput struct {
	OrganizationID uuid.UUID `json:"organizationID"`
	ProjectName    string    `json:"projectName"`
	DisplayName    *string   `json:"displayName"`
	Public         *bool     `json:"public"`
	Description    *string   `json:"description"`
	Site           *string   `json:"site"`
	PhotoURL       *string   `json:"photoURL"`
}

type CreateServiceInput struct {
	OrganizationName string  `json:"organizationName"`
	ProjectName      string  `json:"projectName"`
	ServiceName      string  `json:"serviceName"`
	Description      *string `json:"description"`
	SourceURL        *string `json:"sourceURL"`
	ReadQuota        *int    `json:"readQuota"`
	WriteQuota       *int    `json:"writeQuota"`
	ScanQuota        *int    `json:"scanQuota"`
	UpdateIfExists   *bool   `json:"updateIfExists"`
}

type CreateTableInput struct {
	OrganizationName          string                 `json:"organizationName"`
	ProjectName               string                 `json:"projectName"`
	TableName                 string                 `json:"tableName"`
	SchemaKind                models.TableSchemaKind `json:"schemaKind"`
	Schema                    string                 `json:"schema"`
	Indexes                   *string                `json:"indexes"`
	Description               *string                `json:"description"`
	Meta                      *bool                  `json:"meta"`
	AllowManualWrites         *bool                  `json:"allowManualWrites"`
	UseLog                    *bool                  `json:"useLog"`
	UseIndex                  *bool                  `json:"useIndex"`
	UseWarehouse              *bool                  `json:"useWarehouse"`
	LogRetentionSeconds       *int                   `json:"logRetentionSeconds"`
	IndexRetentionSeconds     *int                   `json:"indexRetentionSeconds"`
	WarehouseRetentionSeconds *int                   `json:"warehouseRetentionSeconds"`
	UpdateIfExists            *bool                  `json:"updateIfExists"`
}

type CreateTableInstanceInput struct {
	TableID        uuid.UUID `json:"tableID"`
	Version        *int      `json:"version"`
	MakePrimary    *bool     `json:"makePrimary"`
	UpdateIfExists *bool     `json:"updateIfExists"`
}

type DeleteProjectInput struct {
	ProjectID uuid.UUID `json:"projectID"`
}

type GetEntityUsageInput struct {
	EntityID uuid.UUID  `json:"entityID"`
	Label    UsageLabel `json:"label"`
	From     *time.Time `json:"from"`
	Until    *time.Time `json:"until"`
}

type GetUsageInput struct {
	EntityKind EntityKind `json:"entityKind"`
	EntityID   uuid.UUID  `json:"entityID"`
	Label      UsageLabel `json:"label"`
	From       *time.Time `json:"from"`
	Until      *time.Time `json:"until"`
}

type NewServiceSecret struct {
	Secret *models.ServiceSecret `json:"secret"`
	Token  string                `json:"token"`
}

type NewUserSecret struct {
	Secret *models.UserSecret `json:"secret"`
	Token  string             `json:"token"`
}

type PrivateOrganization struct {
	OrganizationID    string                                `json:"organizationID"`
	Name              string                                `json:"name"`
	DisplayName       string                                `json:"displayName"`
	Description       *string                               `json:"description"`
	PhotoURL          *string                               `json:"photoURL"`
	CreatedOn         time.Time                             `json:"createdOn"`
	UpdatedOn         time.Time                             `json:"updatedOn"`
	QuotaEpoch        time.Time                             `json:"quotaEpoch"`
	QuotaStartTime    time.Time                             `json:"quotaStartTime"`
	QuotaEndTime      time.Time                             `json:"quotaEndTime"`
	ReadQuota         *int                                  `json:"readQuota"`
	WriteQuota        *int                                  `json:"writeQuota"`
	ScanQuota         *int                                  `json:"scanQuota"`
	PrepaidReadQuota  *int                                  `json:"prepaidReadQuota"`
	PrepaidWriteQuota *int                                  `json:"prepaidWriteQuota"`
	PrepaidScanQuota  *int                                  `json:"prepaidScanQuota"`
	ReadUsage         int                                   `json:"readUsage"`
	WriteUsage        int                                   `json:"writeUsage"`
	ScanUsage         int                                   `json:"scanUsage"`
	Projects          []*models.Project                     `json:"projects"`
	PersonalUserID    *uuid.UUID                            `json:"personalUserID"`
	PersonalUser      *models.User                          `json:"personalUser"`
	Permissions       *models.PermissionsUsersOrganizations `json:"permissions"`
}

func (PrivateOrganization) IsOrganization() {}

type UpdateAuthTicketInput struct {
	AuthTicketID uuid.UUID `json:"authTicketID"`
	Approve      bool      `json:"approve"`
}

type UpdateProjectInput struct {
	ProjectID   uuid.UUID `json:"projectID"`
	DisplayName *string   `json:"displayName"`
	Public      *bool     `json:"public"`
	Description *string   `json:"description"`
	Site        *string   `json:"site"`
	PhotoURL    *string   `json:"photoURL"`
}

type UpdateServiceInput struct {
	OrganizationName string  `json:"organizationName"`
	ProjectName      string  `json:"projectName"`
	ServiceName      string  `json:"serviceName"`
	Description      *string `json:"description"`
	SourceURL        *string `json:"sourceURL"`
	ReadQuota        *int    `json:"readQuota"`
	WriteQuota       *int    `json:"writeQuota"`
	ScanQuota        *int    `json:"scanQuota"`
}

type UpdateTableInput struct {
	TableID           uuid.UUID               `json:"tableID"`
	SchemaKind        *models.TableSchemaKind `json:"schemaKind"`
	Schema            *string                 `json:"schema"`
	Indexes           *string                 `json:"indexes"`
	Description       *string                 `json:"description"`
	AllowManualWrites *bool                   `json:"allowManualWrites"`
}

type UpdateTableInstanceInput struct {
	TableInstanceID uuid.UUID `json:"tableInstanceID"`
	MakeFinal       *bool     `json:"makeFinal"`
	MakePrimary     *bool     `json:"makePrimary"`
}

type Usage struct {
	EntityID     uuid.UUID  `json:"entityID"`
	Label        UsageLabel `json:"label"`
	Time         time.Time  `json:"time"`
	ReadOps      int        `json:"readOps"`
	ReadBytes    int        `json:"readBytes"`
	ReadRecords  int        `json:"readRecords"`
	WriteOps     int        `json:"writeOps"`
	WriteBytes   int        `json:"writeBytes"`
	WriteRecords int        `json:"writeRecords"`
	ScanOps      int        `json:"scanOps"`
	ScanBytes    int        `json:"scanBytes"`
}

type EntityKind string

const (
	EntityKindOrganization  EntityKind = "Organization"
	EntityKindService       EntityKind = "Service"
	EntityKindTableInstance EntityKind = "TableInstance"
	EntityKindTable         EntityKind = "Table"
	EntityKindUser          EntityKind = "User"
)

var AllEntityKind = []EntityKind{
	EntityKindOrganization,
	EntityKindService,
	EntityKindTableInstance,
	EntityKindTable,
	EntityKindUser,
}

func (e EntityKind) IsValid() bool {
	switch e {
	case EntityKindOrganization, EntityKindService, EntityKindTableInstance, EntityKindTable, EntityKindUser:
		return true
	}
	return false
}

func (e EntityKind) String() string {
	return string(e)
}

func (e *EntityKind) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = EntityKind(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid EntityKind", str)
	}
	return nil
}

func (e EntityKind) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

type UsageLabel string

const (
	UsageLabelTotal      UsageLabel = "Total"
	UsageLabelQuotaMonth UsageLabel = "QuotaMonth"
	UsageLabelMonthly    UsageLabel = "Monthly"
	UsageLabelHourly     UsageLabel = "Hourly"
)

var AllUsageLabel = []UsageLabel{
	UsageLabelTotal,
	UsageLabelQuotaMonth,
	UsageLabelMonthly,
	UsageLabelHourly,
}

func (e UsageLabel) IsValid() bool {
	switch e {
	case UsageLabelTotal, UsageLabelQuotaMonth, UsageLabelMonthly, UsageLabelHourly:
		return true
	}
	return false
}

func (e UsageLabel) String() string {
	return string(e)
}

func (e *UsageLabel) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = UsageLabel(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid UsageLabel", str)
	}
	return nil
}

func (e UsageLabel) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}
