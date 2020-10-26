module gitlab.com/beneath-hq/beneath

go 1.12

require (
	cloud.google.com/go v0.65.0
	cloud.google.com/go/bigquery v1.8.0
	cloud.google.com/go/bigtable v1.6.0
	cloud.google.com/go/pubsub v1.3.1
	github.com/99designs/gqlgen v0.13.0
	github.com/alecthomas/participle v0.3.0
	github.com/bluele/gcache v0.0.0-20190518031135-bc40bd653833
	github.com/go-chi/chi v4.0.2+incompatible
	github.com/go-pg/migrations/v7 v7.0.0
	github.com/go-pg/pg v8.0.7+incompatible
	github.com/go-pg/pg/v9 v9.0.0-beta
	github.com/go-playground/locales v0.12.1 // indirect
	github.com/go-playground/universal-translator v0.16.0 // indirect
	github.com/go-redis/cache/v7 v7.0.0
	github.com/go-redis/redis/v7 v7.0.0-beta.4
	github.com/go-redis/redis_rate/v8 v8.0.0-beta
	github.com/go-test/deep v1.0.1
	github.com/gogo/protobuf v1.2.1
	github.com/golang-collections/collections v0.0.0-20130729185459-604e922904d3
	github.com/golang/protobuf v1.4.2
	github.com/google/btree v1.0.0 // indirect
	github.com/googleapis/gax-go/v2 v2.0.5 // indirect
	github.com/gorilla/sessions v1.1.1
	github.com/gorilla/websocket v1.4.2
	github.com/grpc-ecosystem/go-grpc-middleware v1.0.0
	github.com/hashicorp/go-version v1.2.0
	github.com/iancoleman/strcase v0.0.0-20190422225806-e506e3ef7365
	github.com/jackc/pgproto3/v2 v2.0.2 // indirect
	github.com/jinzhu/inflection v1.0.0
	github.com/joho/godotenv v1.3.0
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/leodido/go-urn v1.1.0 // indirect
	github.com/linkedin/goavro/v2 v2.8.5
	github.com/markbates/goth v1.64.2
	github.com/mitchellh/mapstructure v1.2.2
	github.com/mr-tron/base58 v1.1.2
	github.com/rs/cors v1.6.0
	github.com/satori/go.uuid v1.2.0
	github.com/segmentio/ksuid v1.0.2
	github.com/sendgrid/rest v2.6.1+incompatible // indirect
	github.com/sendgrid/sendgrid-go v3.6.4+incompatible
	github.com/spf13/cobra v1.1.0
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.7.1
	github.com/stretchr/testify v1.4.0
	github.com/stripe/stripe-go v66.1.1+incompatible
	github.com/vektah/gqlparser v1.1.2
	github.com/vektah/gqlparser/v2 v2.1.0
	github.com/vmihailenco/msgpack v4.0.4+incompatible
	github.com/xtgo/uuid v0.0.0-20140804021211-a0b114877d4c
	go.uber.org/fx v1.13.1
	go.uber.org/zap v1.10.0
	golang.org/x/sync v0.0.0-20200625203802-6e8e738ad208
	google.golang.org/api v0.31.0
	google.golang.org/grpc v1.31.1
	gopkg.in/go-playground/assert.v1 v1.2.1 // indirect
	gopkg.in/go-playground/validator.v9 v9.29.0
	robpike.io/filter v0.0.0-20150108201509-2984852a2183
)

replace github.com/linkedin/goavro/v2 => github.com/bem7/goavro/v2 v2.0.0-20191009165622-2e928607d532
