package gateway

import (
	"errors"
	"fmt"
	"log"
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
			res := uuid.Nil
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
		log.Panic("lookupCurrentInstanceID error: %v", err)
	}
	if instanceID == uuid.Nil {
		return uuid.Nil, errors.New("stream not found")
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
		log.Panic("lookupInstance error: %v", err)
	}
	if instance.ProjectID == uuid.Nil {
		return nil, errors.New("instance not found")
	}
	return instance, nil
}

type cachedRole struct {
	Read   bool
	Write  bool
	Manage bool
}

func lookupRole(auth string, inst *cachedInstance) (*cachedRole, error) {
	res := &cachedRole{}
	err := codec.Once(&cache.Item{
		Key:        fmt.Sprintf("gw:r:%s:%s", auth, inst.ProjectID),
		Object:     &res,
		Expiration: cacheTime,
		Func: func() (interface{}, error) {
			res := ""
			_, err := beneath.DB.Query(pg.Scan(&res), `
				select coalesce(
					(select k.role
						from keys k
						where k.hashed_key = ?0 and k.project_id = ?1
					),
					(select 'm' as role
						from keys k
						join projects_users pu on k.user_id is not null and k.user_id = pu.user_id
						where k.hashed_key = ?0 and pu.project_id = ?1
					),
					(select '-' as role
						from keys k
						where k.hashed_key = ?0
					),
					''
				)
			`, auth, inst.ProjectID)

			if err != nil {
				return nil, err
			}

			return &cachedRole{
				Read:   res == "r" || res == "rw" || (res == "-" && inst.Public),
				Write:  res == "rw",
				Manage: res == "m",
			}, nil
		},
	})
	if err != nil {
		log.Panic("lookupRole error: %v", err)
	}
	return res, nil
}
