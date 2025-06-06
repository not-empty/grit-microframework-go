# GRIT – Go REST Interface Toolkit

GRIT is a minimalist microservice framework built in pure Go. Designed to simplify RESTful API development without sacrificing performance, it draws inspiration from ALA while embracing Go's philosophy of simplicity, concurrency, and clarity.

---

## Quickstart

### 1. Clone

```bash
git clone https://github.com/not-empty/grit.git
```

### 2. Docker Compose

Bring up the service and its MySQL dependency:

```bash
docker-compose up -d
```

Enter the running container:

```bash
docker exec -it grit sh
```

### 3. Environment

Copy `.env.example` to `.env` and adjust:

```env
APP_ENV=local           # expose errors in response when "local"
APP_LOG=true            # enable HTTP access logs
APP_NO_AUTH=true        # disable auth (not for production)
APP_PORT=8001           # HTTP port

DB_DRIVER=mysql         # only MySQL supported
DB_HOST=grit-mysql
DB_NAME=grit
DB_USER=user
DB_PASS=password
DB_PORT=3306
DB_MAX_CONN=100
DB_MAX_IDLE=10

DB_HOST_TEST=grit-mysql
DB_NAME_TEST=grit
DB_PASS_TEST=password
DB_PORT_TEST=3306
DB_USER_TEST=user

JWT_APP_SECRET=secret   # JWT signing secret
JWT_EXPIRE=900          # expiration seconds
JWT_RENEW=600           # auto-renew threshold seconds
```

Also copy `./config/tokens.json.example` → `./config/tokens.json` to configure valid tokens and contexts.

---

## Run the API

Run GRIT with:

```bash
run go main.go
```

## Endpoints

| Method | Path                        | Description                                |
| ------ | --------------------------- | -------------------------------------------|
| POST   | `/example/add`              | Create a new record                        |
| POST   | `/example/bulk`             | Fetch specific records by IDs              |
| POST   | `/example/bulk_add`         | Create up to 25 records in the same request|
| GET    | `/example/dead_detail/{id}` | Get a deleted record by ID                 |
| GET    | `/example/dead_list`        | List deleted records (paginated)           |
| DELETE | `/example/delete/{id}`      | Soft-delete a record by ID                 |
| GET    | `/example/detail/{id}`      | Get an active record by ID                 |
| PATCH  | `/example/edit/{id}`        | Update specific fields                     |
| GET    | `/example/list`             | List active records (paginated)            |
| GET    | `/example/list_one`         | List one record based on params            |
| POST   | `/example/select_raw`       | Execute a predefined raw SQL query safely  |

---

## Authentication & Authorization

1. **Generate a JWT**

   ```bash
   curl -i -X POST http://localhost:$APP_PORT/auth/generate \
     -H "Content-Type: application/json" \
     -H "Context: <your-context>" \
     -d '{"token":"<token>","secret":"<secret>"}'
   ```

   On success you get HTTP 204 with headers:

   - `X-Token`: JWT
   - `X-Expires`: expiration timestamp

2. **Make API calls**
   ```bash
   curl -i GET http://localhost:$APP_PORT/example/list \
     -H "Authorization: Bearer <JWT>" \
     -H "Context: <your-context>" \
     -H "Accept: application/json"
   ```
   Every valid response return your valid token or renews it if is needed, always in the headers:
   - `X-Token`: new JWT
   - `X-Expires`: new expiration

---

## Request & Response Headers

| Header          | Description                   |
| --------------- | ----------------------------- |
| `X-Request-ID`  | Unique ULID for the request   |
| `X-Profile`     | Profiling timer (seconds)     |
| `X-Token`       | JWT token (on auth or renew)  |
| `X-Expires`     | JWT expiration timestamp      |
| `X-Page-Cursor` | Cursor for next page (string) |

---

## Pagination with Cursor

By default endpoints return up to **25** items and include an `X-Page-Cursor` header when more pages exist.

1. **First page** (no cursor):

   ```bash
   curl -i GET "http://localhost:$APP_PORT/example/list" \
     -H "Authorization: Bearer <JWT>"
   ```

   ```http
   HTTP/1.1 200 OK
   X-Page-Cursor: eyJsYXN0X2lkIjo...   # opaque cursor
   Content-Type: application/json
   [
     {"id":"1","name":"Alice"},
     ...
   ]
   ```

2. **Next page**:
   ```bash
   curl -i GET "http://localhost:$APP_PORT/example/list?page_cursor=<cursor>" \
     -H "Authorization: Bearer <JWT>"
   ```

Once fewer than **25** records return, no `X-Page-Cursor` is emitted (end of list).

---

## Ordering & Field Selection

Works on list and list_one endpoints

- **Order** by any column:
  `?order_by=name&order=desc`

- **Select fields**: the default is all the fields
  `?fields=id,name,created_at`

Example:

```bash
curl -i GET "http://localhost:$APP_PORT/example/list?order_by=age&order=asc&fields=id,name" \
  -H "Authorization: Bearer <JWT>"
```

