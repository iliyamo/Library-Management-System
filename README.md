# Library Management System

## 📚 Overview

The **Library Management System (LMS)** is a full‑stack back‑end service built in Go that powers a digital library. It lets users register, authenticate using JSON Web Tokens (JWT), browse authors and books, and asynchronously borrow or return books. Data is stored in **MySQL** via GORM, **Redis** provides caching and per‑user rate‑limit state, and **RabbitMQ** handles reliable message delivery with a Redis Pub/Sub fall‑back. The HTTP API is served by the Echo web framework and adheres to REST semantics.

Unlike a naive CRUD service, the LMS has been designed with scalability, concurrency and fault tolerance in mind. Expensive operations—such as updating stock when a book is borrowed—are decoupled from the request/response cycle by publishing commands to a message queue. A background worker consumes these commands, performs validations and updates the database, and emits events that other services can react to. Rate limiting and caching are implemented in Redis, and local semaphores protect against race conditions within a single instance.

## 🏗 Architecture

### High‑level architecture

At a glance, the LMS consists of four core layers: a stateless API layer, a persistence layer, a caching/queue layer and a background worker. The following Mermaid diagram shows how these parts interact:

```mermaid
graph TD
  client[User / Admin] -->|HTTP requests| api[API Server]
  api --> jwt[JWT & Rate Limiter]
  jwt --> crud
  crud -->|ORM| mysql[(MySQL)]
  crud -->|Cache / tokens| redis[(Redis)]
  crud -->|Publish borrow/return| mq[(RabbitMQ)]
  mq --> worker[Worker Process]
  worker --> mysql
  worker -->|Publish events| mq
  mq -.->|Fallback| redis
```

* **API Server**: Implements REST endpoints for authentication, authors, books, users and loans. Each request passes through JWT authentication and a token‑bucket rate limiter.
* **MySQL (GORM)**: Stores persistent state—users, authors, books, loans, refresh tokens—and enforces relations.
* **Redis**: Caches frequently read data (book and author search results, user lists) and stores per‑user rate‑limit buckets. It also acts as a fall‑back message broker when RabbitMQ is unavailable.
* **RabbitMQ**: Provides durable queues (`loan_commands` and `loan_events`) with at‑least‑once delivery. Commands are acknowledged only after successful processing; otherwise they are re‑queued. When RabbitMQ is down, Redis Pub/Sub is used, which offers at‑most‑once semantics.
* **Worker**: A background process that consumes commands (`borrow` and `return`) from the queue, performs validations and inventory updates inside the database, and publishes events describing the outcome.

### Domain flows and diagrams

Detailed sequence and domain diagrams are stored under the `Docs/` directory as **.mmd** (Mermaid) source files and their **.svg** renderings. These diagrams capture the behaviour of individual subsystems. For example:

#### Authentication

![Authentication flow](/Docs/Auth.svg)

The authentication flow shows registration, login, refresh, profile and logout endpoints. Passwords are salted and hashed with bcrypt, JWTs are signed using a secret key, and refresh tokens are stored in the database with an expiry column to prevent reuse. A valid refresh token can be exchanged for a new access token.

#### Borrowing and returning

![Borrowing sequence](/Docs/Borrowing.svg)

Borrow and return requests are enqueued rather than handled synchronously. The API validates input and publishes a `CmdBorrow` or `CmdReturn` message to the `loan_commands` queue. The worker atomically decreases or increases `AvailableCopies`, creates a `Loan` record and acknowledges the message. Events such as `LoanRequested` or `LoanReturned` are published to the `loan_events` queue for monitoring or notifications.

#### Catalog search

![Catalog search](/Docs/Catalog_Search.svg)

Search endpoints support full‑text queries and cursor‑based pagination. Each unique combination of search term, cursor and limit is cached under a composite key in Redis for a short TTL. Invalidating the cache happens after write operations (e.g. book update).

> **Tip:** The original Mermaid definitions can be found in `/Docs/*.mmd` if you need to modify or re‑render these diagrams. For example, `/Docs/book.mmd` describes the relationships between `users`, `books` and `loans` tables.

### Additional UML diagrams

The project includes comprehensive flowcharts for **author** and **book** operations. These UML diagrams describe the control flow for create, read, update and delete (CRUD) endpoints, highlight middleware checks and show typical HTTP status codes returned under different conditions. Embedding the diagrams here helps new contributors quickly understand the expected behaviour.

#### Author operations

