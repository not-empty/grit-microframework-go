# GRIT – Go REST Interface Toolkit

GRIT is a minimalist microservice framework built in pure Go. Designed to simplify RESTful API development without sacrificing performance, it draws inspiration from ALA (Automatic Lumen API) while embracing Go's philosophy of simplicity, concurrency, and clarity.

## Docker Execution
Use docker-compose to run the project
```bash
docker-compose up -d
```
Run the container:
```bash
docker exec -it grit sh
```

## Set the Environment
Copy the `.env.example` to a `.env` file to run API
```env
APP_ENV=local # if "local" the errors will be exposed in the body of the request
APP_LOG=true # if "true" the access log will be enabled in terminal
APP_NO_AUTH=true # if "true" will bypass authentication (not recommended in productio)
APP_PORT=8001 # port to use in runtime"

DB_DRIVER=mysql # only has been developed and tested with mysql
DB_HOST=grit-mysql # database url host
DB_MAX_CONN=100 # database max connections
DB_MAX_IDLE=100 # database max iddle connections
DB_NAME=grit # database name
DB_PASS=password  # database password
DB_PORT=3306 # databse port
DB_USER=user # database username

JWT_APP_SECRET=secret # JWT secret to token generation
JWT_EXPIRE=900 # JWT token expiration
JWT_RENEW=600 # JWT automatcly renewing time
```

## JWT Token Configurations
Copy the `./congig/tokens.json.example` to a `./config/tokens.json` file and change the tokens names, secrets and context inside (Higly recommended)

## Run the API
Run GRIT with:
```bash
run go main.go
```

## Request-ID and Profile Headers
Every request should include:
- `X-Request-ID`: unique request identifier (ULID)
- `X-Profile` profile to enable performance profiling

## JWT Authentication
To Generate a JWT token and authenticate in the API (change accord your token configurations):
```bash
curl -X POST 'http://localhost:APP_PORT/auth/generate' \
  -H 'Content-Type: application/json' \
  -H 'Context: <context>' \
  -d '{"token":"<token>","secret":"<secret>"}'
```

After generation, you'll receive a HTTP 204 Response Including these headers:
- `X-Token`: your JWT Token
- `X-Expires`: JWT token expiration datetime

## Authorization
In each request you will need to pass the Autorization Bearer and the Context HTTP headers:
```bash
curl 'http://localhost:${APP_PORT}/example/list' \
  -H 'Authorization: Bearer <jwt>' \
  -H 'Context: <context>' \
  -H 'Accept: application/json'
```
Every valid request will return a JWT Token and its expiration in the headers:
- `X-Token`: A valid JWT Token to do new requests.
- `X-Expires` The expiration datetime.
When your token is about to expire (according to you `.env` configurations), a valid request will always automaticly returns a renewed JWT and its new expiration.

## Endpoints
The API have an example model called `example` and its default endpoints:

| Method | Path                       | Description                           |
| ------ | -------------------------- | ------------------------------------- |
| POST   | /example/add               | Create a new record                   |
| POST   | /example/bulk              | List data for especific records       |
| GET    | /example/dead_detail/{id}  | Get data for especific deleted record |
| GET    | /example/dead_list         | List all deleted data (paginated)     |
| DELETE | /example/delete/{id}       | Delete a specific record              |
| GET    | /example/detail/{id}       | Get data for especific deleted record |
| PATCH  | /example/edit/{id}         | Update fields for a specific record   |
| GET    | /example/list              | List all data (paginated)             |

### Example: Add
```bash
curl -X POST http://localhost:${APP_PORT}/example/add \
  -H 'Content-Type: application/json' \
  -H 'X-Request-Id: <ulid>' \
  -H 'X-Token: <token>' \
  -d '{"name":"John","age":30}'
```

### Testing Example Domain
You can see and run all example endpoints in the `./ops/curl.sh`

## Creating a new Domains
To create a new domain first you have to save the MySQL DDL in the `./cmd/sql/{name}.sql`, you can use the `./cmd/sql/example.sql.example` as base.

The sql must have the fields id, created_at, updated_at and deleted_at and the types must be respected.

And then
```bash
cd cmd
go run domain.go -domain={name}
```
This generates:
- `app/repository/models/{name}_model.go`
- `app/router/domains/{name}_domain.go`
The generated files do not have logic so don't mess with test coverage.

## Generating Routes
Optionally you can creating a new generic route:
```bash
cd cmd
go run route.go -domain={name}
```
This generates:
- `app/router/routes/{name}_route.go`
- `app/controller/{name}_controller.go`
Note that this generated files have logic so they will impact in the test coverage

## Audit and Test Coverage
Check overall test coverage:
```bash
./audit.sh
```
Sample output:
```
🔍 Auditing Unit test coverage per package...

✅ github.com/not-empty/grit - 100.0%
✅ github.com/not-empty/grit/app/controller - 100.0%
✅ github.com/not-empty/grit/app/database - 100.0%
✅ github.com/not-empty/grit/app/helper - 100.0%
✅ github.com/not-empty/grit/app/middleware - 100.0%
✅ github.com/not-empty/grit/app/repository - 100.0%
✅ github.com/not-empty/grit/app/router - 100.0%
✅ github.com/not-empty/grit/app/router/registry - 100.0%
✅ github.com/not-empty/grit/app/util/jwt_manager - 100.0%
✅ github.com/not-empty/grit/app/util/ulid - 100.0%

📊 Total project coverage: 100.0%

🧪 View detailed HTML coverage report:
👉  ./tests/coverage/coverage-unit.html

```

## Running Unit Tests
Run all tests without coverage:
```bash
./test.sh
```

## Filtering Options
Use the `filter` query parameter to filter results:
```
?filter=field:operator:value
```
Supported operators:

| Operator  | Example                                           | SQL Clause                                         |
| --------- | ------------------------------------------------- | -------------------------------------------------- |
| `eq`      | `filter=age:eq:30`                                | `age = 30`                                         |
| `gt`      | `filter=age:gt:30`                                | `age > 30`                                         |
| `lt`      | `filter=age:lt:30`                                | `age < 30`                                         |
| `gte`     | `filter=age:gte:18`                               | `age >= 18`                                        |
| `lte`     | `filter=created_at:lte:2024-12-31`                | `created_at <= '2024-12-31'`                       |
| `like`    | `filter=email:like:gmail.com`                     | `email LIKE '%gmail.com%'`                         |
| `isnull`  | `filter=deleted_at:isnull:true`                   | `deleted_at IS NULL`                               |
| `between` | `filter=created_at:between:2024-01-01,2024-12-31` | `created_at BETWEEN '2024-01-01' AND '2024-12-31'` |

---
For now, this covers the basic usage of Grit. Feel free to suggest improvements!

