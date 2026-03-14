# Data Gathering Checklist for Architecture Improvements

Please fill out or review the following information so we can safely implement the architectural improvements.

## Phase 1: Security & Configuration (P0)
- [ ] **Environment Variables:** Provide a list of all environment variables the application needs to run (e.g., `DATABASE_URL`, `JWT_SECRET`, `PORT`, API keys). This will be used to build a safe `.env.example` file.
- [ ] **Hardcoded Fallback:** Confirm if it is acceptable to permanently remove the hardcoded Neon PostgreSQL database URL from `config/config.go` so the app strictly relies on environment variables.

## Phase 2: Infrastructure Pivot to PostgreSQL (P1)
- [ ] **Data Migration:** Indicate if there is any *live, production data* currently in MongoDB that needs to be migrated to PostgreSQL. If it's only local test data, we can skip writing a migration script.
- [ ] **Mongo Entities:** List the specific models/entities that are currently stored exclusively in MongoDB (e.g., `Recipes`, `Yields`, `Inventory Audits`). This will guide the design of the new Postgres JSONB tables.

## Phase 3: Global State & Dependency Injection (P1)
- [ ] **Background Jobs:** Note if there are any background workers, cron jobs, or init scripts running in the project that rely on the database, or if everything is purely triggered by HTTP requests.

## Phase 4: Performance & Template Management (P2)
- [ ] **Local Development Caching:** When we cache the HTML templates to improve performance, do you want a "Development Mode" flag (like `ENV=development`) that turns caching off? (This ensures you don't have to restart the server every time you edit an HTML file locally).

---
*Once you have gathered this information, we can proceed with Phase 1 execution.*