# Synapse: Agent Skills & Architectural Guidelines

This document defines the core principles, design patterns, and coding standards for all agents working on Project Synapse.

## 1. Core Principles

### 1.1. Strict SOLID Adherence
- **Single Responsibility (SRP):** Each package, file, and function must have one reason to change.
- **Open/Closed (OCP):** Components must be open for extension but closed for modification (use interfaces for adapters).
- **Liskov Substitution (LSP):** Subtypes must be substitutable for their base types.
- **Interface Segregation (ISP):** Clients should not be forced to depend on methods they do not use.
- **Dependency Inversion (DIP):** Depend on abstractions, not concretions (essential for the Data Access Layer).

### 1.2. Design Patterns
- **Adapter Pattern:** Mandatory for all WMS and DIS integrations to handle vendor fragmentation.
- **Strategy Pattern:** Use for different cost calculation or reattempt logic.
- **Repository Pattern:** Abstract the persistence layer (PocketBase/PostgreSQL) from the business logic.
- **State Machine:** Canonical status mapping must be handled by a deterministic state machine.

## 2. Coding Standards & File Structure

### 2.1. Modular File Architecture
To maintain extreme modularity and prevent file bloat, we follow a granular naming and splitting convention:
- **Struct Definition:** `[entity]_struct.go` (e.g., `shipment_struct.go`) - strictly for type and struct definitions.
- **Method Implementation:** `[entity]_[action]_method.go` (e.g., `shipment_update_method.go`) - strictly for logic attached to the struct.
- **Interface Definition:** `[entity]_interface.go`.
- **Factory/Constructor:** `[entity]_factory.go`.

### 2.2. Nomenclature & Self-Documentation
- **No Abbreviations:** Use `order_number` instead of `ord_no`, `delivery_information_system` instead of `dis`.
- **Contextual Variables:** Variable names must be descriptive and context-aware.
- **Explicit Returns:** Use named return parameters where it improves readability for complex logic.
- **Comments:** Code must be self-documenting first; comments should explain *why*, not *what*.

## 3. Persistence Layer Guardrails
- **Data Access Layer (DAL):** Business logic must **never** call PocketBase SDK or internal methods directly.
- **Raw Query Preference:** Use standard Go database/sql or generic query builders that can move between SQLite (PocketBase) and PostgreSQL with minimal friction.
- **Interface Guards:** All database interactions must be behind interfaces to allow for easy swap-outs.
