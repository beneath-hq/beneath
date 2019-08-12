package resolver

import (
	"context"

	"github.com/beneath-core/beneath-go/control/auth"
	"github.com/beneath-core/beneath-go/control/gql"
	"github.com/beneath-core/beneath-go/control/model"
	uuid "github.com/satori/go.uuid"
	"github.com/vektah/gqlparser/gqlerror"
)

// Project returns the gql.ProjectResolver
func (r *Resolver) Project() gql.ProjectResolver {
	return &projectResolver{r}
}

type projectResolver struct{ *Resolver }

func (r *projectResolver) ProjectID(ctx context.Context, obj *model.Project) (string, error) {
	return obj.ProjectID.String(), nil
}

func (r *queryResolver) ExploreProjects(ctx context.Context) ([]*model.Project, error) {
	return model.FindProjects(), nil
}

func (r *queryResolver) ProjectByName(ctx context.Context, name string) (*model.Project, error) {
	project := model.FindProjectByName(name)
	if project == nil {
		return nil, gqlerror.Errorf("Project %s not found", name)
	}

	key := auth.GetKey(ctx)
	if !key.ReadsProject(project.ProjectID) {
		return nil, gqlerror.Errorf("Not allowed to read project %s", name)
	}

	return project, nil
}

func (r *queryResolver) ProjectByID(ctx context.Context, projectID uuid.UUID) (*model.Project, error) {
	key := auth.GetKey(ctx)
	if !key.ReadsProject(projectID) {
		return nil, gqlerror.Errorf("Not allowed to read project %s", projectID.String())
	}

	project := model.FindProject(projectID)
	if project == nil {
		return nil, gqlerror.Errorf("Project %s not found", projectID.String())
	}

	return project, nil
}

func (r *mutationResolver) CreateProject(ctx context.Context, name string, displayName string, site *string, description *string, photoURL *string) (*model.Project, error) {
	key := auth.GetKey(ctx)
	if !key.IsPersonal() {
		return nil, gqlerror.Errorf("Not allowed to create project")
	}

	project := &model.Project{
		Name:        name,
		DisplayName: displayName,
		Site:        DereferenceString(site),
		Description: DereferenceString(description),
		PhotoURL:    DereferenceString(photoURL),
		Public:      true,
	}

	err := project.CreateWithUser(*key.UserID)
	if err != nil {
		return nil, err
	}

	return project, nil
}

func (r *mutationResolver) UpdateProject(ctx context.Context, projectID uuid.UUID, displayName *string, site *string, description *string, photoURL *string) (*model.Project, error) {
	key := auth.GetKey(ctx)
	if !key.EditsProject(projectID) {
		return nil, gqlerror.Errorf("Not allowed to edit project %s", projectID.String())
	}

	project := model.FindProject(projectID)
	if project == nil {
		return nil, gqlerror.Errorf("Project %s not found", projectID.String())
	}

	err := project.UpdateDetails(displayName, site, description, photoURL)
	if err != nil {
		return nil, gqlerror.Errorf(err.Error())
	}

	return project, nil
}

func (r *mutationResolver) AddUserToProject(ctx context.Context, email string, projectID uuid.UUID) (*model.User, error) {
	key := auth.GetKey(ctx)
	if !key.EditsProject(projectID) {
		return nil, gqlerror.Errorf("Not allowed to edit project %s", projectID.String())
	}

	user := model.FindUserByEmail(email)
	if user == nil {
		return nil, gqlerror.Errorf("No user found with that email")
	}

	project := &model.Project{
		ProjectID: projectID,
	}

	err := project.AddUser(user.UserID)
	if err != nil {
		return nil, gqlerror.Errorf(err.Error())
	}

	return user, nil
}

func (r *mutationResolver) RemoveUserFromProject(ctx context.Context, userID uuid.UUID, projectID uuid.UUID) (bool, error) {
	key := auth.GetKey(ctx)
	if !key.EditsProject(projectID) {
		return false, gqlerror.Errorf("Not allowed to edit project %s", projectID.String())
	}

	project := model.FindProject(projectID)
	if project == nil {
		return false, gqlerror.Errorf("Project %s not found", projectID.String())
	}

	if len(project.Users) < 2 {
		return false, gqlerror.Errorf("Can't remove last member of project")
	}

	err := project.RemoveUser(userID)
	if err != nil {
		return false, gqlerror.Errorf(err.Error())
	}

	return true, nil
}
