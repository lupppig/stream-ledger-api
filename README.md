# StreamLedger API

StreamLedger API is a secure, scalable, and well-tested transactional API for managing user wallets and their transactions. It emphasizes **data integrity**, **idempotency**.

---

## Quick Start

1. Clone the repository:

```bash
git clone https://github.com/lupppig/stream-ledger-api.git
cd stream-ledger-api
```

2. Copy environment configuration:

```bash
cp .env.example .env
```

3. run Postgres and Kafka:

```bash
make kafka-up
make postgres-up
```

5. Run the API:

```bash
make run
```

6. **Start the test Postgres container before running tests**:

```bash
make postgres-test
make test
```

7. Stop the test container after tests:

```bash
make postgres-test-down
```

> **Note:** `make postgres-test` must be run before `make test`, otherwise tests will fail.

---

## Objective

The objective of this project is to provide a reliable JSON API for managing user wallets and processing transactions with strong guarantees around **atomicity**, **consistency**, and **idempotency**.

---

## Core Features

1. **Authentication**

   * User registration and login with secure session handling.
   * Endpoints protected and accessible only by authenticated users.

2. **API Endpoints**

   * Versioned endpoints: `/api/v1/...`
   * `GET /api/v1/wallet`: Retrieve authenticated user wallet details.
   * `POST /api/v1/transactions`: Create a new transaction (credit/debit). Prevents negative balances.
   * `GET /api/v1/transactions`: List user transactions with pagination.
   * `POST /api/v1/transactions/export`: Export transaction history to Excel asynchronously via RiverQueue.

3. **Transaction Guarantees**

   * Atomic operations for wallet creation and transaction updates.
   * Idempotency with `trans_id` ensures safe retries without duplicate transactions.

4. **Event Streaming**

   * Successful transactions produce messages to Kafka topic `transactions`.
   * Payload includes `user_id`, `entry`, `amount`, `balance`, `timestamp`.

5. **Background Jobs**

   * Handled by RiverQueue with retry, concurrency control, and durability.

6. **Testing**

   * Covers wallet operations, idempotency enforcement, Kafka events, and transaction correctness.

---

## Technologies

* **Golang**: Core API and job processing
* **PostgreSQL**: Database with ACID guarantees
* **Kafka**: Transaction event streaming
* **RiverQueue**: Background jobs for Excel export

---

## Setup and Installation

### Prerequisites

* Go 1.22+
* Docker
* Make

---

## Database Schema

### Users Table

| Column     | Type      | Constraints                         |
| ---------- | --------- | ----------------------------------- |
| id         | BIGINT    | Primary Key, Auto Increment         |
| first_name | TEXT      | NOT NULL                            |
| last_name  | TEXT      | NOT NULL                            |
| email      | TEXT      | UNIQUE, NOT NULL                    |
| password   | TEXT      | NOT NULL                            |
| created_at | TIMESTAMP | Default current_timestamp, NOT NULL |

**Relationships:** 1:1 → Wallet

---

### Wallets Table

| Column     | Type      | Constraints                         |
| ---------- | --------- | ----------------------------------- |
| id         | BIGINT    | Primary Key, Auto Increment         |
| user_id    | BIGINT    | UNIQUE, NOT NULL, FK → users.id     |
| balance    | BIGINT    | NOT NULL, Default 0                 |
| currency   | TEXT      | NOT NULL, Default 'NGN'             |
| created_at | TIMESTAMP | Default current_timestamp, NOT NULL |
| updated_at | TIMESTAMP | Default current_timestamp, NOT NULL |

**Relationships:** 1:1 ← User, 1:N → Transactions

---

### Transactions Table

| Column     | Type      | Constraints                 |
| ---------- | --------- | --------------------------- |
| id         | BIGINT    | Primary Key, Auto Increment |
| wallet_id  | BIGINT    | NOT NULL, FK → wallets.id   |
| entry      | ENUM      | NOT NULL (credit / debit)   |
| amount     | BIGINT    | NOT NULL                    |
| trans_id   | TEXT      | UNIQUE, NOT NULL            |
| created_at | TIMESTAMP | Default current_timestamp   |

**Relationships:** N:1 ← Wallet

---

### ER Diagram

```
+---------+        1 : 1        +---------+        1 : N        +---------------+
|  Users  |-------------------->| Wallets |-------------------->| Transactions  |
+---------+                     +---------+                     +---------------+
| id (PK) |<----------------+   | id (PK) |                     | id (PK)       |
| email   |                 |   | user_id | (FK → users.id)     | wallet_id (FK)|
| name    |                 |   | balance |                     | entry         |
| ...     |                 |   | currency|                     | amount        |
+---------+                 |   | ...     |                     | trans_id (UK) |
                            |   +---------+                     | created_at    |
                            |                                   +---------------+
                            |
                            |  (each user has exactly one wallet)
```

---

### Postman Collection

[StreamLedger API Postman Workspace](https://www.postman.com/movie4-7051/workspace/shop-lyft/collection/29589431-d1cf921e-98fc-437b-82b4-8d7e02571732?action=share&creator=29589431)

---

## Testing

```bash
make postgres-test   # Start test database
make test            # Run all tests
make postgres-test-down  # Stop test database
```

Tests cover:

* Wallet updates and negative balance prevention
* Transaction creation and idempotency
* Kafka event production
* Background job exports

---

## Conclusion

StreamLedger API ensures **robust, correct, and scalable transaction handling**, combining Go, PostgreSQL, Kafka, and RiverQueue to handle retries, race conditions, and high transaction volumes.
