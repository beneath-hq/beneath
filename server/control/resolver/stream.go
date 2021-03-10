package resolver

import (
	"context"

	uuid "github.com/satori/go.uuid"
	"github.com/vektah/gqlparser/v2/gqlerror"

	"gitlab.com/beneath-hq/beneath/models"
	"gitlab.com/beneath-hq/beneath/server/control/gql"
	"gitlab.com/beneath-hq/beneath/services/middleware"
)

// Stream returns the gql.StreamResolver
func (r *Resolver) Stream() gql.StreamResolver {
	return &streamResolver{r}
}

type streamResolver struct{ *Resolver }

func (r *streamResolver) StreamID(ctx context.Context, obj *models.Stream) (string, error) {
	return obj.StreamID.String(), nil
}

func (r *queryResolver) StreamByID(ctx context.Context, streamID uuid.UUID) (*models.Stream, error) {
	stream := r.Streams.FindStream(ctx, streamID)
	if stream == nil {
		return nil, gqlerror.Errorf("Stream with ID %s not found", streamID.String())
	}

	secret := middleware.GetSecret(ctx)
	perms := r.Permissions.StreamPermissionsForSecret(ctx, secret, stream.StreamID, stream.ProjectID, stream.Project.Public)
	if !perms.Read {
		return nil, gqlerror.Errorf("Not allowed to read stream with ID %s", streamID.String())
	}

	return streamWithPermissions(stream, perms), nil
}

func (r *queryResolver) StreamByOrganizationProjectAndName(ctx context.Context, organizationName string, projectName string, streamName string) (*models.Stream, error) {
	stream := r.Streams.FindStreamByOrganizationProjectAndName(ctx, organizationName, projectName, streamName)
	if stream == nil {
		return nil, gqlerror.Errorf("Stream %s/%s/%s not found", organizationName, projectName, streamName)
	}

	secret := middleware.GetSecret(ctx)
	perms := r.Permissions.StreamPermissionsForSecret(ctx, secret, stream.StreamID, stream.ProjectID, stream.Project.Public)
	if !perms.Read && !perms.Write {
		return nil, gqlerror.Errorf("Not allowed to find stream %s/%s/%s", organizationName, projectName, streamName)
	}

	return streamWithPermissions(stream, perms), nil
}

func (r *queryResolver) StreamInstanceByStreamAndVersion(ctx context.Context, streamID uuid.UUID, version int) (*models.StreamInstance, error) {
	instance := r.Streams.FindStreamInstanceByVersion(ctx, streamID, version)
	if instance == nil {
		return nil, gqlerror.Errorf("Instance for stream %s version %d not found", streamID.String(), version)
	}

	secret := middleware.GetSecret(ctx)
	perms := r.Permissions.StreamPermissionsForSecret(ctx, secret, streamID, instance.Stream.ProjectID, instance.Stream.Project.Public)
	if !perms.Read {
		return nil, gqlerror.Errorf("Not allowed to read instance version %d for stream %s", version, streamID.String())
	}

	return instanceWithPermissions(instance, perms), nil
}

func (r *queryResolver) StreamInstanceByOrganizationProjectStreamAndVersion(ctx context.Context, organizationName string, projectName string, streamName string, version int) (*models.StreamInstance, error) {
	instance := r.Streams.FindStreamInstanceByOrganizationProjectStreamAndVersion(ctx, organizationName, projectName, streamName, version)
	if instance == nil {
		return nil, gqlerror.Errorf("Stream instance %s/%s/%s version %d not found", organizationName, projectName, streamName, version)
	}

	secret := middleware.GetSecret(ctx)
	perms := r.Permissions.StreamPermissionsForSecret(ctx, secret, instance.Stream.StreamID, instance.Stream.ProjectID, instance.Stream.Project.Public)
	if !perms.Read {
		return nil, gqlerror.Errorf("Not allowed to read stream %s/%s/%s", organizationName, projectName, streamName)
	}

	return instanceWithPermissions(instance, perms), nil
}

func (r *queryResolver) StreamInstancesForStream(ctx context.Context, streamID uuid.UUID) ([]*models.StreamInstance, error) {
	stream := r.Streams.FindStream(ctx, streamID)
	if stream == nil {
		return nil, gqlerror.Errorf("Stream with ID %s not found", streamID.String())
	}

	secret := middleware.GetSecret(ctx)
	perms := r.Permissions.StreamPermissionsForSecret(ctx, secret, stream.StreamID, stream.ProjectID, stream.Project.Public)
	if !perms.Read {
		return nil, gqlerror.Errorf("Not allowed to read stream with ID %s", streamID.String())
	}

	instances := r.Streams.FindStreamInstances(ctx, streamID, nil, nil)
	return instances, nil
}

