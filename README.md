# Fube-Go (Food & Beverage Unified Engine)

Fube-Go is a hybrid Monolith serving both a REST API and a Server-Side Rendered (SSR) frontend using HTMX.

## 🚀 Quickstart: Developer Onboarding

The fastest way to onboard a new developer to this codebase is using our automated **Makefile**.

### Step 1: Install Dependencies
Run the setup command. This will download Go modules, install `air` for hot-reloading, and automatically create your local `.env` file from the `.env.example` template.
```bash
make setup
```

### Step 2: Configure Your Environment
Open the newly created `.env` file in the root of the project and fill in your local database credentials (e.g., `DATABASE_URL` and `MONGO_URI`).

### Step 3: Start the Server (with Hot Reload)
Start the application using `air` so that the server automatically restarts whenever you save a `.go` or `.html` file.
```bash
make dev
```

The application will now be running at `http://localhost:8080`.

---

## 🛠️ Other Useful Commands

| Command | Description |
| :--- | :--- |
| `make run` | Runs the application normally (without hot reload). |
| `make test` | Runs the full Go test suite. |
| `make build` | Compiles the application into a single binary inside the `/bin` directory. |
| `make clean` | Removes build artifacts and temporary files. |

---

## 🏗️ Project Architecture
For a deep dive into the current infrastructure and the planned migration from MongoDB to a pure PostgreSQL architecture, please read the documents in the `docs/architecture/` folder.
