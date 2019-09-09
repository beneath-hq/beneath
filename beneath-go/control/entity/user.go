package entity

import (
	"context"
	"regexp"
	"time"

	"github.com/beneath-core/beneath-go/core/log"

	"github.com/go-pg/pg/orm"
	uuid "github.com/satori/go.uuid"
	"gopkg.in/go-playground/validator.v9"

	"github.com/beneath-core/beneath-go/db"
)

// User represents a Beneath user
type User struct {
	UserID    uuid.UUID `sql:",pk,type:uuid,default:uuid_generate_v4()"`
	Username  string    `sql:",unique",validate:"omitempty,gte=3,lte=16"`
	Email     string    `sql:",unique,notnull",validate:"required,email"`
	Name      string    `sql:",notnull",validate:"required,gte=4,lte=50"`
	Bio       string    `validate:"omitempty,lte=255"`
	PhotoURL  string    `validate:"omitempty,url,lte=255"`
	GoogleID  string    `sql:",unique",validate:"omitempty,lte=255"`
	GithubID  string    `sql:",unique",validate:"omitempty,lte=255"`
	CreatedOn time.Time `sql:",default:now()"`
	UpdatedOn time.Time `sql:",default:now()"`
	DeletedOn time.Time
	Projects  []*Project `pg:"many2many:projects_users,fk:user_id,joinFK:project_id"`
	Secrets   []*Secret
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

// FindUser returns the matching user or nil
func FindUser(ctx context.Context, userID uuid.UUID) *User {
	user := &User{
		UserID: userID,
	}
	err := db.DB.ModelContext(ctx, user).WherePK().Column("user.*", "Projects").Select()
	if !AssertFoundOne(err) {
		return nil
	}
	return user
}

// FindUserByEmail returns user with email (if exists)
func FindUserByEmail(ctx context.Context, email string) *User {
	user := &User{}
	err := db.DB.ModelContext(ctx, user).Where("lower(email) = lower(?)", email).Select()
	if !AssertFoundOne(err) {
		return nil
	}
	return user
}

// CreateOrUpdateUser consolidates and returns the user matching the args
func CreateOrUpdateUser(ctx context.Context, githubID, googleID, email, name, photoURL string) (*User, error) {
	user := &User{}
	create := false

	var query *orm.Query
	if githubID != "" {
		query = db.DB.ModelContext(ctx, user).Where("github_id = ?", githubID)
	} else if googleID != "" {
		query = db.DB.ModelContext(ctx, user).Where("google_id = ?", googleID)
	} else {
		panic("CreateOrUpdateUser neither githubID nor googleID set")
	}

	err := query.Select()
	if !AssertFoundOne(err) {
		userByEmail := FindUserByEmail(ctx, email)
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
		err = db.DB.WithContext(ctx).Insert(user)
	} else {
		err = db.DB.WithContext(ctx).Update(user)
	}

	if err != nil {
		return nil, err
	}

	if create {
		log.S.Infow(
			"control created user",
			"user_id", user.UserID,
		)
	} else {
		log.S.Infow(
			"control updated user",
			"user_id", user.UserID,
		)
	}

	return user, nil
}

// Delete removes the user from the database
func (u *User) Delete(ctx context.Context) error {
	return db.DB.WithContext(ctx).Delete(u)
}

// UpdateDescription updates user's name and/or bio
func (u *User) UpdateDescription(ctx context.Context, name *string, bio *string) error {
	if name != nil {
		u.Name = *name
	}
	if bio != nil {
		u.Bio = *bio
	}

	// validate
	err := GetValidator().Struct(u)
	if err != nil {
		return err
	}

	_, err = db.DB.ModelContext(ctx, u).Column("name", "bio").WherePK().Update()
	return err
}
