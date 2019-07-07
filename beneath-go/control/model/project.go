package model

import (
	"regexp"
	"time"

	uuid "github.com/satori/go.uuid"
	"gopkg.in/go-playground/validator.v9"
)

// constants
var (
	projectNameRegex *regexp.Regexp
)

// configure constants and validator
func init() {
	projectNameRegex = regexp.MustCompile("^[_a-z][_\\-a-z0-9]*$")
	GetValidator().RegisterStructValidation(projectValidation, Project{})
}

// custom project validation
func projectValidation(sl validator.StructLevel) {
	p := sl.Current().Interface().(Project)

	if !projectNameRegex.MatchString(p.Name) {
		sl.ReportError(p.Name, "Name", "", "alphanumericorunderscore", "")
	}
}

// Project represents a Beneath project
type Project struct {
	ProjectID   uuid.UUID `sql:",pk,type:uuid"`
	Name        string    `sql:",unique,notnull",validate:"required,gte=3,lte=16"`
	DisplayName string    `sql:",notnull",validate:"required,gte=3,lte=40"`
	Site        string    `validate:"omitempty,url,lte=255"`
	Description string    `validate:"omitempty,lte=255"`
	PhotoURL    string    `validate:"omitempty,url,lte=255"`
	Public      bool      `sql:",notnull,default:true"`
	CreatedOn   time.Time `sql:",default:now()"`
	UpdatedOn   time.Time `sql:",default:now()"`
	Keys        []*Key
	Streams     []*Stream
	Users       []*User `pg:"many2many:users_projects,joinFK:project_id"`
}

// ProjectHasUser returns true iff user is a member of project
func ProjectHasUser(projectID uuid.UUID, userID uuid.UUID) bool {
	// TODO
	// const result = await getConnection()
	//     .createQueryBuilder(Project, "project")
	//     .innerJoin("project.users", "user")
	//     .where("user.userId = :userId", { userId })
	//     .andWhere("project.projectId = :projectId", { projectId })
	//     .select(["project.projectId"])
	//     .cache(`projects_users:${projectId}:${userId}`, 300000)
	//     .getOne();
	//   return !!result;
	return false
}

// AddUser makes user a member of project
func (p *Project) AddUser(user *User) {
	// TODO
	// this.users.push(user);
	// await this.save();
}

// RemoveUserByID removes a member from the project
func (p *Project) RemoveUserByID(userID uuid.UUID) {
	// TODO
	// await Project.createQueryBuilder()
	//   .relation("users")
	//   .of({ projectId: this.projectId })
	//   .remove({ userId });

	// const cache = getConnection().queryResultCache;
	// if (cache) {
	//   await cache.remove([`projects_users:${this.projectId}:${userId}`]);
	// }
}
