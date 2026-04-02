# Product Requirements Document (PRD): Universal Logistics Orchestration Middleware (Project "Synapse")

**Author:** Senior Product Manager  
**Status:** Draft / Conceptual  
**Target Architecture:** High-Concurrency Microservices (Golang)

---

## 1. Executive Summary
Project Synapse is a high-performance middleware designed to act as the "Source of Truth" and "Control Plane" between Warehouse Management Systems (WMS) and Delivery Information Systems (DIS). 

The goal is to decouple WMS and DIS providers completely. Synapse will ingest heterogeneous data, normalize it into a canonical format, apply complex business logic (billing, reattempts, communications), and orchestrate the state across all integrated platforms. The system is designed for massive scale, handling millions of delivery events with sub-millisecond processing latency and guaranteed consistency.

---

## 2. Problem Statement
1.  **Vendor Fragmentation:** Every WMS (EasyEcom, Uniware) and DIS (Loginext, Valkyrie) has a unique schema, status codes, and API limitations.
2.  **Inflexible Orchestration:** Direct integrations make it impossible to switch vendors or apply custom business rules (like automated reattempts) without rewriting code for every pair.
3.  **Communication Gaps:** DIS systems often lack the context of the customer’s interaction history.
4.  **Operational Leakage:** Billing and RTO (Return to Origin) tracking often fail due to mismatched order states between the warehouse and the rider.

---

## 3. Product Vision
To build a **Stateless-Core/Stateful-Edge** middleware that provides a "Universal Language" for logistics, enabling 100% automated lifecycle management of a shipment from manifest creation to final settlement.

---

## 4. Core Functional Modules

### 4.1. The Ingest Gateway (Adapter Layer)
*   **Agnostic Endpoints:** Secure, versioned endpoints capable of accepting raw webhooks from any vendor.
*   **Vendor Signature Validation:** Automatic detection and security handshake based on the inbound source (X-Loginext-Signature, API Keys, etc.).
*   **Raw Persistence (The Audit Trail):** Every incoming payload must be stored in its raw form before processing to ensure non-repudiation and support "Time-Travel" debugging.

### 4.2. Canonical Data Normalization Engine
*   **Field Mapping Registry:** A dynamic registry that maps vendor-specific fields (e.g., `orderNo` vs `awb_number`) to a **Synapse Canonical Schema**.
*   **Status Quantization System:** Mapping hundreds of vendor-specific sub-statuses into a unified **Synapse State Machine**:
    *   *Synapse Statuses:* `READY_TO_SHIP`, `PICKUP_IN_PROGRESS`, `IN_TRANSIT`, `ATTEMPTED_FAILED`, `DELIVERED`, `RTO_INITIATED`, `CANCELLED`.
*   **Entity Resolution:** Identifying the unique Zippee Waybill from any combination of Order IDs, Invoice IDs, or DIS Tracking Numbers.

### 4.3. The Orchestration & Workflow Engine
*   **Event Trigger Matrix:** A logic layer that determines "If DIS Event = X, then WMS Action = Y."
*   **State Conflict Resolver:** Handling out-of-order webhooks (e.g., receiving "Delivered" before "Picked Up") to ensure the internal database reflects the highest-order truth.
*   **Asynchronous Task Spawning:** Breaking down a single incoming event into multiple independent tasks:
    1.  Update WMS state.
    2.  Update Internal Ledger (Billing).
    3.  Trigger Customer Comms.
    4.  Update Analytics/Reporting.

### 4.4. Automated Reattempt & Feedback Loop (V3 Logic)
*   **Condition-Based Reattempt:** Monitor `ATTEMPTED_FAILED` statuses. If specific criteria (e.g., COD order, < 3 attempts) are met, trigger the reattempt workflow.
*   **Interactive Decisioning:** Orchestrating two-way communication (WhatsApp/SMS).
    *   *Customer Choice:* "Reschedule to Friday."
    *   *Middleware Action:* Automatically hit the DIS API to update the SLA/Slot without human intervention.
*   **Automatic RTO Enforcement:** If reattempts are exhausted or the customer requests cancellation, Synapse must automatically initiate the return process in both DIS and WMS.

