# Relay

A Discord-inspired messaging / communication platform built as a microservices system. This is a personal learning project focused on distributed systems, inter-service communication, backend engineering and all the involved tech.

## Architecture

Client-facing REST API with a WebSocket gateway for real-time events. Services communicate synchronously via gRPC and asynchronously via NATS.

```
Client
  │
  ▼
Traefik (API Gateway + Forward Auth)
  │
  ├── Auth Service        → issues JWT tokens, manages sessions
  ├── User Service        → user profiles, avatars
  ├── Guild Service       → guilds, channels
  ├── Message Service     → messages
  └── Gateway Service     → WebSocket, real-time events

Sync:   gRPC
Async:  NATS
```

Each service owns its own PostgreSQL database. No shared databases.

## System Diagram

![Alt text](/docs/diagrams/relay-microservices-system-diagram.png)

## Tech Stack

- Language: Go
- Framework: Gin
- Database: PostgreSQL (per service)
- Migrations: golang-migrate
- Query generation: sqlc
- API Gateway: Traefik
- Sync IPC: gRPC
- Async IPC: NATS
- Object Storage: MinIO (avatars, icons, attachments)
- IDs: Sonyflake
- Docker
- Python with pytest for E2E api tests
- Zap for logging
