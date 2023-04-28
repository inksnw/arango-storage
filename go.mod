module github.com/clusterpedia-io/arango-storage-layer

go 1.19

require (
	github.com/arangodb/go-driver v1.5.2
	github.com/clusterpedia-io/api v0.0.0
	github.com/clusterpedia-io/clusterpedia v0.0.0-00010101000000-000000000000
	github.com/inksnw/gorm-arango v0.1.4
	github.com/jinzhu/configor v1.2.1
	gopkg.in/natefinch/lumberjack.v2 v2.0.0
	gorm.io/datatypes v1.0.7
	gorm.io/gorm v1.24.7-0.20230306060331-85eaf9eeda11
	k8s.io/api v0.25.3
	k8s.io/apimachinery v0.25.3
	k8s.io/apiserver v0.25.3
	k8s.io/component-base v0.25.3
	k8s.io/klog/v2 v2.70.1
)

require (
	github.com/BurntSushi/toml v0.3.1 // indirect
	github.com/arangodb/go-velocypack v0.0.0-20200318135517-5af53c29c67e // indirect
	github.com/cenkalti/backoff/v4 v4.2.0 // indirect
	github.com/docker/distribution v2.8.1+incompatible // indirect
	github.com/go-logr/logr v1.2.3 // indirect
	github.com/go-sql-driver/mysql v1.6.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/google/gofuzz v1.1.0 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/sirupsen/logrus v1.9.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	golang.org/x/net v0.0.0-20220722155237-a158d28d115b // indirect
	golang.org/x/sys v0.5.0 // indirect
	golang.org/x/text v0.3.8 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gorm.io/driver/mysql v1.4.4 // indirect
	k8s.io/apiextensions-apiserver v0.25.2 // indirect
	k8s.io/kubernetes v1.25.2 // indirect
	k8s.io/utils v0.0.0-20220728103510-ee6ede2d64ed // indirect
	sigs.k8s.io/json v0.0.0-20220713155537-f223a00ba0e2 // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.2.3 // indirect
	sigs.k8s.io/yaml v1.3.0 // indirect
)

replace (
	github.com/clusterpedia-io/api => ./clusterpedia/staging/src/github.com/clusterpedia-io/api
	github.com/clusterpedia-io/clusterpedia => ./clusterpedia
)