![Author flow](/Docs/Author.svg)

This diagram covers the lifecycle of an author entity. When creating an author (`POST /authors`), the API verifies the JWT token via the authentication middleware; missing or invalid tokens result in a **401 Unauthorized** error. If the token is valid, the rate‑limiter middleware enforces per‑user limits: too many requests return **429 Too Many Requests**. The request body is then validated; malformed input yields **400 Bad Request**. Duplicate authors cause **409 Conflict**, and successful creation returns **201 Created**. Listing authors (`GET /authors`) and fetching a single author (`GET /authors/{id}`) follow similar steps—checking authentication, rate limits and validating parameters—returning **200 OK** on success, **404 Not Found** if the author does not exist, and caching search results in Redis for 30 seconds. Updating and deleting authors (`PUT`/`DELETE /authors/{id}`) reuse these checks and return **200** or **204** depending on the operation.

#### Book operations

![Book flow](/Docs/Book.svg)

The book flowchart illustrates creating a book (`POST /books`), listing/searching (`GET /books`), fetching details (`GET /books/{id}`), updating (`PUT /books/{id}`), deleting (`DELETE /books/{id}`) and increasing copies (`POST /books/{id}/increase`). Each step goes through the JWT middleware and rate limiter. Invalid or missing tokens lead to **401** responses; exceeding the rate limit produces **429**. Input validation errors return **400**, duplicate ISBNs yield **409**, and unknown book IDs result in **404**. Successful creation returns **201 Created**, while successful updates and deletions return **200 OK** with a JSON message. When creating or increasing copies, `AvailableCopies` is initialised or incremented along with `TotalCopies`, and the semaphore in the handler serialises updates to prevent race conditions.

### Data model

The database schema resides in `Docs/digital_library.sql`. Key tables include:

| Table            | Purpose                                                    |
|------------------|------------------------------------------------------------|
| **users**        | Registers library members; stores full name, email, role, and a bcrypt password hash. |
| **authors**      | Contains author profiles (name, biography, birth_date).    |
| **books**        | Records metadata of each book, including total and available copies, linked to an author (and optional category). |
| **loans**        | Tracks borrowings; each loan has a `user_id`, `book_id`, loan and due dates, optional return date and a status enum (`borrowed`, `returned`). |
| **refresh_tokens** | Stores refresh JWTs along with the associated `user_id`, `expires_at` timestamp and `created_at`. Tokens are invalidated on logout or expiry. |

Foreign keys ensure referential integrity (e.g. `loans.user_id` → `users.id`), and appropriate indexes are defined on frequently queried columns (`email`, `isbn`, `author_id`, `user_id`).

## 📂 Project structure

The repository follows a *clean architecture* approach: handlers handle HTTP concerns, services implement business logic, repositories abstract database operations and caching, and queue packages hide messaging details. A high‑level layout is shown below:

```
Library-Management-System/
├── cmd/
│   └── server/            # entry point (main.go) and RabbitMQ log
├── internal/
│   ├── config/            # environment configuration (e.g. Init MySQL, Redis, RabbitMQ)
│   ├── database/          # DB initialisation with GORM
│   ├── handler/           # HTTP handlers (auth, author, book, user, loan)
│   ├── middleware/        # JWT authentication, per‑user rate limiting
│   ├── model/             # Data models and request/response DTOs
│   ├── queue/             # Messaging abstractions (RabbitMQ and Redis Pub/Sub)
│   ├── repository/        # CRUD repositories with caching & search logic
│   ├── router/            # Route grouping and middleware wiring
│   ├── service/           # Business services (e.g. LoanService with domain logic)
│   └── utils/             # Helpers: JWT generation/validation, password hashing, Redis client
├── Docs/                  # Mermaid diagrams and SQL schema
├── docker-compose.yml     # Multi‑service development environment (app + MySQL + Redis + RabbitMQ)
├── Dockerfile             # Builds the Go application for production
└── .env.example           # Example environment configuration
```

## ⚙️ Setup and running

### Prerequisites

You can either use the provided Docker Compose setup or run services manually. For the latter you need:

* **Go 1.21+** (only required outside Docker).
* **MySQL 8+** – persistence layer for the library.
* **Redis 6+** – caching and rate‑limiting storage (and optional message broker).
* **RabbitMQ 3.8+** – durable message broker for borrow/return commands and events.

### Environment variables

The application reads configuration from environment variables. Copy `.env.example` to `.env` and adjust values to suit your environment. Important variables include:

| Variable           | Description                                         | Default |
|--------------------|-----------------------------------------------------|---------|
| `DB_USER`          | MySQL user                                          | `root` |
| `DB_PASSWORD`      | MySQL password                                      | `password` |
| `DB_HOST`          | MySQL host (e.g. `db` when using Docker Compose)    | `db` |
| `DB_PORT`          | MySQL port                                          | `3306` |
| `DB_NAME`          | Database name                                       | `digital_library` |
| `REDIS_ADDR`       | Address of Redis server                             | `redis:6379` |
| `REDIS_PASSWORD`   | Redis password (leave empty if none)                | (empty) |
| `RABBITMQ_URL`     | AMQP connection URI for RabbitMQ                    | `amqp://guest:guest@rabbitmq:5672/` |
| `JWT_SECRET`       | Secret used to sign access and refresh tokens       | `change_me_dev_only` |
| `APP_PORT`         | TCP port where the API listens                      | `8080` |

### Running with Docker Compose

To launch all dependencies together:

```bash
git clone <your-fork>
cd Library-Management-System
cp .env.example .env   # customise variables as needed
docker-compose up --build
```

This spins up the API server, a MySQL container seeded with `Docs/digital_library.sql`, a Redis server and a RabbitMQ broker with its management UI exposed at `http://localhost:15672` (user `guest`/`guest`). The API will be accessible at `http://localhost:8080/api/v1`.

### Running without Docker

If you prefer to run services manually, set up MySQL, Redis and RabbitMQ on your machine or in your cloud environment. Export the environment variables above, then build and start the server:

```bash
go mod download   # fetch dependencies
go run ./cmd/server
```

Ensure that the database schema is applied by running the SQL in `Docs/digital_library.sql`.

## 🌐 API surface

All endpoints are under `/api/v1`. Endpoints marked **(Auth)** require a valid access token in the `Authorization: Bearer <token>` header and are subject to rate limiting. Below is a summary—see the handler code under `internal/handler/` for full details.

### Authentication & users

| Method & path             | Description | Notes |
|---------------------------|-------------|-------|
| `POST /auth/register`     | Register a new user (full_name, email, password) | Generates hashed password; returns 201 on success. |
| `POST /auth/login`        | Log in with email & password | Returns short‑lived access token and long‑lived refresh token. |
| `POST /auth/refresh`      | Exchange a refresh token for a new access token | Refresh tokens are stored in the `refresh_tokens` table and validated. |
| `POST /auth/logout`       | Invalidate all of a user’s refresh tokens | **(Auth)** |
| `GET /auth/profile`       | Get profile of the currently authenticated user | **(Auth)** |
| `GET /users`              | Search users by full name or email with cursor‑based pagination | **(Auth)** |

### Authors

| Method & path             | Description | Notes |
|---------------------------|-------------|-------|
| `POST /authors`           | Create a new author | **(Auth)** |
| `GET /authors`            | List authors; supports search by name, cursor and limit | **(Auth)** |
| `GET /authors/{id}`       | Get an author by ID | **(Auth)** |
| `PUT /authors/{id}`       | Update an existing author | **(Auth)** |
| `DELETE /authors/{id}`    | Delete an author | **(Auth)** |

### Books

| Method & path                      | Description | Notes |
|------------------------------------|-------------|-------|
| `POST /books`                      | Create a new book record | Initial `AvailableCopies` = `TotalCopies`. **(Auth)** |
| `GET /books`                       | Search and list books with full‑text search and pagination | Cached in Redis. **(Auth)** |
| `GET /books/{id}`                  | Get book details by ID | **(Auth)** |
| `PUT /books/{id}`                  | Update a book’s metadata and total copies | Serialised via a semaphore to avoid races. **(Auth)** |
| `DELETE /books/{id}`               | Delete a book | **(Auth)** |
| `POST /books/{id}/increase`        | Increase total and available copies by a quantity > 0 | Uses the same semaphore. **(Auth)** |

### Loans

| Method & path                      | Description | Notes |
|------------------------------------|-------------|-------|
| `POST /loans`                      | Request to borrow a book (`book_id`, optional `days`) | Publishes a `CmdBorrow` to the queue and returns `202 Accepted`. **(Auth)** |
| `GET /loans` / `GET /loans/my`     | View current and past loans for the current user | **(Auth)** |
| `POST /loans/{id}/return`          | Request to return a loan by ID | Publishes a `CmdReturn` message and returns `202 Accepted`. **(Auth)** |