func (r *queryResolver) StreamInstancesByOrganizationProjectAndStreamName(ctx context.Context, organizationName string, projectName string, streamName string) ([]*models.StreamInstance, error) {
	stream := r.Streams.FindStreamByOrganizationProjectAndName(ctx, organizationName, projectName, streamName)
	if stream == nil {
		return nil, gqlerror.Errorf("Stream %s/%s/%s not found", organizationName, projectName, streamName)
	}

	secret := middleware.GetSecret(ctx)
	perms := r.Permissions.StreamPermissionsForSecret(ctx, secret, stream.StreamID, stream.ProjectID, stream.Project.Public)
	if !perms.Read {
		return nil, gqlerror.Errorf("Not allowed to read stream with ID %s", stream.StreamID.String())
	}

	instances := r.Streams.FindStreamInstances(ctx, stream.StreamID, nil, nil)
	return instances, nil
}

func (r *queryResolver) StreamsForUser(ctx context.Context, userID uuid.UUID) ([]*models.Stream, error) {
	secret := middleware.GetSecret(ctx)
	if !(secret.IsUser() && secret.GetOwnerID() == userID) {
		return nil, gqlerror.Errorf("StreamsForUser can only be called for the calling user")
	}
	return r.Streams.FindStreamsForUser(ctx, userID), nil
}

func (r *queryResolver) CompileSchema(ctx context.Context, input gql.CompileSchemaInput) (*gql.CompileSchemaOutput, error) {
	stream := &models.Stream{}
	err := r.Streams.CompileToStream(stream, input.SchemaKind, input.Schema, input.Indexes, nil)
	if err != nil {
		return nil, gqlerror.Errorf("Error compiling schema: %s", err.Error())
	}

	return &gql.CompileSchemaOutput{
		CanonicalAvroSchema: stream.CanonicalAvroSchema,
		CanonicalIndexes:    stream.CanonicalIndexes,
	}, nil
}

func (r *mutationResolver) CreateStream(ctx context.Context, input gql.CreateStreamInput) (*models.Stream, error) {
	// Handle UpdateIfExists (returns if exists)
	if input.UpdateIfExists != nil && *input.UpdateIfExists {
		stream := r.Streams.FindStreamByOrganizationProjectAndName(ctx, input.OrganizationName, input.ProjectName, input.StreamName)
		if stream != nil {
			return r.updateExistingFromCreateStream(ctx, stream, input)
		}
	}

	project := r.Projects.FindProjectByOrganizationAndName(ctx, input.OrganizationName, input.ProjectName)
	if project == nil {
		return nil, gqlerror.Errorf("Project %s/%s not found", input.OrganizationName, input.ProjectName)
	}

	secret := middleware.GetSecret(ctx)
	perms := r.Permissions.ProjectPermissionsForSecret(ctx, secret, project.ProjectID, project.Public)
	if !perms.Create {
		return nil, gqlerror.Errorf("Not allowed to create or modify resources in project %s/%s", input.OrganizationName, input.ProjectName)
	}

	stream, err := r.Streams.CreateStream(ctx, &models.CreateStreamCommand{
		Project:                   project,
		Name:                      input.StreamName,
		SchemaKind:                input.SchemaKind,
		Schema:                    input.Schema,
		Indexes:                   input.Indexes,
		Description:               input.Description,
		Meta:                      input.Meta,
		AllowManualWrites:         input.AllowManualWrites,
		UseLog:                    input.UseLog,
		UseIndex:                  input.UseIndex,
		UseWarehouse:              input.UseWarehouse,
		LogRetentionSeconds:       input.LogRetentionSeconds,
		IndexRetentionSeconds:     input.IndexRetentionSeconds,
		WarehouseRetentionSeconds: input.WarehouseRetentionSeconds,
	})
	if err != nil {
		return nil, gqlerror.Errorf("Error creating stream: %s", err.Error())
	}

	_, err = r.Streams.CreateStreamInstance(ctx, stream, nil, true)
	if err != nil {
		return nil, gqlerror.Errorf("Error creating first instance: %s", err.Error())
	}

	return streamWithProjectPermissions(stream, perms), nil
}

