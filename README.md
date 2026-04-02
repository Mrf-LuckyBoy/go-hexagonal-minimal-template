# go-hexagonal-minal-template

A minimal, production-ready Go project template following **Hexagonal Architecture** (also known as Ports & Adapters). Designed to help you bootstrap new Go services quickly with a clean, testable, and maintainable structure.

--- 

## Features
- **Hexagonal Architecture** — clear separation between core business logic and external dependencies
- **Dependency Injection** — wired manually via `cmd/di.go` for full transparency and control
- **Config Management** — environment-aware YAML config powered by [Viper](https://github.com/spf13/viper)
- **SQLite + GORM** — included as a working database adapter out of the box (easy to swap, this for sample you can remove it and replace it what driver you use maridb, postgresql)
- **Auto Migration** — GORM auto-migrates domain models on startup
- **`.env` support** — secrets loaded from `.env` via `godotenv`

---

## 📁 Project Structure
 
```
go-hexagonal-minimal-template/
├── cmd/
│   ├── main.go          # Entry point — loads config, bootstraps DI container
│   └── di.go            # Dependency injection — wires repos, services, handlers
│
├── config/
│   ├── config.yaml      # Base configuration
│   ├── dev.yaml         # Development overrides
│   └── prod.yaml        # Production overrides
│
├── internal/
│   ├── config/
│   │   ├── config.go    # Config loader (Viper)
│   │   └── schema.go    # Config struct definitions
│   │
│   └── core/                        # ← The Hexagon (pure business logic)
│       ├── domain/
│       │   ├── model/               # Domain entities / DB models
│       │   │   └── book.go
│       │   └── dto/                 # Data Transfer Objects
│       │       └── book.go
│       ├── port/                    # Interfaces (contracts)
│       │   ├── book_repository.go   # Driven port (storage)
│       │   └── book_service.go      # Driving port (use cases)
│       └── service/                 # Business logic implementations
│           └── book_service.go
│
└── internal/adapters/               # ← The Adapters (infrastructure)
    └── repository/
        └── sqliter/                 # SQLite adapter (implements BookRepository)
            └── book_repository_sqlite.go
```

---
 
## 🏛️ Architecture Overview
 
This template follows **Hexagonal Architecture**, which keeps your core business logic completely isolated from frameworks, databases, and delivery mechanisms.
 
```
                    ┌─────────────────────────────┐
                    │         Core (Hexagon)        │
                    │                               │
  [Driving Ports]   │  port.BookService  (interface)│
  e.g. HTTP, CLI ──►│  service.bookService (impl)   │
                    │                               │
                    │  port.BookRepository(interface)│──► [Driven Ports]
                    │                               │     e.g. SQLite, Postgres
                    └─────────────────────────────┘
```
 
- **Ports** (`internal/core/port/`) — interfaces that define *what* the core needs, not *how* it gets it
- **Adapters** (`internal/adapters/`) — concrete implementations of those ports (databases, HTTP handlers, etc.)
- **Services** (`internal/core/service/`) — pure business logic that depends only on port interfaces
 
This means you can **swap SQLite for PostgreSQL**, or add an HTTP handler, without touching any business logic.
 

## 🚀 Getting Started
 
### Prerequisites
 
- Go 1.21+
- GCC (required for `go-sqlite3` CGO compilation)
 
### 1. Clone the template
 
```bash
git clone https://github.com/Mrf-LuckyBoy/go-hexagonal-minimal-template.git
cd go-hexagonal-minimal-template
```
 
### 2. Set up environment variables
 
```bash
cp .env.example .env
# Edit .env and fill in your secrets
```
 
### 3. Configure the app
 
Edit `config/config.yaml` for base settings, or `config/dev.yaml` / `config/prod.yaml` for environment-specific overrides:
 
```yaml
app:
  name: myapp
  port: 8080
  env: dev
  encryptkey: ""
 
database:
  host: localhost
  port: 5432
  user: admin
  password: ""  # Use .env for secrets
 
server:
  timeout: 30s
```
 
### 4. Run
 
```bash
go run ./cmd/...
```
 
---
 
## ⚙️ Configuration
 
Configuration is loaded via [Viper](https://github.com/spf13/viper) and supports layered overrides:
 
| File | Purpose |
|------|---------|
| `config/config.yaml` | Base defaults |
| `config/dev.yaml` | Development overrides |
| `config/prod.yaml` | Production overrides |
| `.env` | Secret values (never commit this) |
 
The config schema is defined in `internal/config/schema.go`:
 
```go
type Config struct {
    App      App      // name, port, env, encryptkey
    Database Database // host, port, user, password
    Server   Server   // timeout
}
```
 
---
 
## 🔌 Adding a New Feature
 
### 1. Define your domain model
 
```go
// internal/core/domain/model/user.go
type User struct {
    ID        string
    Email     string
    CreatedAt time.Time
}
```
 
### 2. Define port interfaces
 
```go
// internal/core/port/user_repository.go
type UserRepository interface {
    Create(user *model.User) error
    GetByID(id string) (*model.User, error)
}
```
 
### 3. Implement the service
 
```go
// internal/core/service/user_service.go
type userService struct {
    repo port.UserRepository
}
func NewUserService(repo port.UserRepository) port.UserService { ... }
```
 
### 4. Implement the adapter
 
```go
// internal/adapters/repository/sqliter/user_repository_sqlite.go
type UserRepositorySqlite struct { db *gorm.DB }
func (r *UserRepositorySqlite) Create(user *model.User) error { ... }
```
 
### 5. Wire it in DI
 
```go
// cmd/di.go
userRepo := sqliter.NewUserRepositorySqlite(db)
userService := service.NewUserService(userRepo)
```
 
---
 
## 🔄 Swapping the Database
 
Since the repository is behind a port interface, swapping databases only requires a new adapter:
 
```bash
# Example: Add a PostgreSQL adapter
mkdir -p internal/adapters/repository/postgres
```
 
```go
// internal/adapters/repository/postgres/book_repository_postgres.go
type BookRepositoryPostgres struct { db *gorm.DB }
// implement port.BookRepository interface...
```
 
Then update `cmd/di.go` to use the new adapter — no changes needed in service or domain layers.
 
---
 
## 🛠️ Using This Template
 
After cloning, follow these steps to make the project your own.
 
### 1. Rename the Go Module
 
The default module name is `github.com/Mrf-LuckyBoy/go-hexagonal-minimal-template`. Replace it with your own:
 
**Update `go.mod`:**
 
```go
// before
module github.com/Mrf-LuckyBoy/go-hexagonal-minimal-template
 
// after
module github.com/YourUsername/your-project-name
```
 
**Replace all import paths in one command (Linux):**
 
```bash
find . -type f -name "*.go" | xargs sed -i \
  's|github.com/Mrf-LuckyBoy/go-hexagonal-minimal-template|github.com/YourUsername/your-project-name|g'
```
 
> **macOS** — `sed -i` requires an empty string argument:
> ```bash
> find . -type f -name "*.go" | xargs sed -i '' \
>   's|github.com/Mrf-LuckyBoy/go-hexagonal-minimal-template|github.com/YourUsername/your-project-name|g'
> ```
 
**Verify no old references remain:**
 
```bash
grep -r "Mrf-LuckyBoy/go-hexagonal-minimal-template" --include="*.go" .
go mod tidy && go build ./...
```
 
---
 
### 2. Remove the SQLite Sample Code
 
SQLite and the `Book` domain are included only as a working example. Remove them when starting your real project:
 
**Delete the sample files:**
 
```bash
rm -rf internal/adapters/repository/sqliter/
rm internal/core/domain/model/book.go
rm internal/core/domain/dto/book.go
rm internal/core/port/book_repository.go
rm internal/core/port/book_service.go
rm internal/core/service/book_service.go
```
 
**Strip `cmd/di.go` back to a clean skeleton:**
 
```go
package main
 
import "github.com/YourUsername/your-project-name/internal/config"
 
type Container struct{}
 
func BuildContainer(cfg *config.Config) *Container {
    // wire your own repos and services here
    return &Container{}
}
```
 
**Remove SQLite dependencies:**
 
```bash
go get gorm.io/driver/sqlite@none
go get github.com/mattn/go-sqlite3@none
 
# Remove gorm entirely if not using it
go get gorm.io/gorm@none
 
go mod tidy
```
 
**Then add your own database driver:**
 
| Database | Command |
|----------|---------|
| PostgreSQL (GORM) | `go get gorm.io/driver/postgres` |
| MySQL (GORM) | `go get gorm.io/driver/mysql` |
| PostgreSQL (raw) | `go get github.com/jackc/pgx/v5` |
| MongoDB | `go get go.mongodb.org/mongo-driver/mongo` |
 
---
 
## 📦 Dependencies
 
| Package | Purpose |
|---------|---------|
| `gorm.io/gorm` | ORM |
| `gorm.io/driver/sqlite` | SQLite driver |
| `github.com/mattn/go-sqlite3` | SQLite CGO bindings |
| `github.com/spf13/viper` | Config management |
| `github.com/joho/godotenv` | `.env` file loading |
 
---
 
## 📝 License
 
This project is licensed under the MIT License. See [LICENSE](./LICENSE) for details.