### Endpoint behaviour and status codes

While the tables above describe what each endpoint does, the diagrams in `Docs/` also encode the typical HTTP status codes that may be returned. A brief summary:

- **Authentication & users**
  - `POST /auth/register` returns **201 Created** on success, **400 Bad Request** for invalid input, **409 Conflict** if the email is already used and **500 Internal Server Error** for database errors.
  - `POST /auth/login` returns **200 OK** with tokens or **401 Unauthorized** for invalid credentials.
  - `POST /auth/refresh` returns **200 OK** with a new access token or **401** if the refresh token is invalid or expired.
  - `POST /auth/logout` always returns **200 OK**; missing or invalid tokens produce **401**.
  - `GET /auth/profile` and `GET /users` return **200 OK** on success, **401** if authentication fails and **429 Too Many Requests** when the rate limit is exceeded.
- **Authors**
  - `POST /authors` returns **201 Created** on success, **400 Bad Request** for malformed input, **409 Conflict** if the author already exists, **401** for missing/invalid tokens, **429** for rate limiting, and **500** for server errors.
  - `GET /authors` and `GET /authors/{id}` return **200 OK** on success; **404 Not Found** if the author does not exist; **401** or **429** for authentication or rate‑limit failures.
  - `PUT`/`DELETE /authors/{id}` return **200 OK** or **204 No Content** on success, **404** if the author is missing, **400** for invalid parameters and the usual **401/429/500** codes.
- **Books**
  - `POST /books` returns **201 Created** on success, **400 Bad Request** for invalid JSON, **409 Conflict** if the ISBN already exists, and the standard **401/429/500** codes.
  - `GET /books` and `GET /books/{id}` return **200 OK**, **404 Not Found** for missing IDs, and may return **401/429** for authentication or rate limiting.
  - `PUT /books/{id}` and `POST /books/{id}/increase` return **200 OK** on success and **400/404/401/429** for invalid input, missing records or rate‑limit failures.
  - `DELETE /books/{id}` returns **200 OK** when a book is removed or **404** if it does not exist.
- **Loans**
  - `POST /loans` and `POST /loans/{id}/return` return **202 Accepted** immediately after enqueuing the command. Invalid IDs or missing parameters result in **400 Bad Request**. Authentication or rate‑limit failures produce **401** or **429**.
  - `GET /loans` and `GET /loans/my` return **200 OK** with the current user’s loan history, **401** if authentication fails and **429** if the rate limit is exceeded.

These status codes are not exhaustive—the API may return other standard HTTP codes (e.g. **500** on unexpected errors) as indicated by the flowcharts in `Docs/*.mmd`.

## 🧠 Concurrency and race conditions

### Token‑bucket rate limiter

A simple token‑bucket algorithm is implemented in `internal/middleware/rate_limit.go`. Each user has a bucket with a capacity of 20 tokens and a refill rate of one token every 3 seconds. On each request the current state is read from Redis, tokens are refilled based on elapsed time, a token is consumed, and the new state is written back. Because the read/update/write cycle is not atomic, two concurrent API instances might both read the same bucket, each decide that a request is allowed, and decrement tokens concurrently—causing the rate limit to be exceeded. In distributed environments this can be solved by:

* Using a Lua script executed on Redis (`EVAL`) to perform the refill, check and decrement atomically.
* Employing Redis transactions (`WATCH`/`MULTI`) to detect concurrent modifications.
* Centralising the rate limiter behind a single instance or using a third‑party service.

### Updating book inventory

When multiple borrow/return operations or administrative updates are processed concurrently, inventory counts must not become negative or diverge. Within a single API instance the handlers wrap the critical sections in a **semaphore** (`bookOpCh <- struct{}{}`) so only one goroutine can update `TotalCopies` or `AvailableCopies` at a time. In a horizontally scaled setup this is insufficient—use a **distributed lock** (e.g. Redlock or a database advisory lock) to ensure that only one instance modifies a book at any given moment.

### Messaging semantics