func (r *mutationResolver) updateExistingFromCreateStream(ctx context.Context, stream *models.Stream, input gql.CreateStreamInput) (*models.Stream, error) {
	secret := middleware.GetSecret(ctx)
	perms := r.Permissions.ProjectPermissionsForSecret(ctx, secret, stream.ProjectID, stream.Project.Public)
	if !perms.Create {
		return nil, gqlerror.Errorf("Not allowed to create or modify resources in project %s", stream.Project.Name)
	}

	// check use and retention unchanged (probably doesn't belong here, but tricky to move to service)
	if !checkUse(input.UseLog, stream.UseLog) {
		return nil, gqlerror.Errorf("Cannot update useLog on existing stream")
	}
	if !checkUse(input.UseIndex, stream.UseIndex) {
		return nil, gqlerror.Errorf("Cannot update useIndex on existing stream")
	}
	if !checkUse(input.UseWarehouse, stream.UseWarehouse) {
		return nil, gqlerror.Errorf("Cannot update useWarehouse on existing stream")
	}
	if !checkRetention(input.LogRetentionSeconds, stream.LogRetentionSeconds) {
		return nil, gqlerror.Errorf("Cannot update logRetentionSeconds on existing stream")
	}
	if !checkRetention(input.IndexRetentionSeconds, stream.IndexRetentionSeconds) {
		return nil, gqlerror.Errorf("Cannot update indexRetentionSeconds on existing stream")
	}
	if !checkRetention(input.WarehouseRetentionSeconds, stream.WarehouseRetentionSeconds) {
		return nil, gqlerror.Errorf("Cannot update warehouseRetentionSeconds on existing stream")
	}

	// attempt update
	err := r.Streams.UpdateStream(ctx, &models.UpdateStreamCommand{
		Stream:            stream,
		SchemaKind:        &input.SchemaKind,
		Schema:            &input.Schema,
		Indexes:           input.Indexes,
		Description:       input.Description,
		Meta:              input.Meta,
		AllowManualWrites: input.AllowManualWrites,
	})
	if err != nil {
		return nil, gqlerror.Errorf("Error updating existing stream: %s", err.Error())
	}

	return streamWithProjectPermissions(stream, perms), nil
}

func checkUse(incoming *bool, current bool) bool {
	if incoming != nil {
		return *incoming == current
	}
	return current
}

func checkRetention(incoming *int, current int32) bool {
	if incoming != nil {
		return int32(*incoming) == current
	}
	return current == 0
}

func (r *mutationResolver) UpdateStream(ctx context.Context, input gql.UpdateStreamInput) (*models.Stream, error) {
	stream := r.Streams.FindStream(ctx, input.StreamID)
	if stream == nil {
		return nil, gqlerror.Errorf("Stream with ID %s not found", input.StreamID)
	}

	secret := middleware.GetSecret(ctx)
	perms := r.Permissions.ProjectPermissionsForSecret(ctx, secret, stream.ProjectID, stream.Project.Public)
	if !perms.Create {
		return nil, gqlerror.Errorf("Not allowed to create or modify resources in project %s", stream.Project.Name)
	}

	err := r.Streams.UpdateStream(ctx, &models.UpdateStreamCommand{
		Stream:            stream,
		SchemaKind:        input.SchemaKind,
		Schema:            input.Schema,
		Indexes:           input.Indexes,
		Description:       input.Description,
		AllowManualWrites: input.AllowManualWrites,
	})
	if err != nil {
		return nil, gqlerror.Errorf("Error updating stream: %s", err.Error())
	}

	return streamWithProjectPermissions(stream, perms), nil
}

func (r *mutationResolver) DeleteStream(ctx context.Context, streamID uuid.UUID) (bool, error) {
	stream := r.Streams.FindStream(ctx, streamID)
	if stream == nil {
		return false, gqlerror.Errorf("Stream %s not found", streamID.String())
	}

	secret := middleware.GetSecret(ctx)
	perms := r.Permissions.ProjectPermissionsForSecret(ctx, secret, stream.ProjectID, false)
	if !perms.Create {
		return false, gqlerror.Errorf("Not allowed to perform admin functions in project %s", stream.Project.Name)
	}

	err := r.Streams.DeleteStream(ctx, stream)
	if err != nil {
		return false, gqlerror.Errorf("%s", err.Error())
	}

	return true, nil
}

// MaxInstancesPerStream sets a limit for the number of instances for a stream at any given time
const MaxInstancesPerStream = 25

func (r *mutationResolver) CreateStreamInstance(ctx context.Context, input gql.CreateStreamInstanceInput) (*models.StreamInstance, error) {
	if input.UpdateIfExists != nil && *input.UpdateIfExists {
		if input.Version == nil {
			return nil, gqlerror.Errorf("Cannot set updateIfExists=true without providing a version")
		}
		instance := r.Streams.FindStreamInstanceByVersion(ctx, input.StreamID, *input.Version)
		if instance != nil {
			return r.updateExistingFromCreateStreamInstance(ctx, instance, input)
		}
	}

	stream := r.Streams.FindStream(ctx, input.StreamID)
	if stream == nil {
		return nil, gqlerror.Errorf("Stream %s not found", input.StreamID.String())
	}

	secret := middleware.GetSecret(ctx)
	perms := r.Permissions.StreamPermissionsForSecret(ctx, secret, stream.StreamID, stream.ProjectID, stream.Project.Public)
	if !perms.Write {
		return nil, gqlerror.Errorf("Not allowed to write to stream %s/%s/%s", stream.Project.Organization.Name, stream.Project.Name, stream.Name)
	}

	// check not too many instances
	if stream.InstancesCreatedCount-stream.InstancesDeletedCount >= MaxInstancesPerStream {
		return nil, gqlerror.Errorf("You cannot have more than %d instances per stream. Delete an existing instance to make room for more.", MaxInstancesPerStream)
	}

	makePrimary := false
	if input.MakePrimary != nil {
		makePrimary = *input.MakePrimary
	}

	si, err := r.Streams.CreateStreamInstance(ctx, stream, input.Version, makePrimary)
	if err != nil {
		return nil, gqlerror.Errorf("Error creating stream instance: %s", err.Error())
	}

	return instanceWithPermissions(si, perms), nil
}

