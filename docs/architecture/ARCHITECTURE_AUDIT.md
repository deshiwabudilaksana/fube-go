# Fube-Go Architecture Audit & Documentation

**Date:** March 13, 2026
**Project:** Fube-Go
**Type:** Architecture Audit & Remediation Plan

---

## 1. Architectural Summary

The `fube-go` project implements a hybrid **Monolith** architecture designed to serve both a **REST API** and a **Server-Side Rendered (SSR)** frontend utilizing **HTMX**. 

### Data Layer
The system employs a multi-database strategy:
*   **PostgreSQL (via GORM):** Utilized for relational data, authentication, and core entities.
*   **MongoDB:** Utilized for document-based storage of complex, flexible entities such as recipes (Yields) and inventory audits.

### Key Design Patterns
*   **Dependency Injection (Partial):** DI is used in some service layers (e.g., `auth.Service`) but is neglected in others where logic is embedded directly in models or handlers.
*   **Singleton Pattern:** Used for database connections in `database/database.go` and `db/mongo.go` using `sync.Once`.
*   **Context-Aware Operations:** The project demonstrates modern Go practices by correctly propagating `context.Context` through most database and service methods.
*   **Fat Model / Thin Controller (Partial):** Business logic is inconsistently distributed between handlers, services, and model packages.

---

## 2. Critical Flaws & Anti-Patterns

### A. Security: Hardcoded Credentials & Inconsistent Config (High Severity)
*   **The Issue:** `config/config.go` contains a hardcoded production-grade Connection String for a Neon PostgreSQL database acting as the default fallback. Additionally, `services/auth/auth.go` uses `os.Getenv` directly instead of relying on the centralized `config` struct.
*   **Impact:** Committing secrets to version control is a critical security vulnerability.
*   **Recommendation:** Remove hardcoded credentials immediately. Rely entirely on environment variables (e.g., via a `.env` file) for local development and CI/CD secrets for production.

### B. Data Duality & Model Duplication (Severe)
*   **The Issue:** The project maintains two nearly identical sets of models: `models.go` (SQL/GORM) and `mongo_models.go` (MongoDB). Entities like `User`/`UserDoc` and `Vendor`/`VendorDoc` exist in both systems.
*   **Impact:** This creates a massive maintenance burden, source of truth confusion, and high risk of data desynchronization between the two databases.
*   **Recommendation:** Consolidate data ownership. Define clearly which database is the "Source of Truth" for core entities. Use MongoDB strictly for flexible documents and SQL for relational data. Eliminate duplicate models.

### C. Global State & Service Lifecycle (High)
*   **The Issue:** The project relies on a global `database.DB` variable. Handlers (e.g., `auth_handlers.go`) are also repeatedly calling `db.GetDatabase` and `config.Load` on every request.
*   **Impact:** Global state makes unit testing nearly impossible as it prevents mocking the database layer. Loading configuration on every request is inefficient.
*   **Recommendation:** Move to pure Constructor-based Dependency Injection. Initialize the database, configuration, and services **once** in `main.go` and pass them down to handlers as struct dependencies.

### D. Performance Bottlenecks: Template Management (High)
*   **The Issue:** Handlers are repeatedly parsing HTML templates on every single request (e.g., `template.ParseFiles` in `handlers/auth_handlers.go`). 
*   **Impact:** This involves disk I/O and parsing logic every time a user hits a page, which will cause significant latency and CPU spikes in a high-traffic environment.
*   **Recommendation:** Parse all templates **once** at application startup and store them in a global `template.Template` cache.

### E. Error Handling: Crashing on Library Initialization (Medium)
*   **The Issue:** `database.ConnectDB` uses `log.Fatal` if the connection fails.
*   **Impact:** Libraries and helper packages should not terminate the process. This prevents `main.go` from performing graceful shutdowns or logging errors to external monitoring systems.
*   **Recommendation:** Return `error` to the caller (e.g., `main.go`) and handle the fatal exit at the highest level of the application.

---

## 3. Recommended Remediation Plan

To bring the project up to **Tiger Style** and **Safe Golang** standards, the following steps should be prioritized:

| Priority | Category | Finding | Recommended Action |
| :--- | :--- | :--- | :--- |
| **P0** | **Security** | Hardcoded DB URLs in `config.go`. | Remove secrets; enforce `.env` usage for local development. |
| **P1** | **Architecture**| Hybrid GORM/Mongo model duplication. | Consolidate data ownership. Remove duplicate `Doc` structs. |
| **P1** | **Safety** | Global `var DB *gorm.DB` and inline `config.Load()`. | Remove package-level variables. Implement Dependency Injection in `main.go`. |
| **P2** | **Performance** | Template parsing in handlers. | Parse templates during `init` or at `main` startup. Use a global template cache. |
| **P3** | **Logic** | Business logic in `models/` package. | Move logic from `models/inventory_mongo.go` into a dedicated `InventoryService` in `services/inventory`. |
| **P3** | **Error Handling**| `log.Fatal` in `database` package. | Return errors from connection functions; handle `log.Fatal` exclusively in `main.go`. |

---
*Documentation generated via automated architecture audit.*