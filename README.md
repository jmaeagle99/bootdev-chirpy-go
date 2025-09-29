# chirpy-go

## Motivation

This project was developed as part of a learning course on [boot.dev](https://www.boot.dev]) ([Learn HTTP servers in Go](https://www.boot.dev/courses/learn-http-servers-golang)) to get more hands-on practice with implementing basic REST HTTP servers in Go.

While reading documentation for frameworks and runtimes is great, I find that practical application and practice is how I best learn. If you want similar practice and learning, I would highly recommend taking this course.

## Developing

### Prerequisites

- Postgres: `sudo apt install postgresql postgresql-contrib`
- Goose: `go install github.com/pressly/goose/v3/cmd/goose@latest`
- SQLC: `go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest`

### Create Database

```bash
# If you haven't set up postgres previously
sudo passwd postgres

# Start psql shell
sudo -u postgres psql

# Create users and database
CREATE ROLE chirpy_owner WITH LOGIN PASSWORD '<owner_password>';
CREATE ROLE chirpy_user WITH LOGIN PASSWORD '<user_password>';
CREATE DATABASE chirpy OWNER chirpy_owner;

# Exit postgres psql session
exit

# Start chirpy_owner psql shell
psql "postgres://chirpy_owner:<owner_password>@localhost:5432/chirpy"

# Allow connect for user
GRANT CONNECT ON DATABASE chirpy TO chirpy_user;
# Allow schema usage by user
GRANT USAGE ON SCHEMA public TO chirpy_user;

# Allow CRUD to *existing* tables/sequences
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO chirpy_user;
GRANT USAGE, SELECT, UPDATE ON ALL SEQUENCES IN SCHEMA public TO chirpy_user;

# Allow CRUD to *future* tables/sequences
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO chirpy_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT USAGE, SELECT, UPDATE ON SEQUENCES TO chirpy_user;

# Exit chirpy_owner psql session
exit

# Start chirpy_user psql shell
psql "postgres://chirpy_user:<user_password>@localhost:5432/chirpy"

# Check basic SQL query for chirpy_user
SELECT * FROM users;

# Exit chirpy_user psql session
exit
```

### Run SQL migrations

Database must be set up before running. Run `goose ... up` to fill in database schema and connection information on first use before running the app.

#### Migrate Up

```bash
goose postgres "postgres://chirpy_owner:<owner_password>@localhost:5432/chirpy" -dir ./sql/schema up
```

#### Migrate Down

```bash
goose postgres "postgres://chirpy_owner:<owner_password>@localhost:5432/chirpy" -dir ./sql/schema down
```

### Build and Run

```bash
go build
./chirpy
```

### Update SQL Code generation

```bash
sqlc generate
```