func (r *mutationResolver) updateExistingFromCreateStreamInstance(ctx context.Context, instance *models.StreamInstance, input gql.CreateStreamInstanceInput) (*models.StreamInstance, error) {
	secret := middleware.GetSecret(ctx)
	perms := r.Permissions.StreamPermissionsForSecret(ctx, secret, instance.Stream.StreamID, instance.Stream.ProjectID, false)
	if !perms.Write {
		return nil, gqlerror.Errorf("Not allowed to write to instance %s", instance.StreamInstanceID)
	}

	makePrimary := false
	if input.MakePrimary != nil {
		makePrimary = *input.MakePrimary
	}

	err := r.Streams.UpdateStreamInstance(ctx, instance.Stream, instance, false, makePrimary)
	if err != nil {
		return nil, gqlerror.Errorf("Error updating stream instance: %s", err.Error())
	}

	return instanceWithPermissions(instance, perms), nil
}

func (r *mutationResolver) UpdateStreamInstance(ctx context.Context, input gql.UpdateStreamInstanceInput) (*models.StreamInstance, error) {
	instance := r.Streams.FindStreamInstance(ctx, input.StreamInstanceID)
	if instance == nil {
		return nil, gqlerror.Errorf("Stream instance %s not found", input.StreamInstanceID)
	}

	secret := middleware.GetSecret(ctx)
	perms := r.Permissions.StreamPermissionsForSecret(ctx, secret, instance.Stream.StreamID, instance.Stream.ProjectID, false)
	if !perms.Write {
		return nil, gqlerror.Errorf("Not allowed to write to instance %s", input.StreamInstanceID)
	}

	makeFinal := false
	makePrimary := false
	if input.MakeFinal != nil {
		makeFinal = *input.MakeFinal
	}
	if input.MakePrimary != nil {
		makePrimary = *input.MakePrimary
	}

	err := r.Streams.UpdateStreamInstance(ctx, instance.Stream, instance, makeFinal, makePrimary)
	if err != nil {
		return nil, gqlerror.Errorf("Error updating stream instance: %s", err.Error())
	}

	return instanceWithPermissions(instance, perms), nil
}

func (r *mutationResolver) DeleteStreamInstance(ctx context.Context, instanceID uuid.UUID) (bool, error) {
	instance := r.Streams.FindStreamInstance(ctx, instanceID)
	if instance == nil {
		return false, gqlerror.Errorf("Stream instance '%s' not found", instanceID.String())
	}

	secret := middleware.GetSecret(ctx)
	perms := r.Permissions.StreamPermissionsForSecret(ctx, secret, instance.Stream.StreamID, instance.Stream.ProjectID, false)
	if !perms.Write {
		return false, gqlerror.Errorf("Not allowed to write to instance %s", instanceID.String())
	}

	err := r.Streams.DeleteStreamInstance(ctx, instance.Stream, instance)
	if err != nil {
		return false, gqlerror.Errorf("Couldn't delete stream instance: %s", err.Error())
	}

	return true, nil
}

func instanceWithPermissions(si *models.StreamInstance, perms models.StreamPermissions) *models.StreamInstance {
	if si.Stream != nil {
		si.Stream = streamWithPermissions(si.Stream, perms)
	}
	return si
}

func streamWithPermissions(s *models.Stream, perms models.StreamPermissions) *models.Stream {
	// Note: This is a pretty bad approximation, but will do for our current use case
	return streamWithProjectPermissions(s, models.ProjectPermissions{
		View:   perms.Read,
		Create: perms.Write,
		Admin:  perms.Write,
	})
}

func streamWithProjectPermissions(s *models.Stream, perms models.ProjectPermissions) *models.Stream {
	if s.Project != nil {
		s.Project.Permissions = &models.PermissionsUsersProjects{
			View:   perms.View,
			Create: perms.Create,
			Admin:  perms.Admin,
		}
	}
	return s
}
