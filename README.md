<div align="center">

# RabbitMQ Fan-out Pipeline

**An event-driven messaging experiment built with Go, RabbitMQ, and SQLite**

<p>
  <img src="https://img.shields.io/badge/Go-1.26-00ADD8?logo=go&amp;logoColor=white" alt="Go 1.26">
  <img src="https://img.shields.io/badge/RabbitMQ-Broker-FF6600?logo=rabbitmq&amp;logoColor=white" alt="RabbitMQ">
  <img src="https://img.shields.io/badge/SQLite-Embedded-003B57?logo=sqlite&amp;logoColor=white" alt="SQLite">
</p>

</div>

## About the project

This project is a hands-on exploration of event-driven architecture with Go and RabbitMQ. A producer generates synthetic barcode events and publishes them to a fan-out exchange. Two independent consumers receive every event, identify their processing path, and forward the enriched message to a writer process that persists it in SQLite.

The application demonstrates publish/subscribe messaging, process decoupling, message enrichment, queue routing, and centralized persistence through four small command-line programs.

## Architecture

```text
                              ┌──────────────────┐
                         ┌───►│  barcode_work1   │───► Consumer 1 ───┐
                         │    └──────────────────┘                    │
Producer ──► barcode_logs│                                            ▼
            fanout       │                                      writer_queue
            exchange     │                                            │
                         │    ┌──────────────────┐                    │
                         └───►│  barcode_work2   │───► Consumer 2 ───┘
                              └──────────────────┘                    │
                                                                      ▼
                                                               Writer process
                                                                      │
                                          ┌───────────────────────────┴───────────────────────────┐
                                          ▼                                                       ▼
                                 barcodes_work1 table                                    barcodes_work2 table
```

Because `barcode_logs` is a fan-out exchange, each published event is copied to both consumer queues. The consumers add their respective `source` value before sending the message to `writer_queue`. The writer uses that value to select the destination table.

## Message flow

1. The producer generates 10,000 synthetic barcode events.
2. Each event is serialized as JSON and published to `barcode_logs`.
3. RabbitMQ broadcasts every event to `barcode_work1` and `barcode_work2`.
4. Each consumer deserializes the event and adds its processing source.
5. The enriched event is published to `writer_queue`.
6. The writer stores it in the SQLite table associated with that consumer.

Example payload produced at the start of the pipeline:

```json
{
  "source": "",
  "barcode": "12345678",
  "data_envio": "2026-05-04T12:30:00-03:00",
  "code": "GT87654321"
}
```

After processing, `source` is set to either `consumer1` or `consumer2`.

## Key concepts demonstrated

- Fan-out exchanges and publish/subscribe messaging;
- independent consumer processes;
- JSON serialization and shared message contracts;
- message enrichment between pipeline stages;
- queue-to-queue forwarding through the default exchange;
- centralized database writes through a dedicated writer;
- dynamic routing to separate SQLite tables;
- pure-Go SQLite integration without a CGO dependency.

## Technology stack

| Area | Technology |
| --- | --- |
| Language | Go 1.26 |
| Message broker | RabbitMQ / AMQP 0-9-1 |
| AMQP client | `rabbitmq/amqp091-go` |
| Persistence | SQLite |
| SQLite driver | `modernc.org/sqlite` |
| Message format | JSON |

## Project structure

```text
cmd/
├── producer/             # Generates and publishes barcode events
├── consumer1/            # First fan-out subscriber
├── consumer2/            # Second fan-out subscriber
└── writer/               # Persists enriched events in SQLite
internal/
├── models/               # Shared JSON message contract
└── sqlite/               # Local SQLite database
```

## Requirements

- Go 1.26.2 or later;
- RabbitMQ available at `localhost:5672`;
- the default local RabbitMQ credentials (`guest` / `guest`), as currently configured in the source code.

## Running RabbitMQ with Docker

If RabbitMQ is not installed locally, start an instance with the management interface enabled:

```bash
docker run --name rabbitmq-lab \
  -p 5672:5672 \
  -p 15672:15672 \
  rabbitmq:management
```

The AMQP endpoint will be available at `localhost:5672`, and the management interface at `http://localhost:15672`.

> The default credentials are suitable only for local development. Use dedicated users and secrets outside source code in shared or production environments.

## Running the pipeline

Download the dependencies:

```bash
go mod download
```

Open four terminals in the project root and start the components in this order.

### 1. Writer

```bash
go run ./cmd/writer
```

The writer creates the required SQLite tables automatically and waits for enriched events.

### 2. Consumer 1

```bash
go run ./cmd/consumer1
```

### 3. Consumer 2

```bash
go run ./cmd/consumer2
```

### 4. Producer

```bash
go run ./cmd/producer
```

The producer publishes 10,000 events and then exits. Since the exchange broadcasts each event to both consumers, a complete run can generate up to 10,000 rows in each destination table.

## Inspecting the results

The writer stores data in `internal/sqlite/barcodes.db`. Use the SQLite CLI or a database browser to inspect it:

```sql
SELECT COUNT(*) FROM barcodes_work1;
SELECT COUNT(*) FROM barcodes_work2;

SELECT *
FROM barcodes_work1
ORDER BY id DESC
LIMIT 10;
```

The database file is runtime data. For a public repository, it is preferable to exclude the populated file from version control and let the writer create a clean database locally.

## Building

Build all four executables:

```bash
mkdir bin
go build -o bin/producer ./cmd/producer
go build -o bin/consumer1 ./cmd/consumer1
go build -o bin/consumer2 ./cmd/consumer2
go build -o bin/writer ./cmd/writer
```

On Windows, add the `.exe` extension to the output filenames if desired.

## Validation

Format and validate the codebase with:

```bash
go fmt ./...
go vet ./...
go test ./...
```

The project currently has no automated test files. The `go test` command still compiles and validates all packages.

## Current delivery semantics

This repository is an educational prototype rather than a production-ready messaging service. Its current behavior includes:

- automatic acknowledgements, meaning a consumer acknowledges messages before persistence is confirmed;
- non-durable consumer queues;
- non-persistent published messages;
- no dead-letter queue or retry policy for failed messages;
- ignored errors in the producer and consumer processes;
- hard-coded local RabbitMQ connection settings;
- a fixed batch of 10,000 generated events.

These trade-offs keep the example compact while making the next architectural improvements explicit.

## Potential improvements

- Add environment-based configuration for broker URLs and batch size;
- validate every AMQP and JSON operation;
- adopt manual acknowledgements after successful downstream processing;
- make queues and messages durable where delivery guarantees require it;
- introduce retry queues and dead-letter exchanges;
- add graceful shutdown and reconnect handling;
- use publisher confirms and consumer quality-of-service limits;
- move database location and schema initialization into configuration and migrations;
- add unit tests, broker integration tests, and end-to-end pipeline tests;
- package the services with Docker Compose for reproducible local execution.

## Project status

This portfolio project is a functional messaging proof of concept focused on RabbitMQ fan-out behavior and multi-stage event processing. It is intentionally small enough to study while exposing clear paths toward a production-grade distributed system.

---

Developed as an event-driven architecture study using Go, RabbitMQ, and SQLite.
