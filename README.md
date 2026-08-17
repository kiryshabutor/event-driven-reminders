# JWT Auth & Reminders

A Go microservices project implementing JWT authentication and a deferred reminders system.

## Project Goals

The primary goal is hands-on learning and demonstration of modern microservice development in Go.
Key areas:

- **Architecture**: Building a scalable microservice architecture.
- **Security**: Reliable JWT-based authentication with Access + Refresh token pairs.
- **Asynchrony**: Message broker (Apache Kafka) for deferred tasks and event-driven workflows.
- **DevOps**: Containerization and orchestration with Docker and Docker Compose.

## Features

- **Authentication & Authorization**: JWT issuance and validation (Access & Refresh tokens).
- **Microservices**:
  - **Auth Service**: gRPC service for registration, login, and token management.
  - **Reminder Service**: Create and manage reminders with transactional outbox pattern.
  - **Analytics Service**: Processes lifecycle events from Kafka with exactly-once semantics.
  - **Notification Service**: Kafka consumer that handles reminder notifications.
  - **API Gateway**: Single REST entry point proxying requests to internal gRPC services.
- **Asynchronous Notifications**: Kafka topics for reminder lifecycle and notification delivery.
- **Exactly-Once Delivery**: Guaranteed single-event processing in Analytics Service via an idempotency table.
- **Storage**: PostgreSQL (primary data), Redis (caching / token blacklist).

## Exactly-Once Delivery

Analytics Service implements the **Exactly-Once** pattern for processing Kafka events:

1. Each event carries a unique `EventID` (UUID v7).
2. On processing, the service checks the `analytics.processed_events` table.
3. If the event was already processed, it is skipped (idempotency).
4. All logic runs inside a single SQL transaction.

```
┌─────────────┐    ┌─────────────┐    ┌──────────────────────┐
│   Kafka     │───▶│  Consumer   │───▶│  ProcessEvent()      │
│  (message)  │    │  (fetch)    │    │  ┌────────────────┐  │
└─────────────┘    └─────────────┘    │  │ BEGIN TX       │  │
                                     │  │ Check EventID  │  │
                                     │  │ Update Stats   │  │
                                     │  │ Save EventID   │  │
                                     │  │ COMMIT         │  │
                                     │  └────────────────┘  │
                                     └──────────────────────┘
```

## Tech Stack

- **Language**: Go (Golang)
- **Database**: PostgreSQL
- **Cache**: Redis
- **Message Broker**: Apache Kafka
- **Communication**: gRPC (Protobuf), REST (HTTP/JSON)
- **Containerization**: Docker, Docker Compose

## Getting Started

### Prerequisites

- Go 1.25+
- Docker & Docker Compose

### Run with Docker Compose (Recommended)

The easiest way to run all components.

1. Clone the repository:
   ```bash
   git clone https://github.com/kiryshabutor/JWT.git
   cd JWT
   ```

2. Create a configuration file:
   ```bash
   cp .env.example .env
   ```
   Edit `.env` if needed (passwords, ports).

3. Start the services:
   ```bash
   docker compose up --build
   ```

This brings up: PostgreSQL, Redis, Zookeeper, Kafka, Auth Service, Reminder Service, Analytics Service, Notification Service, and API Gateway.

## Configuration

All settings live in `.env`. See `.env.example` for a template.

Key variables:

- `DB_*`: PostgreSQL connection settings.
- `JWT_SECRET`: Secret key for signing tokens. **Change this in production!**
- `GRPC_PORT`: gRPC service ports.
- `KAFKA_BROKERS`: Kafka broker addresses.

## Project Structure

The project is organized as a multi-module monorepo:

```
.
├── go.work                  # workspace linking all modules
├── shared/                  # shared Go module
│   ├── consts/              # event-type and status constants
│   ├── errorsx/             # cross-service error type
│   ├── logger/              # slog-based logger setup
│   ├── proto/               # protobuf contracts + generated code
│   │   ├── auth/
│   │   ├── reminder/
│   │   └── analytics/
│   └── types/               # wire types (LifecycleEvent, Reminder)
├── services/
│   ├── auth/                # authentication & token management
│   ├── reminder/            # reminders, outbox, background workers
│   ├── analytics/           # lifecycle event processing
│   ├── notification/        # Kafka notification consumer
│   └── gateway/             # HTTP API Gateway & gRPC clients
├── build/                   # Dockerfiles per service
├── docker/                  # Postgres init script
└── docker-compose.yml
```

Each service has its own `go.mod` and follows a layered layout:

```
service/
├── cmd/            # thin entry point → app.Run()
├── config/         # service-specific config
├── app/            # composition root (db, grpc, kafka, redis, run.go)
├── entity/
│   ├── service/    # domain types (no persistence tags)
│   └── repository/ # DB projections (GORM-tagged)
├── handler/        # inbound: gRPC, HTTP, Kafka consumers
├── gateway/        # outbound: gRPC clients, Kafka producers
├── service/        # business logic
├── repository/     # interface + implementation (pg, redis)
└── worker/         # background processes (reminder only)
```

Local development links modules through the root `go.work` and `replace` directives pointing at `../../shared`.

## API Endpoints

The API is accessible through the **API Gateway** (default port `8080`).

| Method   | Path                 | Auth | Description              |
| -------- | -------------------- | ---- | ------------------------ |
| `POST`   | `/auth/register`     | -    | Register a new user      |
| `POST`   | `/auth/login`        | -    | Login and get tokens     |
| `POST`   | `/auth/refresh`      | -    | Refresh token pair       |
| `POST`   | `/auth/logout`       | Bearer | Logout (revoke token)  |
| `GET`    | `/auth/profile`      | Bearer | Get current user profile |
| `POST`   | `/reminders`         | Bearer | Create a reminder      |
| `GET`    | `/reminders`         | Bearer | List reminders         |
| `GET`    | `/reminders/:id`     | Bearer | Get a reminder by ID   |
| `PUT`    | `/reminders/:id`     | Bearer | Update a reminder      |
| `DELETE` | `/reminders/:id`     | Bearer | Delete a reminder      |
| `GET`    | `/analytics/me`      | Bearer | Get user statistics    |
| `GET`    | `/health`            | -    | Health check             |

**[Full API documentation](docs/API.md)**
