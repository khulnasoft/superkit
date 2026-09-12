# Superkit 🚀

> **Build high-performance apps swiftly with minimal team resources in Go.**

**Superkit** is a full-stack web framework designed for speed and simplicity. It provides essential tools and libraries to help developers build modern web applications with ease.

⚠️ **Currently in Experimental Phase**

---

## 📖 Table of Contents

- [🌟 Features](#-features)
- [📥 Installation](#-installation)
- [🚀 Getting Started](#-getting-started)
  - [📂 Project Structure](#-project-structure)
  - [🎮 Development Server](#-development-server)
  - [🔥 Hot Reloading](#-hot-reloading)
- [📊 Database Migrations](#-database-migrations)
- [🛠 Creating Views with Templ](#-creating-views-with-templ)
- [✅ Validations](#-validations)
- [🧪 Testing](#-testing)
- [📦 Production Release](#-production-release)

---

## 🌟 Features
✅ **Lightweight & Fast** – Built on Go for blazing-fast performance.  
✅ **Modular Design** – Well-structured and easy to extend.  
✅ **Built-in Database Support** – Migrations, seeds, and ORM included.  
✅ **Templ-based Views** – Type-safe templating engine for UI components.  
✅ **Hot Reloading** – Instant feedback during development.  
✅ **One-Binary Deployment** – Compiles your app into a single executable.  

---

## 📥 Installation

Create a new **Superkit** project with a single command:

```sh
# Initialize a new Superkit project
go run github.com/khulnasoft/superkit@master [yourprojectname]

# Navigate into your project
cd [yourprojectname]

# Install TailwindCSS & esbuild
cd bootstrap && npm install && cd ..

# Copy environment configuration
cp bootstrap/.env.example bootstrap/.env
# Edit bootstrap/.env and set a secure SUPERKIT_SECRET (32+ characters)

# Resolve Go dependencies
go clean -modcache && go get -u ./...

# Initialize database migrations (if authentication plugin is enabled)
cd bootstrap && make db-up && cd ..
```

---

## 🚀 Getting Started

### 📂 Project Structure

```
├── bootstrap           # Example application (runs with `make dev` from bootstrap/)
│   ├── app
│   │   ├── assets      # Static files (CSS, JS)
│   │   ├── conf        # Configuration files
│   │   ├── db          # Database migrations
│   │   ├── events      # Custom event handlers
│   │   ├── handlers    # Request handlers (controllers)
│   │   ├── types       # Data models and interfaces
│   │   ├── views       # HTML templates (Templ)
│   │   ├── routes.go   # Route & middleware setup
│   │   ├── events.go   # Event registration
│   │   ├── db          # GORM/SQLite setup
│   │   ├── public      # Embedded production assets
│   ├── cmd
│   │   ├── app         # Application entrypoint
│   │   ├── scripts     # CLI commands & seed scripts
│   ├── plugins
│   │   ├── auth        # Authentication module
│   ├── public          # Development static assets
│   ├── Makefile        # Bootstrap build & run scripts
│   ├── go.mod          # Bootstrap module dependencies
├── kit                 # Core framework utilities
├── db                  # Database connection utilities
├── event               # Async event subscription & dispatch
├── validate            # Validation utilities
├── view                # View engine utilities
├── ui                  # Reusable UI components
├── Makefile            # Repository-level build & run scripts
├── go.mod              # Framework module dependencies
└── README.md           # Project documentation
```

---

### 🎮 Development Server

Run the development server from the `bootstrap/` directory:
```sh
cd bootstrap
make dev
```

This starts all development watchers: Templ (with live reload proxy), Air (Go rebuilds), TailwindCSS, and esbuild.

> **Note**: On Windows (WSL2), you might need to run asset watchers separately:
> ```sh
> make watch-assets
> ```

---

### 🔥 Hot Reloading
Hot reloading is enabled by default for CSS & JS.

> **Note**: On Windows (WSL2), you might need to run this command separately:
> ```sh
> make watch-assets
> ```

---

## 📊 Database Migrations

### Create a New Migration
```sh
make db-mig-create add_users_table
```
➡️ Generates a new migration SQL file in `app/db/migrations/`

### Apply Migrations
```sh
make db-up
```

### Reset the Database
```sh
make db-reset
```

### Seed the Database
```sh
make db-seed
```
➡️ Runs the seed script in `cmd/scripts/seed/main.go`

---

## 🛠 Creating Views with Templ

**SUPERKIT** uses [Templ](https://templ.guide) for type-safe UI components.  
Create structured, reusable HTML fragments with Go templates.

---

## ✅ Validations (Coming Soon)
> Stay tuned for built-in validation utilities!

---

## 🧪 Testing

### Test Framework Core
```sh
make test
```

### Test Bootstrap Application
```sh
make app-test
```

---

## 📦 Production Release

Compile your application into a single binary:
```sh
make app-build
```
➡️ Creates a production-ready binary at `bootstrap/bin/app_prod`.

Set the environment to **production**:
```sh
SUPERKIT_ENV=production
```

---

🚀 **Start building with SUPERKIT today!** 💙

