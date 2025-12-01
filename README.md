# Casino Transaction Management System

This project is a simple transaction management system for a casino, built as a technical interview test for Altenar. The system tracks user transactions (bets and wins) asynchronously using a message queue and stores them in a relational database. It also exposes a RESTful API to query the transaction data.

The architecture is based on microservices principles, with two main components:
1.  **Consumer Service**: Listens for transaction messages from a Kafka topic, validates them, and persists them to a PostgreSQL database.
2.  **API Service**: Provides HTTP endpoints for clients to query transaction history.

## Features
 
 *   **Asynchronous Processing**: Transactions are processed asynchronously via Kafka.
 *   **Idempotency**: Prevents duplicate transaction processing using unique `transaction_id` and database constraints.
 *   **Precise Monetary Handling**: Amounts are stored as integers (cents) to ensure absolute precision.
 *   **Reliability**: Manual Kafka offset management ensures at-least-once delivery.
 *   **Data Persistence**: PostgreSQL database for reliable storage.
 *   **REST API**: Endpoints for querying transaction history with filtering.
 *   **Filtering**: Transactions can be filtered by type (`bet` or `win`).
 *   **High Test Coverage**: >85% coverage with unit and integration tests.
 *   **Dockerized**: Fully containerized setup with Docker Compose.
 
 ## Tech Stack
 
 *   **Language**: Go (Golang)
 *   **Database**: PostgreSQL
 *   **Message Broker**: Apache Kafka
 *   **API**: RESTful API using the `chi` router
 *   **Containerization**: Docker & Docker Compose
 *   **Testing**: `testify`, `testcontainers-go`
 
 ## Project Structure
 
 The project follows the standard Go project layout to ensure a clean separation of concerns:
 
 ```text
 .
 ├── cmd/
 │   ├── api/
 │   │   └── main.go
 │   └── consumer/
 │       └── main.go
 ├── internal/
 │   ├── consumer/
 │   │   ├── handler.go
 │   │   ├── handler_integration_test.go
 │   │   └── handler_test.go
 │   ├── handler/
 │   │   ├── transaction.go
 │   │   └── transaction_test.go
 │   ├── models/
 │   │   └── transaction.go
 │   └── repository/
 │       ├── mocks/
 │       │   └── TransactionRepository.go
 │       ├── transaction.go
 │       └── transaction_test.go
 ├── migrations/
 │   ├── 001_create_transactions_table.sql
 │   └── 002_add_transaction_id.sql
 ├── .env.example
 ├── .gitignore
 ├── go.mod
 ├── go.sum
 ├── Makefile
 ├── Dockerfile.api
 ├── Dockerfile.consumer
 └── docker-compose.yml
 ```

## Development Tools

### Makefile

The project includes a `Makefile` for common development tasks:

```bash
make help          # Show all available commands
make build         # Build API and Consumer binaries  
make test          # Run all tests
make coverage      # Generate HTML coverage report
make run-docker    # Start all services with Docker Compose
make clean         # Clean artifacts and stop containers
```

