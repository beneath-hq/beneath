package permissions

import (
	"context"
	"fmt"

	uuid "github.com/satori/go.uuid"

	"github.com/beneath-hq/beneath/models"
)

// TablePermissionsForSecret gets the secret owner's permissions for a table
func (s *Service) TablePermissionsForSecret(ctx context.Context, secret models.Secret, tableID uuid.UUID, projectID uuid.UUID, public bool) models.TablePermissions {
	switch secret := secret.(type) {
	case *models.UserSecret:
		return s.tablePermissionsForUserSecret(ctx, secret, tableID, projectID, public)
	case *models.ServiceSecret:
		return s.tablePermissionsForServiceSecret(ctx, secret, tableID, projectID, public)
	case *models.AnonymousSecret:
		return models.TablePermissions{Read: public}
	default:
		panic(fmt.Errorf("unrecognized secret type %T", secret))
	}
}

func (s *Service) tablePermissionsForUserSecret(ctx context.Context, secret *models.UserSecret, tableID uuid.UUID, projectID uuid.UUID, public bool) models.TablePermissions {
	projectPerms := s.CachedUserProjectPermissions(ctx, secret.UserID, projectID)
	return models.TablePermissions{
		Read:  (projectPerms.View && !secret.PublicOnly) || public,
		Write: projectPerms.Create && !secret.ReadOnly && (!secret.PublicOnly || public),
	}
}

func (s *Service) tablePermissionsForServiceSecret(ctx context.Context, secret *models.ServiceSecret, tableID uuid.UUID, projectID uuid.UUID, public bool) models.TablePermissions {
	return s.CachedServiceTablePermissions(ctx, secret.ServiceID, tableID)
}

// ProjectPermissionsForSecret gets the secret owner's permissions for a project
func (s *Service) ProjectPermissionsForSecret(ctx context.Context, secret models.Secret, projectID uuid.UUID, public bool) models.ProjectPermissions {
	switch secret := secret.(type) {
	case *models.UserSecret:
		return s.projectPermissionsForUserSecret(ctx, secret, projectID, public)
	case *models.ServiceSecret:
		return s.projectPermissionsForServiceSecret(ctx, secret, projectID, public)
	case *models.AnonymousSecret:
		return models.ProjectPermissions{View: public}
	default:
		panic(fmt.Errorf("unrecognized secret type %T", secret))
	}
}

func (s *Service) projectPermissionsForUserSecret(ctx context.Context, secret *models.UserSecret, projectID uuid.UUID, public bool) models.ProjectPermissions {
	if secret.PublicOnly && !public {
		return models.ProjectPermissions{}
	}
	if secret.ReadOnly && public {
		return models.ProjectPermissions{View: true}
	}
	perms := s.CachedUserProjectPermissions(ctx, secret.UserID, projectID)
	if public {
		perms.View = true
	}
	if secret.ReadOnly {
		perms.Admin = false
		perms.Create = false
	}
	return perms
}

func (s *Service) projectPermissionsForServiceSecret(ctx context.Context, secret *models.ServiceSecret, projectID uuid.UUID, public bool) models.ProjectPermissions {
	return models.ProjectPermissions{}
}

// OrganizationPermissionsForSecret gets the secret owner's permissions for a organization
func (s *Service) OrganizationPermissionsForSecret(ctx context.Context, secret models.Secret, organizationID uuid.UUID) models.OrganizationPermissions {
	switch secret := secret.(type) {
	case *models.UserSecret:
		return s.organizationPermissionsForUserSecret(ctx, secret, organizationID)
	case *models.ServiceSecret:
		return s.organizationPermissionsForServiceSecret(ctx, secret, organizationID)
	case *models.AnonymousSecret:
		return models.OrganizationPermissions{} // Everyone has basic view access for orgs. The View property here indicates insider-view.
	default:
		panic(fmt.Errorf("unrecognized secret type %T", secret))
	}
}

func (s *Service) organizationPermissionsForUserSecret(ctx context.Context, secret *models.UserSecret, organizationID uuid.UUID) models.OrganizationPermissions {
	if secret.ReadOnly || secret.PublicOnly {
		return models.OrganizationPermissions{}
	}
	return s.CachedUserOrganizationPermissions(ctx, secret.UserID, organizationID)
}

func (s *Service) organizationPermissionsForServiceSecret(ctx context.Context, secret *models.ServiceSecret, organizationID uuid.UUID) models.OrganizationPermissions {
	return models.OrganizationPermissions{}
}
