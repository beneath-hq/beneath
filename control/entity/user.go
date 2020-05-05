package entity

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/go-pg/pg/v9/orm"
	uuid "github.com/satori/go.uuid"
	"gopkg.in/go-playground/validator.v9"

	"gitlab.com/beneath-hq/beneath/control/taskqueue"
	"gitlab.com/beneath-hq/beneath/internal/hub"
	"gitlab.com/beneath-hq/beneath/pkg/log"
	"gitlab.com/beneath-hq/beneath/pkg/timeutil"
)

// User represents a Beneath user
type User struct {
	UserID                 uuid.UUID `sql:",pk,type:uuid,default:uuid_generate_v4()"`
	Username               string    `sql:",notnull",validate:"gte=3,lte=50"`
	Email                  string    `sql:",notnull",validate:"required,email"`
	Name                   string    `sql:",notnull",validate:"required,gte=1,lte=50"`
	Bio                    string    `validate:"omitempty,lte=255"`
	PhotoURL               string    `validate:"omitempty,url,lte=400"`
	GoogleID               string    `sql:",unique",validate:"omitempty,lte=255"`
	GithubID               string    `sql:",unique",validate:"omitempty,lte=255"`
	CreatedOn              time.Time `sql:",default:now()"`
	UpdatedOn              time.Time `sql:",default:now()"`
	PersonalOrganizationID uuid.UUID `sql:",on_delete:restrict,notnull,type:uuid"`
	PersonalOrganization   *Organization
	BillingOrganizationID  uuid.UUID `sql:",on_delete:restrict,notnull,type:uuid"`
	BillingOrganization    *Organization
	Projects               []*Project `pg:"many2many:permissions_users_projects,fk:user_id,joinFK:project_id"`
	Secrets                []*UserSecret
	Master                 bool `sql:",notnull,default:false"`
	ReadQuota              *int64
	WriteQuota             *int64
}

var (
	userUsernameRegex    *regexp.Regexp
	nonAlphanumericRegex *regexp.Regexp
)

const (
	usernameMinLength = 3
	usernameMaxLength = 50
)

// configure constants and validator
func init() {
	userUsernameRegex = regexp.MustCompile("^[_a-z][_\\-a-z0-9]*$")
	nonAlphanumericRegex = regexp.MustCompile("[^a-zA-Z0-9]+")
	GetValidator().RegisterStructValidation(userValidation, User{})
}

// custom user validation
func userValidation(sl validator.StructLevel) {
	u := sl.Current().Interface().(User)

	if !userUsernameRegex.MatchString(u.Username) {
		sl.ReportError(u.Username, "Username", "", "alphanumericorunderscore", "")
	}
}

// FindUser returns the matching user or nil
func FindUser(ctx context.Context, userID uuid.UUID) *User {
	user := &User{
		UserID: userID,
	}
	err := hub.DB.ModelContext(ctx, user).WherePK().Column("user.*", "Projects", "Projects.Organization.name", "BillingOrganization", "PersonalOrganization").Select()
	if !AssertFoundOne(err) {
		return nil
	}
	return user
}

// FindUserByEmail returns user with email (if exists)
func FindUserByEmail(ctx context.Context, email string) *User {
	user := &User{}
	err := hub.DB.ModelContext(ctx, user).Where("lower(email) = lower(?)", email).Select()
	if !AssertFoundOne(err) {
		return nil
	}
	return user
}

// FindUserByUsername returns user with username (if exists)
func FindUserByUsername(ctx context.Context, username string) *User {
	user := &User{}
	err := hub.DB.ModelContext(ctx, user).Where("lower(username) = lower(?)", username).Column("user.*", "Projects", "Projects.Organization.name").Select()
	if !AssertFoundOne(err) {
		return nil
	}
	return user
}

// CreateOrUpdateUser consolidates and returns the user matching the args
func CreateOrUpdateUser(ctx context.Context, githubID, googleID, email, nickname, name, photoURL string) (*User, error) {
	defaultBillingPlan := FindDefaultBillingPlan(ctx)

	user := &User{}
	create := false

	tx, err := hub.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback() // defer rollback on error

	// find user by ID (and lock for update)
	var query *orm.Query
	if githubID != "" {
		query = tx.Model(user).Where("github_id = ?", githubID)
	} else if googleID != "" {
		query = tx.Model(user).Where("google_id = ?", googleID)
	} else {
		panic(fmt.Errorf("CreateOrUpdateUser neither githubID nor googleID set"))
	}

	err = query.For("UPDATE").Select()
	if !AssertFoundOne(err) {
		// find user by email
		err = tx.Model(user).Where("lower(email) = lower(?)", email).For("UPDATE").Select()
		if !AssertFoundOne(err) {
			create = true
		}
	}

	// set user fields
	if githubID != "" {
		user.GithubID = githubID
	}
	if googleID != "" {
		user.GoogleID = googleID
	}
	if email != "" {
		user.Email = email
	}
	if user.Name == "" {
		user.Name = name
	}
	if user.PhotoURL == "" {
		user.PhotoURL = photoURL
	}

	// if updating, finalize and return
	if !create {
		// validate
		err = GetValidator().Struct(user)
		if err != nil {
			return nil, err
		}

		// update
		user.UpdatedOn = time.Now()
		err = tx.Update(user)
		if err != nil {
			return nil, err
		}

		// log
		log.S.Infow(
			"control updated user",
			"user_id", user.UserID,
		)

		// commit
		err = tx.Commit()
		if err != nil {
			return nil, err
		}

		return user, nil
	}

	// we're creating a new user

	// not a platform admin
	user.Master = false

	// prepare "personal" organization
	org := &Organization{
		Personal: true,
	}

	// try out all possible usernames
	usernameSeeds := user.usernameSeeds(nickname)
	for _, username := range usernameSeeds {
		// savepoint in case username is taken
		_, err = tx.Exec("SAVEPOINT bi")
		if err != nil {
			return nil, err
		}

		// update org and user
		org.Name = username
		user.Username = username
		if user.Name == "" {
			user.Name = username
		}

		// insert user and org
		err = tx.Insert(org)
		if err == nil {
			user.PersonalOrganizationID = org.OrganizationID
			user.BillingOrganizationID = org.OrganizationID
			err = tx.Insert(user)
		}

		// success
		if err == nil {
			break
		}

		// rollback on name error
		if isUniqueError(err) {
			// rollback to before error, then try next username
			_, err = tx.Exec("ROLLBACK TO SAVEPOINT bi")
			if err != nil {
				return nil, err
			}
			continue
		}

		// unexpected error
		return nil, err
	}

	// add user to user-organization permissions table
	err = tx.Insert(&PermissionsUsersOrganizations{
		UserID:         user.UserID,
		OrganizationID: org.OrganizationID,
		View:           true,
		Create:         true,
		Admin:          true,
	})
	if err != nil {
		return nil, err
	}

	// create billing info
	bi := &BillingInfo{
		OrganizationID: org.OrganizationID,
		BillingPlanID:  defaultBillingPlan.BillingPlanID,
	}
	_, err = tx.Model(bi).Insert()
	if err != nil {
		return nil, err
	}

	// commit it all
	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	// log new user
	log.S.Infow(
		"control created user",
		"user_id", user.UserID,
	)

	return user, nil
}

