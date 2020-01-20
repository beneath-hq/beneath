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

	"github.com/beneath-core/beneath-go/core/log"
	"github.com/beneath-core/beneath-go/db"
)

// User represents a Beneath user
type User struct {
	UserID         uuid.UUID `sql:",pk,type:uuid,default:uuid_generate_v4()"`
	Username       string    `sql:",notnull",validate:"gte=3,lte=50"`
	Email          string    `sql:",notnull",validate:"required,email"`
	Name           string    `sql:",notnull",validate:"required,gte=1,lte=50"`
	Bio            string    `validate:"omitempty,lte=255"`
	PhotoURL       string    `validate:"omitempty,url,lte=400"`
	GoogleID       string    `sql:",unique",validate:"omitempty,lte=255"`
	GithubID       string    `sql:",unique",validate:"omitempty,lte=255"`
	CreatedOn      time.Time `sql:",default:now()"`
	UpdatedOn      time.Time `sql:",default:now()"`
	DeletedOn      time.Time
	OrganizationID *uuid.UUID `sql:",type:uuid"`
	Organization   *Organization
	Projects       []*Project `pg:"many2many:permissions_users_projects,fk:user_id,joinFK:project_id"`
	Secrets        []*UserSecret
	ReadQuota      int64
	WriteQuota     int64
}

var (
	userUsernameRegex    *regexp.Regexp
	nonAlphanumericRegex *regexp.Regexp
)

const (
	// DefaultUserReadQuota is the Free plan's read quota for user keys
	DefaultUserReadQuota = 100000000 // unit: bytes

	// DefaultUserWriteQuota is the Free plan's write quota for user keys
	DefaultUserWriteQuota = 100000000 // unit: bytes

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
	err := db.DB.ModelContext(ctx, user).WherePK().Column("user.*", "Projects", "Organization").Select()
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

// FindUserByUsername returns user with username (if exists)
func FindUserByUsername(ctx context.Context, username string) *User {
	user := &User{}
	err := db.DB.ModelContext(ctx, user).Where("lower(username) = lower(?)", username).Relation("Projects").Select()
	if !AssertFoundOne(err) {
		return nil
	}
	return user
}

// CreateOrUpdateUser consolidates and returns the user matching the args
func CreateOrUpdateUser(ctx context.Context, githubID, googleID, email, nickname, name, photoURL string) (*User, error) {
	user := &User{}
	create := false

	tx, err := db.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback() // defer rollback on error

	// find user by ID
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
	if user.ReadQuota == 0 {
		user.ReadQuota = DefaultUserReadQuota
	}
	if user.WriteQuota == 0 {
		user.WriteQuota = DefaultUserWriteQuota
	}

	// set username
	usernameSeeds := user.usernameSeeds(nickname)
	if create {
		user.Username = usernameSeeds[0]
		if user.Name == "" {
			user.Name = user.Username
		}
	}

	// validate
	err = GetValidator().Struct(user)
	if err != nil {
		return nil, err
	}

	// insert or update
	if !create {
		user.UpdatedOn = time.Now()
		err = tx.Update(user)
		if err != nil {
			return nil, err
		}
	} else {
		// try out all username seeds
		for _, username := range usernameSeeds {
			// savepoint in case insert fails
			_, err = tx.Exec("SAVEPOINT bi")
			if err != nil {
				return nil, err
			}

			// insert
			user.Username = username
			err = tx.Insert(user)
			if err == nil {
				// success
				break
			} else if isUniqueUsernameError(err) {
				// rollback to before error, then try next username
				_, err = tx.Exec("ROLLBACK TO SAVEPOINT bi")
				if err != nil {
					return nil, err
				}
				continue
			} else {
				// unexpected error
				return nil, err
			}
		}

		// create a "personal" organization for the user
		org := &Organization{
			Name:     user.Username,
			Personal: true,
		}

		_, err := tx.Model(org).Insert()
		if err != nil {
			return nil, err
		}

		// update user.OrganizationID
		user.OrganizationID = &org.OrganizationID
		err = tx.Update(user)
		if err != nil {
			return nil, err
		}

		// add user to user-organization permissions table
		err = tx.Insert(&PermissionsUsersOrganizations{
			UserID:         user.UserID,
			OrganizationID: org.OrganizationID,
			View:           true,
			Admin:          true,
		})
		if err != nil {
			return nil, err
		}

		// create billing info and setup with anarchism payments driver
		bi := &BillingInfo{}
		bi.OrganizationID = org.OrganizationID
		bi.BillingPlanID = uuid.FromStringOrNil(FreeBillingPlanID)
		bi.PaymentsDriver = AnarchismDriver
		driverPayload := make(map[string]interface{})
		bi.DriverPayload = driverPayload

		_, err = tx.Model(bi).Insert()
		if err != nil {
			return nil, err
		}
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	// send "identify" call to Segment
	// userID should be the user.UserID, secretID or the browserID?
	// Segment.Client.Enqueue(analytics.Identify{
	// UserId: user.UserID,
	// Traits: analytics.NewTraits().
	//   Set("organization", user.OrganizationID),
	// })

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

// JoinOrganization changes the user's Organization
func (u *User) JoinOrganization(ctx context.Context, organizationID uuid.UUID) (*User, error) {
	// TODO: this whole function should happen in one transaction
	bi := FindBillingInfo(ctx, organizationID)
	if bi == nil {
		panic("could not find billing info for the new organization")
	}

	// commit current usage to the old organization's bill
	err := commitCurrentUsageToNextBill(ctx, *u.OrganizationID, UserEntityKind, u.UserID, false)
	if err != nil {
		return nil, err
	}

	// commit usage credit to the new organization's bill for the user's current month's usage
	err = commitCurrentUsageToNextBill(ctx, organizationID, UserEntityKind, u.UserID, true)
	if err != nil {
		return nil, err
	}

	u.OrganizationID = &organizationID
	u.ReadQuota = bi.BillingPlan.SeatReadQuota
	u.WriteQuota = bi.BillingPlan.SeatWriteQuota
	u.UpdatedOn = time.Now()

	_, err = db.DB.WithContext(ctx).Model(u).
		Column("organization_id", "read_quota", "write_quota", "updated_on").
		WherePK().
		Update()
	if err != nil {
		return nil, err
	}

	// add prorated seat to the new organization's next month's bill
	err = commitProratedSeatsToBill(ctx, bi, []uuid.UUID{u.UserID}, false)
	if err != nil {
		panic("unable to commit prorated seat to bill")
	}

	// after updating the organization_id, the new Organization table is not immediately available on the user object
	// so we need to refetch the user in order to get the new Organization details
	user := FindUser(ctx, u.UserID)
	if user == nil {
		panic("unable to get updated user")
	}

	return user, nil
}

// Delete removes the user from the database
func (u *User) Delete(ctx context.Context) error {
	err := commitCurrentUsageToNextBill(ctx, *u.OrganizationID, UserEntityKind, u.UserID, false)
	if err != nil {
		return err
	}
	return db.DB.WithContext(ctx).Delete(u)
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
	_, err = db.DB.ModelContext(ctx, u).Column("username", "name", "bio", "photo_url", "updated_on").WherePK().Update()
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

func isUniqueUsernameError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(strings.ToLower(err.Error()), "violates unique constraint \"users_username_key\"")
}
