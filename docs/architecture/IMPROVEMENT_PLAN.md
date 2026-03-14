# Fube-Go Improvement & Data Gathering Plan

This document outlines the step-by-step plan to implement the architecture improvements identified in the audit. 

Crucially, it details the **Data, Decisions, and Context** you need to gather *before* we can safely implement each phase.

---

## Phase 1: Security & Configuration (P0)
**Goal:** Remove hardcoded secrets from `config.go`, centralize configuration loading, and enforce `.env` usage.

### 📋 Data & Decisions to Gather:
*   [ ] **List of All Environment Variables:** Identify every environment variable the application relies on (e.g., `DATABASE_URL`, `MONGO_URI`, `JWT_SECRET`, `PORT`, any third-party API keys).
*   [ ] **Create `.env.example`:** We need a template file that lists all required keys with dummy values so other developers know what to configure.
*   [ ] **Production Secret Management:** Decide how secrets will be injected in production (e.g., Docker environment variables, a PaaS like Railway/Render/Heroku, or a secrets manager).

---

## Phase 2: Infrastructure Pivot - The Power of One (P1)
**Goal:** Deprecate MongoDB entirely and consolidate all data to PostgreSQL to cut infrastructure overhead in half, simplify local development, and gain ACID guarantees across the application.

### 📋 Data & Decisions to Gather:
*   [ ] **Identify Mongo Documents:** List all entities currently stored in MongoDB (e.g., `Recipe`, `Inventory`, `Yield`).
*   [ ] **PostgreSQL Schema Design:** Decide how to structure these document entities in PostgreSQL.
    *   *Recommendation:* Use PostgreSQL `JSONB` columns for unstructured or highly flexible document data (like dynamic recipe attributes). Use standard relational tables for structured data.
*   [ ] **Migration Strategy:** Do we have live data in MongoDB that needs to be migrated to PostgreSQL? If yes, we need to map out a one-off data migration script.
*   [ ] **Deployment Strategy:** Are we ready to move to a PaaS like Railway, Render, or Fly.io using Docker to host our Go binary alongside a managed PostgreSQL 16+ database?

---

## Phase 3: Global State & Dependency Injection (P1)
**Goal:** Eliminate global variables like `database.DB` and implement Constructor-based Dependency Injection in `main.go`.

### 📋 Data & Decisions to Gather:
*   [ ] **Audit of Global Usages:** Identify all handlers, services, and middlewares that currently call the global `database.DB` or `config.Load()`.
*   [ ] **Application Struct Design:** Decide on the structure that will hold our dependencies. Usually, this looks like:
    ```go
    type Server struct {
        Config *config.Config
        PG     *gorm.DB
        Mongo  *mongo.Database
        // Services...
    }
    ```
*   [ ] **Background Processes:** Are there any background workers, cron jobs, or init scripts that also rely on the global DB that we need to refactor?

---

## Phase 4: Performance & Template Management (P2)
**Goal:** Stop parsing HTML templates on every request. Parse them once at startup and cache them.

### 📋 Data & Decisions to Gather:
*   [ ] **Template Inventory:** List all HTML templates and partials used across the project.
*   [ ] **Template Structure:** How are templates currently nested? (e.g., does every page include a `base.html` layout?). This determines how we build the `template.Template` cache.
*   [ ] **Local Development Needs:** Do you want a flag (like `ENV=development`) that bypasses the cache and reloads templates on every request *only* during local development, so you don't have to restart the server when tweaking HTML?

---

## Phase 5: Code Organization & Error Handling (P3)
**Goal:** Move business logic out of the `models/` directory into dedicated `services/` and remove `log.Fatal` from library packages.

### 📋 Data & Decisions to Gather:
*   [ ] **Identify Rogue Logic:** List all files in the `models/` directory that contain active business logic (e.g., `inventory_mongo.go` calculating totals) instead of just struct definitions.
*   [ ] **Graceful Shutdown Strategy:** When `main.go` catches a database connection error, should it simply exit, or do we have other resources (like closing HTTP servers) that need to gracefully shut down first?

---

### How to Proceed
Once you have gathered the information for **Phase 1**, let me know, and we can immediately execute the code refactoring for that phase!