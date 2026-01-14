# 📦 Real-Time Distributed Order Tracker

A high-performance, event-driven logistics backend built with **Go**, **Kafka**, **Redis**, and **PostgreSQL**.

This system handles order ingestion, real-time geospatial dispatching to nearby shoppers, and atomic order claiming, simulating the core logistics engine of a modern delivery platform.

## 🎥 Demo
[![Demo Video](https://img.youtube.com/vi/dI_CtcXSh2s/0.jpg)](https://www.youtube.com/watch?v=dI_CtcXSh2s)

## 🚀 Key Features

*   **Real-Time Geospatial Dispatching**: Uses **Redis Geo** to instantly find shoppers within a 15km radius of a store.
*   **Event-Driven Architecture**: Decoupled services using **Apache Kafka** for asynchronous communication (Order Service -> Dispatch Service -> Notification Service).
*   **WebSocket Push Notifications**: Instantly pushes offers to connected shoppers via a WebSocket hub.
*   **Concurrency Safe**: Implements **Atomic Order Claiming** using PostgreSQL Transactional Locking (`FOR UPDATE`) to prevent race conditions (double-booking).
*   **Session Persistence**: Couriers can refresh their browser or re-login without losing their active delivery state.
*   **Smart Simulator**: Includes a dynamic "Ghost Car" simulator that mimics shopper behavior, automatically respawning specifically for the logged-in user for testing.
*   **Scalable Design**: Built with modular Clean Architecture principles.

## 🛠️ Tech Stack

*   **Language**: Go (Golang)
*   **API Framework**: Gin
*   **Database**: PostgreSQL (pgx driver)
*   **Caching/Geospatial**: Redis
*   **Message Broker**: Apache Kafka
*   **Real-Time**: Native WebSockets
*   **Infrastructure**: Docker & Docker Compose

## 🏗️ Architecture

### High-Level System Design
```mermaid
graph TD
    Client[Client App / Curl] -->|POST /orders| API[Order Service]
    API -->|Write Order| DB[(PostgreSQL)]
    API -->|Publish: order.created| Kafka{Kafka}
    
    Disp[Dispatch Service] -->|Consume: order.created| Kafka
    Disp -->|Geo Search| Redis[(Redis)]
    Disp -->|Publish: offer.created| Kafka
    
    WS[WebSocket Hub] -->|Consume: offer.created| Kafka
    WS -->|Push: NEW_OFFER| Shopper[Shopper App / Websocat]
    
    Sim[Shopper Simulator] -->|Update Location| API
    API -->|GeoAdd| Redis
```

### Database Schema (ERD)
![Database ERD](assets/erd.png)

### Dispatch Flow Sequence
```mermaid
sequenceDiagram
    participant C as Customer
    participant API as Order Service
    participant K as Kafka
    participant D as Dispatch Service
    participant R as Redis (Geo)
    participant W as WebSocket Hub
    participant S as Shopper

    C->>API: Create Order
    API->>K: Event: ORDER_CREATED
    K->>D: Consume Event
    D->>R: Find Shoppers nearby (15km)
    R-->>D: Returns [Shopper A, Shopper B]
    loop For Each Shopper
        D->>K: Event: OFFER_CREATED
    end
    K->>W: Consume Event
    W->>S: Push WebSocket Message {"type": "NEW_OFFER"}
```

## ⚡ Getting Started

### Prerequisites
*   Docker & Docker Compose
*   Go 1.22+ (optional, for local dev)

### 1. Start Infrastructure
Start Postgres, Redis, Kafka, and Zookeeper.
```bash
docker-compose up -d
```

### 2. Run Database Migrations
Initialize the schema and seed data.
```bash
go run cmd/migrate/main.go up
```

### 3. Start the API Server
```bash
go run cmd/api/main.go
```
The server will start on `localhost:8080`.

### 4. Start the Simulator (Optional)
Simulates 5 shoppers moving around Austin, TX.
```bash
go run cmd/simulator/main.go
```

## 🧪 Verification & Testing

### 1. Connect a WebSocket Client (The "Shopper")
Use `websocat` (or any WS client) to listen as a specific shopper.
```bash
# Listening as Shopper "...33"
websocat "ws://localhost:8080/ws?user_id=a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a33"
```

### 2. Create an Order (The "Customer")
In a new terminal:
```bash
curl -X POST http://localhost:8080/api/v1/orders \
  -H "Content-Type: application/json" \
  -d '{
    "customer_id": "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11",
    "store_id": "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a22",
    "total_cents": 5500,
    "items": [{"name": "Milk", "qty": 1}]
  }'
```
**Result**: The WebSocket terminal should immediately receive a `NEW_OFFER` JSON payload.

### 3. Claim the Order (Concurrency Test)
Copy the `id` from the created order.

**Shopper 1 Claims (Success):**
```bash
curl -X POST http://localhost:8080/api/v1/orders/YOUR_ORDER_ID/claim \
  -H "Content-Type: application/json" \
  -d '{"shopper_id": "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a33"}'
```

**Shopper 2 Claims (Fail - 409 Conflict):**
```bash
curl -X POST http://localhost:8080/api/v1/orders/YOUR_ORDER_ID/claim \
  -H "Content-Type: application/json" \
  -d '{"shopper_id": "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a34"}'
```

## 🧩 Engineering Challenges & Solutions

### 1. The "Ghost Car" Problem (Session Persistence)
**Challenge**: In a stateless simulator, if a courier refreshed their page, the simulator would forget they existed, leaving them stranded on the map.
**Solution**: Implemented a **Presence Protocol**. On mount, the frontend announces `SHOPPER_ONLINE` to the backend. The Simulator listens for this event and dynamically "spawns" or "adopts" the existing courier, resuming their state and route exactly where they left off.

### 2. Atomic Order Claiming (Concurrency)
**Challenge**: High-concurrency environments risk double-booking if two shoppers claim an order simultaneously.
**Solution**: Leveraged PostgreSQL's pessimistic locking (`SELECT ... FOR UPDATE`).
```sql
BEGIN;
SELECT status FROM orders WHERE id = $1 FOR UPDATE;
-- Check status (must be OFFERED)
UPDATE orders SET status = 'CLAIMED', shopper_id = $2 WHERE id = $1;
COMMIT;
```
This ensures that even with millisecond-difference requests, only one transaction succeeds.

### 3. Real-Time State Sync (The "Split Brain" Fix)
**Challenge**: Keeping the Courier's UI, the Customer's receipt, and the Simulator in sync without polling.
**Solution**: A consolidated **WebSocket Bridge**.
*   **Source of Truth**: The Backend (PostgreSQL) is the only state authority.
*   **Propagation**: Order updates publish to Kafka (`orders.lifecycle`).
*   **Consumption**: A specific WebSocket consumer reads Kafka and pushes generic JSON payloads to the frontend.
*   **Optimistic UI**: The frontend updates locally on user action, but reverts if the WebSocket confirmation doesn't arrive, ensuring perceived performance without data corruption.

## 💻 Frontend Architecture (Next.js)

The `web/` directory contains a sophisticated **Next.js 15** application designed for speed and interactivity.

*   **Tech Stack**: Next.js, React Query, TailwindCSS, Lucide Icons.
*   **State Management**: `React Query` for server state, `React Context` for Auth.
*   **Map Integration**: Custom Google Maps wrapper handling markers, directions service, and real-time movement interpolation.
*   **Optimistic Updates**: Immediate UI feedback for actions like "Claim Order", backed by eventual consistency via WebSocket updates.

## 📂 Project Structure

Verified Clean Architecture (Port & Adapter Pattern):

```
├── cmd
│   ├── api          # Main REST API entrypoint
│   └── simulator    # Autonomous Shopper Simulator
├── internal
│   ├── core         # Pure Domain Logic (No external dependencies)
│   │   ├── entity   # Domain Models (Order, User, Store)
│   │   ├── ports    # Interfaces (Repository & Service definitions)
│   │   └── service  # Business Logic implementation
│   └── adapters     # Infrastructure Implementation
│       ├── handler  # HTTP Handlers (Gin)
│       ├── repository # Postgres & Redis implementations
│       └── event    # Kafka Producers & Consumers
├── web              # Next.js Frontend Monorepo
│   ├── src/app      # App Router (Shopper & Customer views)
│   ├── src/components # Reusable UI & Map Components
│   └── src/lib      # API Clients & Utilities
└── migrations       # Database Schema Versioning
```

## 📝 License
MIT