### 4.5. Transactional Ledger & Billing (The "Wallets")
*   **Real-time Cost Calculation:** Based on order weight, dimensions, distance, and payment mode (COD vs Prepaid).
*   **Atomic Wallet Deductions:** Interfacing with the Billing module to ensure every successful delivery or RTO is billed instantly.
*   **Dispute Reconciliation:** Flagging discrepancies between the "Claimed Weight" from WMS and the "Billed Weight" from DIS.

---

## 5. Non-Functional Requirements (High-Scale Advanced)

### 5.1. Idempotency & Guaranteed Delivery
*   **Message Deduplication:** Ensure that even if a DIS provider sends the same "Delivered" webhook five times, Synapse only bills the customer and updates the WMS once.
*   **Dead Letter Queue (DLQ) Management:** Any event that fails normalization or orchestration must be captured in a DLQ for automated or manual replay.

### 5.2. Multi-Tenancy
*   **Namespace Isolation:** Support for multiple brands, each with their own set of WMS/DIS credentials and custom status-mapping overrides.

### 5.3. Latency & Throughput
*   **Concurrent Execution:** The middleware should process WMS updates, Comms, and Billing tasks in parallel for a single event.
*   **Pressure Handling:** Ability to handle spikes (e.g., 10x normal load during "Big Billion Days") using internal rate limiting and back-pressure mechanisms.

### 5.4. Observability (The "Log DNA")
*   **Traceability:** Every event must carry a unique `Correlation-ID` that links the DIS Webhook -> Synapse Processing -> WMS Update -> WhatsApp Sent.
*   **Status Dashboard:** Real-time visibility into "Webhook Health" (Success rate per vendor).

### 5.5. Persistence & Scalability Strategy
*   **Initial Backend (PocketBase):** For the initial development and prototype phases, **PocketBase** will be used as the primary data store and administrative interface due to its rapid development capabilities and all-in-one architecture.
*   **Data Layer Decoupling:** To prevent vendor lock-in and ensure long-term scalability, the application logic must remain decoupled from PocketBase internals. 
    *   *Abstraction Requirement:* All data operations must be channeled through a dedicated **Data Access Layer (DAL)**.
    *   *Internal Avoidance:* Developer logic should avoid direct usage of PocketBase-specific Go SDK methods or client libraries where possible. Instead, lean towards standard patterns and raw SQL queries (or generic interfaces) that can be easily ported.
*   **Migration Path (PostgreSQL):** The system architecture must be designed with the explicit goal of migrating to **PostgreSQL** (or a similar high-performance RDBMS) if the project experiences significant scale or requires more complex query optimization and ACID guarantees beyond SQLite's scope.

---

## 6. User Journeys (Examples)

### 6.1. The Standard Delivery Success
1.  **DIS (Valkyrie)** sends a `DELIVERED` webhook.
2.  **Synapse Ingest** validates the signature.
3.  **Normalization Engine** converts `Valkyrie.status.DELIVERED` to `Synapse.Status.3`.
4.  **Orchestrator** triggers:
    *   **Billing:** Calculates cost and deducts from Brand Wallet.
    *   **WMS (EasyEcom):** Hits `/updateTrackingStatus` with ID `3`.
    *   **Comms:** Sends "Your order is delivered" WhatsApp.

### 6.2. The Customer-Driven Reschedule
1.  **DIS (Loginext)** sends `ATTEMPTED_FAILED`.
2.  **Synapse** checks reattempt count. Count = 1.
3.  **Synapse Comms** sends WhatsApp with a Date Picker.
4.  **Customer** selects "Tomorrow, 2 PM."
5.  **Synapse Ingest** receives the WhatsApp reply.
6.  **Synapse Orchestrator** hits **Loginext API** to update the `deliverStartTimeWindow` and resets the Synapse state to `READY_TO_SHIP`.

---

## 7. Success Metrics
1.  **Decoupling Efficiency:** Time taken to onboard a new DIS vendor (Goal: < 2 days).
2.  **Processing Latency:** Time from DIS webhook receipt to WMS update sent (Goal: < 500ms).
3.  **Automation Rate:** % of `ATTEMPTED_FAILED` orders resolved without manual ops intervention (Goal: > 80%).
4.  **Data Integrity:** Zero discrepancy between Synapse shipment state and WMS shipment state.
