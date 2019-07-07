package model

import (
	"time"

	uuid "github.com/satori/go.uuid"
	"gopkg.in/go-playground/validator.v9"
)

// configure validator
func init() {
	GetValidator().RegisterStructValidation(keyValidation, Key{})
}

// custom key validation
func keyValidation(sl validator.StructLevel) {
	k := sl.Current().Interface().(Key)

	if k.UserID == uuid.Nil && k.ProjectID == uuid.Nil {
		sl.ReportError(k.ProjectID, "ProjectID", "", "projectid_userid_empty", "")
	}

	if k.Role != "r" && k.Role != "rw" && k.Role != "m" {
		sl.ReportError(k.Role, "Role", "", "bad_role", "")
	}
}

// export enum KeyRole {
//     Readonly = "r",
//     Readwrite = "rw",
//     Manage = "m",
// }

// Key represents an access token to read/write data or a user session token
type Key struct {
	KeyID       uuid.UUID `sql:",pk,type:uuid"`
	Description string    `validate:"omitempty,lte=32"`
	Prefix      string    `sql:",notnull",validate:"required,len=4"`
	HashedKey   string    `sql:",unique,notnull",validate:"required,lte=64"`
	Role        string    `sql:",notnull",validate:"required,lte=3"`
	UserID      uuid.UUID `sql:"on_delete:CASCADE,type:uuid"`
	User        *User
	ProjectID   uuid.UUID `sql:"on_delete:CASCADE,type:uuid"`
	Project     *Project
	CreatedOn   time.Time `sql:",default:now()"`
	UpdatedOn   time.Time `sql:",default:now()"`
	KeyString   string    `sql:"-"`
}

// GenerateKeyString returns a random generated key string
func GenerateKeyString() string {
	// TODO
	// const arrayBuffer = new Array();
	// uuidv4(null, arrayBuffer, 0);
	// uuidv4(null, arrayBuffer, 16);
	// const buffer = new Buffer(arrayBuffer);
	// return buffer.toString("base64");
	return ""
}

// HashKeyString safely hashes keyString
func HashKeyString(keyString string) string {
	// TODO
	// return crypto.createHash("sha256").update(new Buffer(key, "base64")).digest().toString("base64");
	return ""
}

// NewKey creates a new, unconfigured key -- use NewUserKey or NewProjectKey instead
func NewKey() *Key {
	// TODO
	// const key = new Key();
	//   key.keyString = Key.generateKey();
	//   key.hashedKey = Key.hashKey(key.keyString);
	//   key.prefix = key.keyString.slice(0, 4);
	//   return key;
	return nil
}

// NewUserKey creates a new key to manage a user
func NewUserKey(userID uuid.UUID, role string, description string) *Key {
	// TODO
	// const key = this.newKey();
	//     key.description = description;
	//     key.role = role;
	//     key.user = new User();
	//     key.user.userId = userId;
	//     await key.save();
	//     return key;
	return nil
}

// NewProjectKey creates a new read or readwrite key for a project
func NewProjectKey(projectID uuid.UUID, role string, description string) *Key {
	// TODO
	// const key = this.newKey();
	//   key.description = description;
	//   key.role = role;
	//   key.project = new Project();
	//   key.project.projectId = projectId;
	//   await key.save();
	//   return key;
	return nil
}

// AuthenticateKeyString returns the key object matching keyString (if there's a match)
func AuthenticateKeyString(keyString string) *Key {
	// const hashedKey = this.hashKey(keyString);
	// const key: Key = await this.findOne({ hashedKey }, {
	//   cache: { id: `key:${hashedKey}`, milliseconds: 3600000 }
	// });
	// if (!key) {
	//   return null;
	// }
	// return key;
	return nil
}

// Revoke deletes the key
func (k *Key) Revoke() {
	// await this.remove();
	//   const cache = getConnection().queryResultCache;
	//   if (cache) {
	//     await cache.remove([`key:${this.hashedKey}`]);
	//   }
}
