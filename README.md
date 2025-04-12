# GRIT – Go REST Interface Toolkit

GRIT is a minimalist microservice framework built in pure Go. Designed to simplify RESTful API development without sacrificing performance, it draws inspiration from ALA (Automatic Lumen API) while embracing Go's philosophy of simplicity, concurrency, and clarity.

---

## Features

- **Intuitive Routing** – Clean and easy route definitions
- **Modular Architecture** – Compose scalable services with clear interfaces
- **High Performance** – Powered by native Go performance and concurrency
- **Pluggable Middleware** – Add auth, logging, rate-limiting, and more
- **Testing First** – Built with testability and simplicity in mind
- **OpenAPI/Swagger Ready** – Docs generated automatically

---

## Getting Started

```bash
go get github.com/not-empty/grit
```

```go
package main

import "github.com/not-empty/grit"

func main() {
    app := grit.New()

    app.GET("/ping", func(ctx grit.Context) {
        ctx.JSON(200, grit.H{"message": "pong"})
    })

    app.Run(":8080")
}
```

?filter=field:operator:value

eq	filter=age:eq:30	age = 30
gt	filter=age:gt:30	age > 30
lt	filter=age:lt:30	age < 30
gte	filter=age:gte:18	age >= 18
lte	filter=created_at:lte:2024-12-31	created_at <= '2024-12-31'
like	filter=email:like:gmail.com	email LIKE '%gmail.com%'
isnull	filter=deleted_at:isnull:true	deleted_at IS NULL or IS NOT NULL
between	filter=created_at:between:2024-01-01,2024-12-31	created_at BETWEEN '2024-01-01' AND '2024-12-31'