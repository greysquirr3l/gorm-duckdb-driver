module debug

go 1.24

toolchain go1.24.4

require (
	gorm.io/driver/duckdb v0.0.0
	gorm.io/gorm v1.25.12
)

require (
	github.com/apache/arrow-go/v18 v18.4.0 // indirect
	github.com/duckdb/duckdb-go-bindings v0.1.17 // indirect
	github.com/duckdb/duckdb-go-bindings/darwin-amd64 v0.1.12 // indirect
	github.com/duckdb/duckdb-go-bindings/darwin-arm64 v0.1.12 // indirect
	github.com/duckdb/duckdb-go-bindings/linux-amd64 v0.1.12 // indirect
	github.com/duckdb/duckdb-go-bindings/linux-arm64 v0.1.12 // indirect
	github.com/duckdb/duckdb-go-bindings/windows-amd64 v0.1.12 // indirect
	github.com/go-viper/mapstructure/v2 v2.4.0 // indirect
	github.com/goccy/go-json v0.10.5 // indirect
	github.com/google/flatbuffers v25.2.10+incompatible // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/klauspost/cpuid/v2 v2.3.0 // indirect
	github.com/marcboeker/go-duckdb/arrowmapping v0.0.10 // indirect
	github.com/marcboeker/go-duckdb/mapping v0.0.11 // indirect
	github.com/marcboeker/go-duckdb/v2 v2.3.3 // indirect
	github.com/pierrec/lz4/v4 v4.1.22 // indirect
	github.com/zeebo/xxh3 v1.0.2 // indirect
	golang.org/x/exp v0.0.0-20250718183923-645b1fa84792 // indirect
	golang.org/x/mod v0.26.0 // indirect
	golang.org/x/sync v0.16.0 // indirect
	golang.org/x/sys v0.34.0 // indirect
	golang.org/x/text v0.27.0 // indirect
	golang.org/x/tools v0.35.0 // indirect
	golang.org/x/xerrors v0.0.0-20240903120638-7835f813f4da // indirect
)

replace gorm.io/driver/duckdb => ../