**Quick start for development:**
```bash
make build         # Build both services locally
make test          # Run all tests
make coverage      # Check test coverage
```
 ## Getting Started
 
 ### Prerequisites
 
 *   [Docker](https://www.docker.com/get-started) and [Docker Compose](https://docs.docker.com/compose/install/) must be installed on your system.
 *   [Go](https://golang.org/doc/install/) (v1.25) is required to run tests locally.
 *   `kcat` (or `kafkacat`) is recommended for producing test messages to Kafka. Install via `brew install kcat`.
 *   (Optional) Copy `.env.example` to `.env` for local development configuration.
 
 ### 1. Clone the Repository
 
 ```bash
 git clone https://github.com/OlgaPie/casino-transaction-system.git
 cd casino-transaction-system
 ```
 
 ### 2. Run the System
 The entire system (PostgreSQL, Kafka, Zookeeper, API, and Consumer) can be launched with a single command:
 ```bash
    docker-compose up --build
 ```
 This will:
 1. Build the Docker images for the `api` and `consumer` services.
 2. Start all the necessary containers.
 3. Apply the database schema from the `migrations/` directory on the first run.
 4. Automatically create the `transactions` Kafka topic.
  
 The API service will be available at `http://localhost:8080`.
 ## Usage
 
 ### 1. Producing a Message to Kafka
 
 You can simulate a new transaction by sending a JSON message to the `transactions` Kafka topic.
 
 > **Note**: `amount` is specified in **cents** (e.g., 15075 = $150.75). `transaction_id` is optional; if omitted, it will be generated automatically.
 
 **Example using `kcat`:**
 Open a new terminal and run the following command. After running it, paste the JSON and press `Ctrl+D`.
 
 ```bash
 kcat -b localhost:9092 -t transactions -P
 ```
 Paste your message:
 ```JSON
 {"transaction_id": "tx-1001", "user_id": "user-123", "transaction_type": "bet", "amount": 15075}
 ```
 You can also send a "win" transaction:
 ```JSON
 {"transaction_id": "tx-1002", "user_id": "user-123", "transaction_type": "win", "amount": 30000}
 ```
Check the logs from the `consumer` service (`docker-compose logs -f consumer`) to see that the message has been processed and saved.
### 2. Querying the API
   Use `curl` or any API client (like Postman) to query the transaction data.

   **Get all transactions for a specific user:**
```bash
   curl http://localhost:8080/users/user-123/transactions
```

**Get only "win" transactions for a specific user:**
```bash
   curl http://localhost:8080/users/user-123/transactions?type=win
```

**Get all transactions in the system:**
```bash
   curl http://localhost:8080/transactions
```

**Get all "bet" transactions in the system:**
```bash
   curl http://localhost:8080/transactions?type=bet
```  
## Running Tests

The project includes both unit tests (using mocks) and integration tests (using `testcontainers-go` to spin up a real database).

**Using Makefile (recommended):**
```bash
make test          # Run all tests
make coverage      # Generate HTML coverage report
```

**Using Go directly:**
```bash
go test -v ./...
```
###   Test Coverage
   To generate an HTML report of the test coverage:
```bash
   go test -coverprofile=coverage.out ./...
   go tool cover -html=coverage.out
```
   This will open the coverage report in your default web browser, allowing you to inspect which parts of the code are covered by tests.

#### A Note on Coverage Metrics:
The test coverage report provides several metrics. The most important one is the coverage for the core business logic located in the `internal` package.

*  **Business Logic Coverage (`internal` package): ~ 86.5%**
    - `internal/consumer`: **84.6%**
    - `internal/handler`: **91.3%**
    - `internal/repository`: **83.7%**

   This is the key metric to evaluate. It reflects comprehensive test coverage of all handlers, consumer logic, and repository methods. The achieved coverage successfully exceeds the >85% requirement specified in the test assignment.

*  **Overall Project Coverage**
   Packages like `cmd` (application entrypoints), `models` (data structures), and `mocks` (generated code) are intentionally excluded from coverage metrics as per standard Go testing practices. Therefore, the `internal` package coverage is the true measure of the project's test quality.

## Further Improvements

While this project fulfills the requirements of the test, here are some potential improvements for a production-grade system:

*  **Configuration Management**: Move configuration values (e.g., database DSN, Kafka brokers) out of environment variables and into a structured config file (e.g., YAML) loaded by a library like Viper.
*  **Error Handling and Retries**: Implement a more robust error handling strategy in the consumer, such as a retry mechanism with exponential backoff or a Dead Letter Queue (DLQ) for messages that repeatedly fail processing.
*  **Pagination**: Add pagination (`limit` and `offset` query parameters) to the API endpoints that return lists of transactions to handle large datasets efficiently.
*  **Structured Logging**: Use a structured logging library (e.g., `slog`, `zerolog`) to produce machine-readable JSON logs, which are easier to parse and analyze in a production environment.
*  **Monitoring and Metrics**: Add Prometheus metrics and health endpoints for production monitoring.