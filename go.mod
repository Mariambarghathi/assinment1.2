module backend-project

go 1.22.0

toolchain go1.22.6

require github.com/go-michi/michi v0.0.1

require github.com/joho/godotenv v1.5.1

require github.com/jmoiron/sqlx v1.4.0

require (
	github.com/Masterminds/squirrel v1.5.4
	github.com/golang-migrate/migrate/v4 v4.17.1
	github.com/google/uuid v1.6.0
	github.com/lib/pq v1.10.9
	golang.org/x/crypto v0.27.0
)

require (
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/lann/builder v0.0.0-20180802200727-47ae307949d0 // indirect
	github.com/lann/ps v0.0.0-20150810152359-62de8c46ede0 // indirect
	go.uber.org/atomic v1.7.0 // indirect
)
