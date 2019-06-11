package gateway

import (
	"fmt"
	"time"

	"github.com/beneath-core/beneath-gateway/beneath"
	"github.com/go-pg/pg"
	"github.com/go-redis/cache"
	uuid "github.com/satori/go.uuid"
	"github.com/vmihailenco/msgpack"
)

const (
	cacheTime      = time.Hour
	cacheLRULength = 1000
	cacheLRUTime   = 10 * time.Minute
)

var (
	codec *cache.Codec
)

func init() {
	codec = &cache.Codec{
		Redis: beneath.Redis,
		Marshal: func(v interface{}) ([]byte, error) {
			return msgpack.Marshal(v)
		},
		Unmarshal: func(b []byte, v interface{}) error {
			return msgpack.Unmarshal(b, v)
		},
	}
	codec.UseLocalCache(cacheLRULength, cacheLRUTime)
}

func lookupCurrentInstanceID(projectName string, streamName string) (uuid.UUID, error) {
	var instanceID uuid.UUID
	err := codec.Once(&cache.Item{
		Key:        fmt.Sprintf("gw:iid:%s:%s", projectName, streamName),
		Object:     &instanceID,
		Expiration: cacheTime,
		Func: func() (interface{}, error) {
			res := ""
			_, err := beneath.DB.Query(pg.Scan(&res), `
				select s.current_stream_instance_id
				from streams s
				join projects p on s.project_id = p.project_id
				where s.name = lower(?) and p.name = lower(?)
			`, streamName, projectName)
			return res, err
		},
	})
	if err != nil {
		return uuid.Nil, NewHTTPError(500, err.Error())
	}
	if instanceID == uuid.Nil {
		return uuid.Nil, NewHTTPError(404, "stream not found")
	}
	return instanceID, nil
}

type cachedInstance struct {
	Public     bool
	Manual     bool
	ProjectID  uuid.UUID
	AvroSchema string
}

func lookupInstance(instanceID uuid.UUID) (*cachedInstance, error) {
	instance := &cachedInstance{}
	err := codec.Once(&cache.Item{
		Key:        fmt.Sprintf("gw:i:%s", instanceID),
		Object:     instance,
		Expiration: cacheTime,
		Func: func() (interface{}, error) {
			res := &cachedInstance{}
			_, err := beneath.DB.Query(res, `
				select p.public, s.manual, s.project_id, s.avro_schema
				from stream_instances si
				join streams s on si.stream_id = s.stream_id
				join projects p on s.project_id = p.project_id
				where si.stream_instance_id = ?
			`, instanceID)
			return res, err
		},
	})
	if err != nil {
		return nil, NewHTTPError(500, err.Error())
	}
	if instance.ProjectID == uuid.Nil {
		return nil, NewHTTPError(404, "instance not found")
	}
	return instance, nil
}

type cachedRole struct {
	read   bool
	write  bool
	manage bool
}

func lookupRole(auth string, inst *cachedInstance) (*cachedRole, error) {
	res := &cachedRole{}
	err := codec.Once(&cache.Item{
		Key:        fmt.Sprintf("gw:r:%s:%s", auth, inst.ProjectID),
		Object:     &res,
		Expiration: cacheTime,
		Func: func() (interface{}, error) {
			// TODO
			return &cachedRole{true, false, false}, nil
		},
	})
	if err != nil {
		return nil, NewHTTPError(500, err.Error())
	}
	return res, nil
}

// // read from database when redis returns empty
// if err == redis.Nil {
// 	// query
// 	_, err := beneath.DB.QueryOne(pg.Scan(&role), `
// 		select coalesce(
// 			-- check key is project with access
// 			(select k.role from keys k where k.hashed_key = ? and k.project_id = p.project_id),
// 			-- check key is user with access
// 			(select if() from keys k join projects_users pu on k.user_id = pu.user_id where k.hashed_key = ? and k.user_id is not null and p.project_id = pu.project_id)
// 			-- check key exists
// 			-- nothing
// 			, ""
// 		) as role
// 		from projects p
// 		where p.name = ?
// 	`
// 		`SELECT ?, ?`,
// 		"foo", "bar"
// 	)

// 	// return empty if key doesn't exist
// 	// return "r" if key is
// 	// return "rw" if key is rw or m and in project or manually editable
// 	// select
// 	// 	coalesce(
// 	// 		(
// 	// 			select
// 	// 				k.role
// 	// 			from keys k
// 	// 			where
// 	// 				k.hashed_key = ?
// 	// 				and k.project_id = p.project_id
// 	// 		),
// 	// 		(
// 	// 			select
// 	// 				k.project_id = pu.project_id
// 	// 			from keys k
// 	// 			join projects_users pu on k.user_id = pu.user_id
// 	// 			where
// 	// 				k.hashed_key = ?
// 	// 				and k.user_id is not null
// 	// 		),
// 	// 		false
// 	// 	) as
// 	// 	(

// 	// 	)
// 	// 	(select k.project_id = p.project_id from keys k  where k.hashed_key = ?)
// 	// from projects p
// 	// where p.name = ?

// 	if err != nil {
// 		return "", err
// 	}

// 	// update role
// 	role = res[0].Role
