# Synapse Logistics Middleware

Synapse is a high-performance logistics middleware built with Go and PocketBase. It acts as a bridge between multiple WMS (Warehouse Management Systems) and DIS (Delivery Integration Systems).

## Prerequisites

- [Go 1.25+](https://go.dev/dl/)
- [.env file](./.env) with required configuration

## Quick Start

### 1. Environment Configuration
Create a `.env` file in the root directory:
```env
PB_URL=http://localhost:8090
PORT=8080
ENV=development
```

### 2. Install Dependencies
```bash
go mod tidy
```

### 3. Database Migrations
Run the initial migrations to establish the schema (Brands, Warehouses, Orders, etc.):
```bash
PB_URL=http://localhost:8090 go run cmd/synapse/main.go migrate up --dir=internal/data/pocketbase/migrations
```

### 4. Run the Server
#### Development (with hot-reload)
```bash
PB_URL=http://localhost:8090 air
```

#### Production / Standard Run
```bash
PB_URL=http://localhost:8090 go run cmd/synapse/main.go serve
```

## API Testing

### Webhook Ingestion
You can test the webhook ingestion by sending a POST request to the ingest endpoint:

```bash
curl -X POST http://localhost:8080/api/v1/webhook/ingest \
  -H "Content-Type: application/json" \
  -H "X-Vendor-Source: LOGINEXT" \
  -d '{"order_id": "ORD-123", "status": "shipped"}'
```

### Health Check
```bash
curl http://localhost:8080/health
```

## Project Structure

- `cmd/synapse/`: Application entry point.
- `internal/adapter/`: Ingest and Provider adapters.
- `internal/core/domain/`: Core entities and domain models.
- `internal/data/pocketbase/migrations/`: Database schema definitions.
- `internal/core/config/`: Type-safe environment configuration.
- `internal/core/logger/`: Structured logging system.

## Troubleshooting

- **Database Locks:** If migrations or the server hang, ensure no stale `air` or `synapse` processes are running (`pkill -f synapse`).
- **Missing Variables:** Ensure `PB_URL` is set in your environment or `.env` file.
