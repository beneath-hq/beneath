package user

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-pg/pg/v9/orm"

	"gitlab.com/beneath-hq/beneath/infrastructure/db"
	"gitlab.com/beneath-hq/beneath/models"
)

// CreateOrUpdateUser consolidates and returns the user matching the args
func (s *Service) CreateOrUpdateUser(ctx context.Context, githubID, googleID, email, nickname, name, photoURL string) (*models.User, error) {
	var createdUser *models.User
	err := s.DB.InTransaction(ctx, func(ctx context.Context) error {
		tx := s.DB.GetDB(ctx)

		found, user, org, err := s.getUserForCreateOrUpdate(ctx, githubID, googleID, email)
		if err != nil {
			return err
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
		if found {
			// validate and save user
			user.UpdatedOn = time.Now()
			err = user.Validate()
			if err != nil {
				return err
			}
			err = tx.Update(user)
			if err != nil {
				return err
			}

			// validate and save org
			org.UpdatedOn = time.Now()
			err = org.Validate()
			if err != nil {
				return err
			}
			err = tx.Update(org)
			if err != nil {
				return err
			}

			// publish event and commit
			err := s.Bus.Publish(ctx, models.UserUpdatedEvent{
				User: user,
			})
			if err != nil {
				return err
			}

			// commit
			return nil
		}

		// we're creating a new user

		// try out all possible usernames
		usernameSeeds := usernameSeeds(email, name, nickname)
		err = s.createWithUsernameSeeds(ctx, user, org, usernameSeeds)
		if err != nil {
			return err
		}

		// add user to user-organization permissions table
		err = tx.Insert(&models.PermissionsUsersOrganizations{
			UserID:         user.UserID,
			OrganizationID: org.OrganizationID,
			View:           true,
			Create:         true,
			Admin:          true,
		})
		if err != nil {
			return err
		}

		// send events
		err = s.Bus.Publish(ctx, &models.UserCreatedEvent{
			User: user,
		})
		if err != nil {
			return err
		}

		err = s.Bus.Publish(ctx, &models.OrganizationCreatedEvent{
			Organization: org,
		})
		if err != nil {
			return err
		}

		// commit
		return nil
	})
	if err != nil {
		return nil, err
	}

	return createdUser, nil
}

func (s *Service) getUserForCreateOrUpdate(ctx context.Context, githubID, googleID, email string) (bool, *models.User, *models.Organization, error) {
	tx := s.DB.GetDB(ctx) // must be run in a tx
	org := &models.Organization{}
	user := &models.User{}
	found := true

	// create query to find by id
	var userQuery *orm.Query
	if githubID != "" {
		userQuery = tx.Model(user).Where("github_id = ?", githubID)
	} else if googleID != "" {
		userQuery = tx.Model(user).Where("google_id = ?", googleID)
	} else {
		return false, nil, nil, fmt.Errorf("Neither githubID nor googleID set for user")
	}

	// select for update
	err := userQuery.For("UPDATE").Select()
	if !db.AssertFoundOne(err) {
		// find by email instead
		err = tx.Model(user).Where("lower(email) = lower(?)", email).For("UPDATE").Select()
		if !db.AssertFoundOne(err) {
			found = false
		}
	}

	// if user found, select personal org for update
	if found {
		err = tx.Model(org).Where("user_id = ?", user.UserID).For("UPDATE").Select()
		if err != nil {
			return false, nil, nil, fmt.Errorf("Couldn't find personal organization for user: %s", err.Error())
		}
	}

	// if not found, set quota epoch
	if !found {
		org.QuotaEpoch = time.Now()
		user.QuotaEpoch = org.QuotaEpoch
	}

	// done
	return found, user, org, nil
}

func (s *Service) createWithUsernameSeeds(ctx context.Context, user *models.User, org *models.Organization, usernameSeeds []string) error {
	tx := s.DB.GetDB(ctx) // must be run in a tx
	created := false
	for _, username := range usernameSeeds {
		// update org and user
		org.Name = username

		// validate org
		err := org.Validate()
		if err != nil {
			// continue (probably username is blacklisted)
			continue
		}

		// savepoint in case name is taken
		_, err = tx.Exec("SAVEPOINT bi")
		if err != nil {
			return err
		}

		// insert org
		err = tx.Insert(org)
		if err == nil {
			// insert user (with org as billing)
			user.BillingOrganizationID = org.OrganizationID
			err = tx.Insert(user)
			if err != nil {
				return err
			}

			// update org with user ID
			org.UserID = &user.UserID
			_, err = tx.Model(org).Column("user_id").WherePK().Update()
			if err != nil {
				return err
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
				return err
			}
			continue
		}

		// unexpected error
		return err
	}

	// if not created, error
	if !created {
		return fmt.Errorf("couldn't create user for any username seeds")
	}

	return nil
}

func isUniqueError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(strings.ToLower(err.Error()), "violates unique constraint")
}