// Delete removes the user from the database
// if user is the last one in an organization, in resolver, need to check if there are any services or projects that are still tied to the organization
func (u *User) Delete(ctx context.Context) error {
	err := hub.DB.WithContext(ctx).Delete(u)
	if err != nil {
		return err
	}

	org := FindOrganization(ctx, u.BillingOrganizationID)
	if org == nil {
		panic("could not find organization")
	}

	// if the organization now has no more members, trigger bill (which will also delete organization)
	if len(org.Users) == 0 {
		billingInfo := FindBillingInfo(ctx, org.OrganizationID)
		if billingInfo == nil {
			panic("could not find billing info for the user's organization")
		}

		err := taskqueue.Submit(context.Background(), &ComputeBillResourcesTask{
			OrganizationID: org.OrganizationID,
			Timestamp:      timeutil.Next(time.Now(), billingInfo.BillingPlan.Period), // simulate that we are at the beginning of the next billing cycle
		})
		if err != nil {
			log.S.Errorw("Error creating task", err)
		}
	}

	return nil
}

// UpdateDescription updates user's name and/or bio
func (u *User) UpdateDescription(ctx context.Context, username *string, name *string, bio *string, photoURL *string) error {
	if username != nil {
		u.Username = *username
	}
	if name != nil {
		u.Name = *name
	}
	if bio != nil {
		u.Bio = *bio
	}
	if photoURL != nil {
		u.PhotoURL = *photoURL
	}

	// validate
	err := GetValidator().Struct(u)
	if err != nil {
		return err
	}

	u.UpdatedOn = time.Now()
	_, err = hub.DB.ModelContext(ctx, u).Column("username", "name", "bio", "photo_url", "updated_on").WherePK().Update()
	return err
}

// UpdateQuotas change the user's quotas
func (u *User) UpdateQuotas(ctx context.Context, readQuota *int64, writeQuota *int64) error {
	u.ReadQuota = readQuota
	u.WriteQuota = writeQuota

	// validate
	err := GetValidator().Struct(u)
	if err != nil {
		return err
	}

	u.UpdatedOn = time.Now()
	_, err = hub.DB.ModelContext(ctx, u).Column("read_quota", "write_quota", "updated_on").WherePK().Update()
	return err
}

func (u *User) usernameSeeds(nickname string) []string {
	// gather candidates
	var seeds []string
	var shortest string

	// base on nickname
	username := nonAlphanumericRegex.ReplaceAllString(nickname, "")
	if len(username) >= usernameMinLength {
		seeds = append(seeds, finalizeUsernameSeed(username))
		shortest = username
	}
	if len(username) < len(shortest) {
		shortest = username
	}

	// base on email
	username = strings.Split(u.Email, "@")[0]
	username = nonAlphanumericRegex.ReplaceAllString(username, "")
	if len(username) >= usernameMinLength {
		seeds = append(seeds, finalizeUsernameSeed(username))
	}
	if len(username) < len(shortest) {
		shortest = username
	}

	// base on name
	username = nonAlphanumericRegex.ReplaceAllString(u.Name, "")
	if len(username) >= usernameMinLength {
		seeds = append(seeds, finalizeUsernameSeed(username))
	}
	if len(username) < len(shortest) {
		shortest = username
	}

	// final fallback -- a uuid
	username = shortest + uuid.NewV4().String()[0:8]
	seeds = append(seeds, finalizeUsernameSeed(username))

	return seeds
}

func finalizeUsernameSeed(seed string) string {
	seed = strings.ToLower(seed)
	if len(seed) == (usernameMinLength-1) && seed[0] != '_' {
		seed = "_" + seed
	}
	if unicode.IsDigit(rune(seed[0])) {
		seed = "b" + seed
	}
	if len(seed) > usernameMaxLength {
		seed = seed[0:usernameMaxLength]
	}
	return seed
}

func isUniqueError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(strings.ToLower(err.Error()), "violates unique constraint")
}
