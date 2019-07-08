package model

import (
	"log"
	"regexp"
	"time"

	"github.com/go-pg/pg/orm"
	uuid "github.com/satori/go.uuid"
	"gopkg.in/go-playground/validator.v9"

	"github.com/beneath-core/beneath-go/control/db"
)

// User represents a Beneath user
type User struct {
	UserID    uuid.UUID  `sql:",pk,type:uuid,default:uuid_generate_v4()"`
	Username  string     `sql:",unique",validate:"omitempty,gte=3,lte=16"`
	Email     string     `sql:",unique,notnull",validate:"required,email"`
	Name      string     `sql:",notnull",validate:"required,gte=4,lte=50"`
	Bio       string     `validate:"omitempty,lte=255"`
	PhotoURL  string     `validate:"omitempty,url,lte=255"`
	GoogleID  string     `sql:",unique",validate:"omitempty,lte=255"`
	GithubID  string     `sql:",unique",validate:"omitempty,lte=255"`
	CreatedOn time.Time  `sql:",default:now()"`
	UpdatedOn time.Time  `sql:",default:now()"`
	Projects  []*Project `pg:"many2many:users_projects,joinFK:user_id"`
	Keys      []*Key
}

// UserToProject represnts the many-to-many relationship between users and projects
type UserToProject struct {
	tableName struct{}  `sql:"users_projects,alias:up"`
	UserID    uuid.UUID `sql:",pk,type:uuid"`
	User      *User
	ProjectID uuid.UUID `sql:",pk,type:uuid"`
	Project   *Project
}

var (
	userUsernameRegex *regexp.Regexp
)

// configure constants and validator
func init() {
	userUsernameRegex = regexp.MustCompile("^[_a-z][_\\-a-z0-9]*$")
	GetValidator().RegisterStructValidation(userValidation, User{})
}

// custom user validation
func userValidation(sl validator.StructLevel) {
	u := sl.Current().Interface().(User)

	if u.Username != "" {
		if !userUsernameRegex.MatchString(u.Username) {
			sl.ReportError(u.Username, "Username", "", "alphanumericorunderscore", "")
		}
	}
}

// FindOneUserByEmail returns user with email (if exists)
func FindOneUserByEmail(email string) *User {
	user := &User{}
	err := db.DB.Model(user).Where("lower(email) = lower(?)", email).Select()
	if !AssertFoundOne(err) {
		return nil
	}
	return user
}

// CreateOrUpdateUser consolidates and returns the user matching the args
func CreateOrUpdateUser(githubID, googleID, email, name, photoURL string) (*User, error) {
	user := &User{}
	create := false

	var query *orm.Query
	if githubID != "" {
		query = db.DB.Model(user).Where("github_id = ?", githubID)
	} else if googleID != "" {
		query = db.DB.Model(user).Where("google_id = ?", googleID)
	} else {
		log.Panic("CreateOrUpdateUser neither githubID nor googleID set")
	}

	err := query.Select()
	if !AssertFoundOne(err) {
		userByEmail := FindOneUserByEmail(email)
		if userByEmail == nil {
			create = true
		} else {
			user = userByEmail
		}
	}

	user.GithubID = githubID
	user.GoogleID = googleID
	user.Email = email
	user.Name = name
	user.PhotoURL = photoURL

	// validate
	err = GetValidator().Struct(user)
	if err != nil {
		return nil, err
	}

	// insert or update
	err = nil
	if create {
		err = db.DB.Insert(user)
	} else {
		err = db.DB.Update(user)
	}

	if err != nil {
		return nil, err
	}

	if create {
		log.Printf("Created userID <%s>", user.UserID)
	} else {
		log.Printf("Updated userID <%s>", user.UserID)
	}

	return user, nil
}

// Delete removes the user from the database
func (u *User) Delete() error {
	return db.DB.Delete(u)
}
