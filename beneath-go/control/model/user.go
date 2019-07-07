package model

import (
	"regexp"
	"time"

	uuid "github.com/satori/go.uuid"
	"gopkg.in/go-playground/validator.v9"
)

/*
indexes:
- IDX_UQ_USERS_USERNAME unique on username
- IDX_UQ_USERS_EMAIL unique on email
- IDX_UQ_USERS_GOOGLE_ID google_id
- IDX_UQ_USERS_GITHUB_ID github_id
- IDX_UQ_PROJECTS_NAME name
- IDX_UQ_KEYS_HASHED_KEY
- @Index("IDX_UQ_STREAMS_NAME_PROJECT", ["project", "name"], { unique: true })
*/

// constants
var (
	userUsernameRegex *regexp.Regexp
)

// configure constants and validator
func init() {
	userUsernameRegex = regexp.MustCompile("^[_a-z][_\\-a-z0-9]*$")
	GetValidator().RegisterStructValidation(userValidation, User{})
}

// User represents a Beneath user
type User struct {
	UserID    uuid.UUID  `sql:",pk,type:uuid"`
	Username  string     `validate:"omitempty,gte=3,lte=16"`
	Email     string     `sql:",unique,notnull",validate:"required,email"`
	Name      string     `sql:",unique,notnull",validate:"required,gte=4,lte=50"`
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
	// TODO
	return nil
	//     return await getConnection()
	//       .createQueryBuilder(User, "user")
	//       .where("lower(user.email) = lower(:email)", { email })
	//       .getOne();
}

// CreateOrUpdateUser consolidates and returns the user matching the args
func CreateOrUpdateUser(githubID, googleID, email, name, photoURL string) *User {
	// TODO:
	//     let user = null;
	//     let created = false;
	//     if (githubId) {
	//       user = await User.findOne({ githubId });
	//     } else if (googleId) {
	//       user = await User.findOne({ googleId });
	//     }
	//     if (!user) {
	//       user = await User.findOne({ email });
	//     }
	//     if (!user) {
	//       user = new User();
	//       created = true;
	//     }

	//     user.githubId = user.githubId || githubId;
	//     user.googleId = user.googleId || googleId;
	//     user.email = email;
	//     user.name = name;
	//     user.photoUrl = photoUrl;

	//     await user.save();
	//     if (created) {
	//       logger.info(`Created userId <${user.userId}>`);
	//     } else {
	//       logger.info(`Updated userId <${user.userId}>`);
	//     }
	//     return user;
	return nil
}