---

## Filtering

Works on list and list_one endpoints

Use `filter` params:

```
?filter=age:eql:30&filter=name:lik:John
```

Supported operators:

- `eql` → `=`
- `neq` → `!=`
- `lik` → `LIKE` (contains)
- `gt` → `>`
- `lt` → `<`
- `gte` → `>=`
- `lte` → `<=`
- `btw` → `BETWEEN` (value1,value2)
- `nul` → `IS NULL`
- `nnu` → `IS NOT NULL`
- `in` → `IN` (comma list)

---

## Raw Selects

Allows execution of pre-registered raw SQL queries with named parameters. Queries must be registered in your model.

1 - Register queries in app/repository/models/<domain>_raw.go:
```golang
package models

import "github.com/not-empty/grit/app/helper"

func init() {
    helper.RegisterRawQueries("example", map[string]string{
        // key is query name, value is SQL template
        "count_active": `
          SELECT COUNT(1) AS total
          FROM example
          WHERE age = :age
        `,
    })
}
```

2 - Request format:

```bash
POST /example/select_raw
Content-Type: application/json

{
  "query": "count_active",
  "params": {
    "age" : 22
  }
}
```

3 - Response:

[200 OK] with JSON array of rows (each row is an object)

[400 Bad Request] if query is unknown, parameters mismatch, or contains forbidden terms

[500 Internal Server Error] on execution errors

4 - Limitations & rules:

Allowed: only SELECT or WITH statements

Denied substrings: ;, --, /*, */

Denied keywords: drop, alter, truncate, delete, update, insert, create, merge, replace, grant, revoke, commit, rollback, savepoint, lock, unlock, exec, call, use, set, limit, offset, join

All named parameters (e.g. :id) in the query must be provided in the params object, and no extra parameters are allowed.

Maximum rows returned is 25 (hard-coded).

## Generators

- **New Domain** (with DDL in `./cmd/sql/{name}.sql`):
  ```bash
  cd cmd/domain
  go run main.go -domain=name
  ```

Generated files:

- `app/repository/models/{name}_model.go`
- `app/repository/models/{name}_raw.go`
- `app/router/domains/{name}_domain.go`

> Generated code for new domains are test free since they are abstract of the basic implementations.

- **Generate Route**:
  ```bash
  cd cmd/route
  go run main.go -route=name
  ```

Generated files:

- `app/controller/{name}_controller.go`
- `app/router/routes/{name}_router.go`

> Generated code for new routes will counts toward coverage—tests since they are new logic.

## Adding a Record with manual ID

You can add a new record by sending a `POST` request to `/example/add`. By default, the API will generate a unique ID automatically.

However, if you prefer to use a custom ID, you can include the `id` field in the request body. In that case, the API will use the provided ID and skip the automatic ID generation.

## Validation

You can add validation in fields including the validation statement in models or in the fields comments in the DDL file before generating the domain:

Either way, you need to use the validate statement from https://github.com/go-playground/validator and its options.

> You can add or change in the model just including or editing the validate statement and the choosed options on the selected fields:

```golang
type Example struct {
	ID        string           `json:"id"`
	Name      string           `json:"name" validate:"required,min=5"`
	Age       int              `json:"age" validate:"required,number,gt=0,lt=100"`
	LastLogin *helper.JSONTime `json:"last_login"`
	CreatedAt *time.Time       `json:"created_at"`
	UpdatedAt *time.Time       `json:"updated_at"`
	DeletedAt *time.Time       `json:"deleted_at"`
}
```

> Or you can add a comment -- validate in the sql DDL inside the cmd/sql folder and regerate the model (recommended):

```sql
CREATE TABLE example (
  `id` CHAR(26) NOT NULL,
  `name` TEXT NOT NULL, -- validate: "min=5" -- sanitize-html
  `age` INT DEFAULT 0, -- validate: "required,number,gt=0,lt=100"
  `last_login` DATETIME DEFAULT NULL,
  `created_at` DATETIME DEFAULT NULL,
  `updated_at` DATETIME DEFAULT NULL,
  `deleted_at` DATETIME DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_example_deleted_at` (`deleted_at`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

## Default Behavior

> GRIT automatically detects and omits any model fields that have a database‐side DEFAULT clause when inserting new records and these fields are not present i the request. In practice, this means:

If in your SQL you have a field with default value like:

```sql
...
status_name VARCHAR(20) NOT NULL DEFAULT 'active',
notes TEXT DEFAULT NULL,
...
```

And if you make an "/add" request without informing these fields, they will be omitted in the insert query, letting the database use his DEFAULT VALUE.

---

## Testing & Coverage

Run all tests with coverage:

```bash
./audit.sh
```

See `./tests/coverage/coverage-unit.html` for details.

---

For more request examples, see `./ops/curl.sh`. Suggestions and contributions welcome!

If you are using a REST API software like Insomnia or Postman you can import the `./ops/rest_environment.json` and `./ops/rest_collection.json` files.