Commands are published to RabbitMQ with persistent delivery and consumed by the worker. On success the worker calls `Ack`; on transient failures it calls `Nack` with re‑queue enabled. This yields **at‑least‑once** processing: a message may be delivered multiple times but will not be lost. When RabbitMQ is unreachable, the system falls back to Redis Pub/Sub. Redis Pub/Sub provides **at‑most‑once** delivery—messages are broadcast to subscribers but are not persisted. Consumers must therefore be stateless and idempotent to handle duplicate or missing messages gracefully. For mission‑critical operations, consider implementing the **Transactional Outbox** pattern: write pending messages to a `outbox` table in the same database transaction that updates the domain state, then have a separate process publish them to RabbitMQ.

### Refresh token storage

Refresh tokens are stored in the `refresh_tokens` table with a non‑NULL `expires_at` column. If you run MySQL in **STRICT** mode and omit this value, inserts will fail silently and tokens will be invalidated immediately. The repository layer therefore explicitly sets `expires_at` to **7 days after creation** when storing a token. Refresh tokens are deleted on logout or on expiry (the `Validate` method checks presence in the database). For a more secure design, refresh tokens could be hashed before storage, rotated on each use, and pruned by a background job.

### Pagination

All list endpoints (authors, books, users, loans) implement **cursor‑based pagination**. Clients pass an optional `cursor` query parameter (the ID of the last item in the previous page) and a `limit` parameter (how many items to return; defaults to 10 and is capped at 100). The API responds with a `next_cursor` field containing the ID to use for the next page. If `next_cursor` is `0`, there are no more records. Cursor values allow stable pagination even when new items are inserted or removed between requests, and they are included in the Redis cache key to avoid mixing pages.

## 📦 Caching, search and Redis

Redis acts both as a cache and a queue fallback:

* **Search caching**: Search results for books, authors and users are stored under keys that include the query string, cursor and limit. Cached entries expire after 30 seconds to strike a balance between performance and freshness. Handlers invalidate related keys upon writes (e.g. after creating, updating or deleting a book) to avoid serving stale data.
* **Rate limiting**: Each user’s token bucket is stored as a JSON blob keyed by `rate:user:{user_id}` with a TTL computed from the bucket size and refill rate.
* **Fallback broker**: If RabbitMQ cannot be reached on startup, the system uses Redis Pub/Sub to publish and consume loan commands. Because Redis Pub/Sub is not persistent, messages will be lost if no consumer is subscribed at the time of publishing. Monitoring the presence of consumers is therefore important.

## 🔐 Security considerations

Security is woven throughout the LMS:

* **Password hashing**: User passwords are hashed with bcrypt (default cost) before storage. Plaintext passwords never leave memory.
* **JWT tokens**: Access tokens expire after 15 minutes; refresh tokens expire after 7 days. The signing key is loaded from `JWT_SECRET` and cached via `sync.Once`.
* **Refresh token blacklisting**: Refresh tokens are persisted in MySQL, so they can be invalidated on logout or after a breach. Only tokens present in the database and not expired can be exchanged for new access tokens.
* **Role IDs**: A `role_id` claim is embedded in JWTs. Currently only one role (`member`) is used, but the claim makes it easy to add admin‑only endpoints in the future.
* **Environment configuration**: Secrets such as DB passwords and JWT secrets should be provided via environment variables or a secrets manager. Do not commit real secrets to version control.
* **HTTPS and transport security**: In production, terminate TLS at a proxy (e.g. nginx, Caddy) and configure secure connections to MySQL, Redis and RabbitMQ.

## 🧭 Notes & improvements

This project is a learning exercise demonstrating how to structure a Go back‑end with messaging and caching. Possible improvements include:

* **Atomic rate limiting:** Replace the current read–modify–write cycle in Redis with a Lua script or Redis transactions to eliminate races.
* **Distributed locking:** Use a distributed mutex (Redlock) for book inventory updates or adopt optimistic locking in the database (`version` columns).
* **Transactional outbox:** Persist pending messages in the same transaction as domain updates and use a separate publisher to achieve exactly‑once semantics.
* **Monitoring and metrics:** Expose Prometheus metrics (request rate, queue lengths, processing latencies) and integrate distributed tracing (OpenTelemetry) to identify bottlenecks.
* **Testing:** Add unit and integration tests for handlers, repositories and middleware. Use testcontainers to spin up MySQL, Redis and RabbitMQ in CI.
* **Role‑based authorisation:** Extend the `role_id` claim to enforce admin‑only endpoints (e.g. for approving loans or managing users).
* **Front‑end:** Build a React/Vue/Angular interface that consumes this API.

If you have questions, discover bugs or want to contribute new features, please open an issue or submit a pull request. Enjoy exploring, modifying and extending the **Library Management System**!
