# StreamLedger API

StreamLedger API is a secure, scalable, and well-tested transactional API for managing user wallets and their transactions. It is designed with strong emphasis on **data integrity**, **idempotency**, and **fault-tolerance** in distributed environments.

---

## Objective

The objective of this project is to provide a reliable JSON API for user wallet management and transaction processing with guarantees around **atomicity**, **consistency**, and **fault-tolerance**.

---

## Core Features

1. **Authentication**

   * User registration and login with secure session handling.
   * All endpoints protected and accessible only by authenticated users.

2. **API Endpoints**

   * Versioned endpoints: `/api/v1/...`
   * `GET /api/v1/wallet`: Retrieve authenticated user details and wallet balance.
   * `POST /api/v1/transactions`: Create a new transaction (credit or debit). Debits are disallowed if they result in negative balances.
   * `GET /api/v1/transactions`: List user transactions with pagination.
   * `POST /api/v1/transactions/export`: Export transaction history to Excel (asynchronously via RiverQueue).

3. **Transaction Guarantees**

   * **Atomic operations**: User and wallet creation are wrapped in a single transaction to ensure consistency.
   * **Atomic debit and wallet update**: Wallet balance updates and transaction creation occur in a single commit, preventing race conditions and partial writes.

4. **Background Jobs**

   * Handled by [RiverQueue](https://riverqueue.com/) (a distributed task queue for Go).
   * RiverQueue ensures **safe retries, concurrency control, and durability** — ideal for exporting large datasets such as transaction history.

5. **Event Streaming**

   * Every successful transaction produces a JSON message to a Kafka topic named `transactions`.
   * Payload includes: `user_id`, `entry`, `amount`, `balance`, `timestamp`.

6. **Idempotency with `trans_id`**

   * Each transaction request requires a unique `trans_id`.
   * Ensures network retries or duplicate requests do not cause double-spending.
   * The system will return the original transaction response if the same `trans_id` is replayed.

7. **Testing**

   * Comprehensive tests cover wallet operations, idempotency enforcement, and transaction correctness.

---

## Technologies

* **Golang**: Core API and job processing.
* **PostgreSQL**: Database with ACID guarantees.
* **Kafka**: Transaction event streaming.
* **RiverQueue**: Background jobs for Excel export.

---

## Design Choices and Trade-offs

* **Atomic operations over speed**: Prioritized **correctness and consistency** over raw performance.
* **Idempotency key (`trans_id`)**: Ensures strong guarantees against duplicate transactions without locking overhead.
* **RiverQueue**: Chosen for its Go-native design, reliable retry mechanism, and ability to scale concurrency safely.
* **Kafka**: Selected over simpler pub/sub tools for its durability, replay capability, and ability to support analytics/fraud detection downstream.

---

## Setup and Installation

### Prerequisites

* [Go 1.22+](https://go.dev/dl/)
* [Docker](https://www.docker.com/)
* [PostgreSQL](https://www.postgresql.org/)
* [Kafka](https://kafka.apache.org/)

### Installation

1. Clone the repository:

   ```bash
   git clone https://github.com/lupppig/stream-ledger-api.git
   cd stream-ledger-api
   ```

2. Copy environment configuration:

   ```bash
   cp .env.example .env
   ```

3. Start dependencies (Postgres, Kafka):

   ```bash
   make postgres-up
   ```

4. Start the application:

   ```bash
   make run
   ```

5. Run tests:

   ```bash
   make tests
   ```

*Note: Database migrations are handled automatically on application startup.*

---

## Database Schema

### **Users Table**

| Column     | Type      | Constraints                         | Description                            |
| ---------- | --------- | ----------------------------------- | -------------------------------------- |
| id         | BIGINT    | Primary Key, Auto Increment         | Unique user identifier                 |
| first_name | TEXT      | NOT NULL                            | User’s first name                      |
| last_name  | TEXT      | NOT NULL                            | User’s last name                       |
| email      | TEXT      | UNIQUE, NOT NULL                    | User’s email address                   |
| password   | TEXT      | NOT NULL                            | User’s hashed password (hidden in API) |
| created_at | TIMESTAMP | DEFAULT current_timestamp, NOT NULL | Record creation time                   |

**Relationships:**

* **1 : 1 → Wallet** (a user has exactly one wallet).

---

### **Wallets Table**

| Column     | Type      | Constraints                              | Description                     |
| ---------- | --------- | ---------------------------------------- | ------------------------------- |
| id         | BIGINT    | Primary Key, Auto Increment              | Unique wallet identifier        |
| user_id    | BIGINT    | UNIQUE, NOT NULL, Foreign Key → users.id | Each wallet belongs to one user |
| balance    | BIGINT    | NOT NULL, DEFAULT 0                      | Wallet balance in **kobo**      |
| currency   | TEXT      | NOT NULL, DEFAULT `'NGN'`                | Currency code (default NGN)     |
| created_at | TIMESTAMP | DEFAULT current_timestamp, NOT NULL      | Record creation time            |
| updated_at | TIMESTAMP | DEFAULT current_timestamp, NOT NULL      | Record last update time         |

**Relationships:**

* **1 : 1 ← User** (each wallet is linked to one user).
* **1 : N → Transactions** (a wallet can have multiple transactions).

---

### **Transactions Table**

| Column     | Type      | Constraints                        | Description                            |
| ---------- | --------- | ---------------------------------- | -------------------------------------- |
| id         | BIGINT    | Primary Key, Auto Increment        | Unique transaction identifier          |
| wallet_id  | BIGINT    | NOT NULL, Foreign Key → wallets.id | The wallet this transaction belongs to |
| entry      | ENUM      | NOT NULL (credit / debit)          | Type of transaction                    |
| amount     | BIGINT    | NOT NULL                           | Transaction amount in **kobo**         |
| trans_id   | TEXT      | UNIQUE, NOT NULL                   | Idempotency key for safe retries       |
| created_at | TIMESTAMP | DEFAULT current_timestamp          | Record creation time                   |

**Relationships:**

* **N : 1 ← Wallet** (many transactions belong to one wallet).

---

### ER Diagram (ASCII)

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
                            |
```

---

## API Documentation

### `POST /api/v1/auth/signup`

Registers a new user.

### `POST /api/v1/auth/login`

Logs in a user and returns authentication tokens.

### `GET /api/v1/wallet`

Retrieves wallet details and balance for the authenticated user.

### `POST /api/v1/transactions`

Creates a new transaction.

* Request body:

  ```json
  {
    "entry": "credit",
    "amount": 1000,
    "trans_id": "unique-key-123"
  }
  ```

### `GET /api/v1/transactions`

Lists all user transactions with pagination.

### `POST /api/v1/transactions/export`

Exports user transactions to Excel (asynchronously via RiverQueue).

---

## Testing

Run tests:

```bash
make tests
```

Tests cover:

* Transaction creation (credit/debit).
* Wallet updates and negative balance prevention.
* Idempotency with `trans_id`.
* Kafka event production.
* Export background jobs.

---

## Conclusion

StreamLedger API  handles wallet transactions with strong guarantees around correctness, consistency, and scalability. By combining Go’s performance, PostgreSQL’s reliability, Kafka’s event streaming, and RiverQueue’s background job processing, the system is built to withstand real-world challenges like retries, race conditions, and high transaction volumes.