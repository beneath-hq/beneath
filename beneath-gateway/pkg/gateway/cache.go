package gateway

import (
	"fmt"
	"time"

	"github.com/beneath-core/beneath-gateway/pkg/beneath"
	"github.com/go-redis/cache"
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

func lookupCurrentInstanceID(projectName string, streamName string) (string, error) {
	instanceID := ""
	err := codec.Once(&cache.Item{
		Key:        fmt.Sprintf("gw:iid:%s:%s", projectName, streamName),
		Object:     &instanceID,
		Expiration: cacheTime,
		Func: func() (interface{}, error) {
			// TODO
			return "temp", nil
		},
	})
	return instanceID, NewHTTPError(500, err.Error())
}

type cachedInstance struct {
	public     bool
	manual     bool
	projectID  string
	avroSchema string
}

func lookupInstance(instanceID string) (*cachedInstance, error) {
	res := &cachedInstance{}
	err := codec.Once(&cache.Item{
		Key:        fmt.Sprintf("gw:i:%s", instanceID),
		Object:     &res,
		Expiration: cacheTime,
		Func: func() (interface{}, error) {
			// TODO
			return &cachedInstance{true, true, "hey", "{}"}, nil
		},
	})
	return res, NewHTTPError(500, err.Error())
}

type cachedRole struct {
	read   bool
	write  bool
	manage bool
}

func lookupRole(auth string, inst *cachedInstance) (*cachedRole, error) {
	res := &cachedRole{}
	err := codec.Once(&cache.Item{
		Key:        fmt.Sprintf("gw:r:%s:%s", auth, inst.projectID),
		Object:     &res,
		Expiration: cacheTime,
		Func: func() (interface{}, error) {
			// TODO
			return &cachedRole{true, false, false}, nil
		},
	})
	return res, NewHTTPError(500, err.Error())
}
