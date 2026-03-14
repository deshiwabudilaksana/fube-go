# Project FUBE (Food & Beverage Unified Engine)
**System Documentation**  
**Version:** 1.0  
**Status:** Active Development  

---

## 1. Project Overview
**Project FUBE** is a specialized, high-performance Point of Sale (POS) and Business Intelligence platform designed for the Food & Beverage industry. 

### Core Value Proposition: The "Little-by-Little" Migration
Unlike traditional POS systems that require an "all-or-nothing" switch, FUBE is designed as an **Inventory Sidecar**. It allows restaurant owners to keep their existing front-of-house hardware (Square, Toast, etc.) while migrating their back-of-house intelligence (costs, yields, and inventory) to FUBE for immediate profitability insights.

---

## 2. Tech Stack
*   **Backend:** Go (Golang) 1.23+  
*   **Database:** MongoDB (Local)  
*   **Frontend:** HTMX + Go HTML Templates + Tailwind CSS  
*   **Security:** JWT-based Auth (Scoped to `VendorID` and `StoreID`)  
*   **Testing:** Go `testing` package with `testify`  

---

## 3. System Architecture
FUBE uses a **Hyper-Productive SSR (Server-Side Rendering)** architecture. Instead of a heavy JavaScript SPA, it uses Go to render HTML fragments that are dynamically swapped into the DOM by HTMX.

### Key Components:
1.  **The Calculation Engine:** A Go service that uses **MongoDB Aggregation Pipelines** to calculate recipe costs in a single database round-trip.
2.  **The Migration Bridge:** A suite of tools (CSV Importer + Mapping UI) that ingests legacy POS data and links it to FUBE's high-accuracy ingredient models.
3.  **The Document Store:** A flexible MongoDB schema that supports nested recipes (Yields) and arbitrary menu attributes.

---

## 4. Database Schema (MongoDB Models)
The schema is designed for **Multi-Unit Scalability**. Every document is scoped by `VendorID` (the business owner) and `StoreID` (the specific branch).

### 4.1 Core Models
*   **`MaterialDoc`**: Raw ingredients (e.g., "Beef Patty") with price-per-unit.
*   **`YieldDoc`**: A "prep" recipe (e.g., "Burger Sauce") that embeds multiple materials.
*   **`MenuDoc`**: The sellable item (e.g., "Cheeseburger") that embeds one or more yields and an `ExternalPosID` for legacy system mapping.
*   **`OrderDoc`**: Full transaction data with split-payment and tax compliance.
*   **`InventoryDoc`**: Real-time stock levels with batch and expiry tracking.

---

## 5. Core Logic: High-Performance Costing
FUBE avoids the "N+1 Query" problem common in recipe management by performing calculations on the database server.

### The Aggregation Pipeline:
When a user requests a menu cost, the Go backend executes a multi-stage pipeline:
1.  **`$match`**: Filter by MenuID and VendorID.
2.  **`$lookup`**: Join the nested `Yields` and their `Materials`.
3.  **`$addFields`**: Multiply ingredient quantities by material prices.
4.  **`$group`**: Sum the total cost into a single decimal result.

---

## 6. Migration Bridge Workflow
This is the "Step-by-Step" process for onboarding a new restaurant:

1.  **Import:** Owner uploads a CSV export from their current POS (`/import`).
2.  **Map:** Owner uses the HTMX Mapping UI (`/mapping`) to link their imported "Item #101" to a "FUBE Burger Recipe."
3.  **Analyze:** Owner views the Planning Dashboard (`/yield-planning`) to see their real-time costs and margins.
4.  **Sync:** Owner exports a FUBE Cost Report (`/export/costs`) and uploads it back to their main POS to update their financial reporting.

---

## 7. Development Guide

### 7.1 Local Setup
1.  **Start MongoDB:** Ensure `mongod` is running on `localhost:27017`.
2.  **Configuration:** Create a `fube-go/.env` file:
    ```env
    MONGO_URI=mongodb://localhost:27017
    MONGO_DB_NAME=fube_local
    ```
3.  **Seed Data:** Run the seeder to populate sample F&B data:
    ```bash
    cd fube-go
    go run scripts/seed.go
    ```
4.  **Run Server:**
    ```bash
    go run main.go
    ```

### 7.2 Running Tests
FUBE uses dedicated test databases (`fube_test` and `fube_importer_test`) to ensure safety.
```bash
cd fube-go
go test -v ./...
```

---

## 8. Future Roadmap
*   **Phase 3:** Ingredient-Level Inventory Tracking (Theoretical vs. Physical).
*   **Phase 3:** Automated Supplier Ordering.
*   **Phase 4:** Kitchen Display System (KDS) using WebSockets.
*   **Phase 4:** Mobile Tablet App for table-side ordering.

---

**Orchestrator Note:** *This documentation reflects the current state of the codebase and its strategic direction. It is maintained by the Tech Lead for consistency across all specialized subagents.*
