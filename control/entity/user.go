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

	"gitlab.com/beneath-hq/beneath/internal/hub"
	"gitlab.com/beneath-hq/beneath/pkg/log"
)

// User represents a Beneath user
type User struct {
	UserID                uuid.UUID `sql:",pk,type:uuid,default:uuid_generate_v4()"`
	Email                 string    `sql:",notnull",validate:"required,email"`
	GoogleID              string    `sql:",unique",validate:"omitempty,lte=255"`
	GithubID              string    `sql:",unique",validate:"omitempty,lte=255"`
	CreatedOn             time.Time `sql:",default:now()"`
	UpdatedOn             time.Time `sql:",default:now()"`
	Master                bool      `sql:",notnull,default:false"` // NOTE: when updating value, clear secret cache
	ConsentTerms          bool      `sql:",notnull,default:false"`
	ConsentNewsletter     bool      `sql:",notnull,default:false"`
	ReadQuota             *int64    // NOTE: when updating value, clear secret cache
	WriteQuota            *int64    // NOTE: when updating value, clear secret cache
	BillingOrganizationID uuid.UUID `sql:",on_delete:restrict,notnull,type:uuid"` // NOTE: when updating value, clear secret cache
	BillingOrganization   *Organization
	Projects              []*Project      `pg:"many2many:permissions_users_projects,fk:user_id,joinFK:project_id"`
	Organizations         []*Organization `pg:"many2many:permissions_users_organizations,fk:user_id,joinFK:organization_id"`
	Secrets               []*UserSecret
}

var (
	nonAlphanumericRegex *regexp.Regexp
)

const (
	// used for generating username seeds
	usernameMinLength = 3
	usernameMaxLength = 50
)

func init() {
	nonAlphanumericRegex = regexp.MustCompile("[^a-zA-Z0-9]+")
}

// FindUser returns the matching user or nil
func FindUser(ctx context.Context, userID uuid.UUID) *User {
	user := &User{
		UserID: userID,
	}
	err := hub.DB.ModelContext(ctx, user).WherePK().Column("user.*").Select()
	if !AssertFoundOne(err) {
		return nil
	}
	return user
}

// CreateOrUpdateUser consolidates and returns the user matching the args
func CreateOrUpdateUser(ctx context.Context, githubID, googleID, email, nickname, name, photoURL string) (*User, error) {
	defaultBillingPlan := FindDefaultBillingPlan(ctx)

	user := &User{}
	org := &Organization{}
	create := false

	tx, err := hub.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback() // defer rollback on error

	// find user by ID (and lock for update)
	var userQuery *orm.Query
	if githubID != "" {
		userQuery = tx.Model(user).Where("github_id = ?", githubID)
	} else if googleID != "" {
		userQuery = tx.Model(user).Where("google_id = ?", googleID)
	} else {
		panic(fmt.Errorf("CreateOrUpdateUser neither githubID nor googleID set"))
	}

	err = userQuery.For("UPDATE").Select()
	if !AssertFoundOne(err) {
		// find user by email
		err = tx.Model(user).Where("lower(email) = lower(?)", email).For("UPDATE").Select()
		if !AssertFoundOne(err) {
			create = true
		}
	}

	// if not create, select personal org for update
	if !create {
		err = tx.Model(org).Where("user_id = ?", user.UserID).For("UPDATE").Select()
		if err != nil {
			panic(fmt.Errorf("Couldn't find personal organization for user: %s", err.Error()))
		}
	}

	// set fields
	if githubID != "" {
		user.GithubID = githubID
	}
	if googleID != "" {
		user.GoogleID = googleID
	}
	if email != "" {
		user.Email = email
	}
	if org.DisplayName == "" {
		org.DisplayName = name
	}
	if org.PhotoURL == "" {
		org.PhotoURL = photoURL
	}

	// if updating, finalize and return
	if !create {
		// validate and save user
		user.UpdatedOn = time.Now()
		err = GetValidator().Struct(user)
		if err != nil {
			return nil, err
		}
		err = tx.Update(user)
		if err != nil {
			return nil, err
		}

		// validate and save org
		org.UpdatedOn = time.Now()
		err = GetValidator().Struct(org)
		if err != nil {
			return nil, err
		}
		err = tx.Update(org)
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

	// set the organization's default quotas
	org.PrepaidReadQuota = &defaultBillingPlan.ReadQuota
	org.PrepaidWriteQuota = &defaultBillingPlan.WriteQuota
	org.ReadQuota = &defaultBillingPlan.ReadQuota
	org.WriteQuota = &defaultBillingPlan.WriteQuota

	// try out all possible usernames
	created := false
	usernameSeeds := usernameSeeds(email, name, nickname)
	for _, username := range usernameSeeds {
		// update org and user
		org.Name = username
		if org.DisplayName == "" {
			org.DisplayName = username
		}

		// validate org
		err = GetValidator().Struct(org)
		if err != nil {
			// continue (probably username is blacklisted)
			continue
		}

		// savepoint in case name is taken
		_, err = tx.Exec("SAVEPOINT bi")
		if err != nil {
			return nil, err
		}

		// insert org
		err = tx.Insert(org)
		if err == nil {
			// insert user (with org as billing)
			user.BillingOrganizationID = org.OrganizationID
			err = tx.Insert(user)
			if err != nil {
				return nil, err
			}

			// update org with user ID
			org.UserID = &user.UserID
			_, err = tx.Model(org).Column("user_id").WherePK().Update()
			if err != nil {
				return nil, err
			}

			// success
			created = true
			break
		}

		// failure, err != nil

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

	// if not created, error
	if !created {
		return nil, fmt.Errorf("couldn't create user for any username seeds")
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

// RegisterConsent updates the user's consent preferences
func (u *User) RegisterConsent(ctx context.Context, terms *bool, newsletter *bool) error {
	if !u.ConsentTerms && (terms == nil || !*terms) {
		return fmt.Errorf("Cannot continue without consent to the terms of service")
	}

	if terms != nil {
		if !*terms {
			return fmt.Errorf("You cannot withdraw consent to the terms of service; instead, delete your user")
		}
		u.ConsentTerms = *terms
	}

	if newsletter != nil {
		u.ConsentNewsletter = *newsletter
	}

	// validate and save
	u.UpdatedOn = time.Now()
	err := GetValidator().Struct(u)
	if err != nil {
		return err
	}
	_, err = hub.DB.ModelContext(ctx, u).Column("consent_terms", "consent_newsletter", "updated_on").WherePK().Update()
	if err != nil {
		return err
	}

	return nil
}

// UpdateQuotas change the user's quotas
func (u *User) UpdateQuotas(ctx context.Context, readQuota *int64, writeQuota *int64) error {
	u.ReadQuota = readQuota
	u.WriteQuota = writeQuota

	// validate and save
	u.UpdatedOn = time.Now()
	err := GetValidator().Struct(u)
	if err != nil {
		return err
	}
	_, err = hub.DB.ModelContext(ctx, u).Column("read_quota", "write_quota", "updated_on").WherePK().Update()
	if err != nil {
		return err
	}

	getSecretCache().ClearForUser(ctx, u.UserID)
	return nil
}

func usernameSeeds(email string, name string, nickname string) []string {
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
	username = strings.Split(email, "@")[0]
	username = nonAlphanumericRegex.ReplaceAllString(username, "")
	if len(username) >= usernameMinLength {
		seeds = append(seeds, finalizeUsernameSeed(username))
	}
	if len(username) < len(shortest) {
		shortest = username
	}

	// base on name
	username = nonAlphanumericRegex.ReplaceAllString(name, "")
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
