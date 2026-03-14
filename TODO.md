# Project FUBE: Task Tracking

## Current Stack
- **Backend:** Go (Golang)
- **Database:** MongoDB (Local)
- **Frontend:** HTMX + Go Templates + Tailwind CSS

---

## Todo List

### Phase 1: Foundation (Local Mongo + Go + HTMX)
- [x] Design MongoDB Schema for Yields, Menus, and Inventory (Document-oriented)
- [x] Initialize Go-Mongo Driver and Connection Pool in 'fube-go'
- [x] Configure fube-go to use local MongoDB (localhost:27017)
- [x] Implement MongoDB Aggregation Pipelines for Yield/Recipe costing logic
- [x] Set up HTMX-compatible Go Handlers (returning HTML fragments)
- [x] Create a 'Base Layout' using Go html/template and HTMX for the Dashboard
- [x] Update HTMX dashboard to display live data from local MongoDB
- [x] Perform end-to-end verification of local Mongo + Go + HTMX flow

### Phase 2: Data & Migration Bridge
- [x] Create MongoDB Seeder script for sample F&B data (Menus, Yields, Materials)
- [x] Implement CSV Importer for legacy POS data (Toast, Square, etc.)
- [x] Develop 'Mapping UI' in HTMX to link External POS IDs to FUBE Documents
- [x] Implement Export Service for reporting costs back to the main POS

### Phase 3: Advanced F&B Features
- [x] Ingredient-Level Inventory Tracking (Theoretical vs. Physical Stock)
- [x] Develop Kitchen Display System (KDS) using HTMX real-time updates
- [x] Implement Order State Machine (Pending -> Preparing -> Ready -> Served)
- [x] Supplier Ordering System (Auto-generate POs based on Planned Sales)
- [x] Authentication & Onboarding Flow (HTMX Sign-up/Login)

### Phase 4: Quality & Reliability
- [x] Implement unit tests for Go MongoDB Aggregation Pipelines
- [x] Implement unit tests for CSV Importer logic
- [x] Create Product Onboarding & User Guide
- [ ] Implement integration tests for Go + HTMX endpoints
- [ ] Perform security and concurrency stress tests on POS Order logic

### Phase 5: Full POS Evolution (The "Front-of-House" Leap)
- [ ] **Dynamic Table & Seat Management**: Visual floor plan UI for opening bills and tracking table status.
- [ ] **Tablet-Optimized Ordering UI**: High-speed, touch-friendly menu for server entry.
- [ ] **Advanced Split-Bill Logic**: Backend math and UI for itemized and percentage splitting.
- [ ] **Payment Gateway Integration**: Connection to Stripe Terminal or Adyen for physical card readers.
- [ ] **Multi-Region Tax Engine**: Flexible VAT/GST/Service Charge calculation logic.
- [ ] **Integrated Loyalty & CRM**: Link 'fube-membership' for point redemption at checkout.
- [ ] **Offline-Sync Architecture**: Local edge database for zero-downtime operations during outages.
- [ ] **Hardware Integration (Printing)**: Support for ESC/POS thermal receipt and kitchen printers.

---

## MVP Status: RELEASED (v1.1)
Project FUBE is now officially at its **v1.1 "Proper UI"** stage. It is ready for local deployment and initial merchant testing.


---

## Technical Notes
- **Multi-tenancy:** All MongoDB collections must include `vendor_id`.
- **Performance:** Use MongoDB Aggregation Pipelines for all cost calculations to ensure single-trip database operations.
- **HTMX Pattern:** Prefer `hx-post` with `hx-target` for partial DOM updates to maintain "No-JS" simplicity.
