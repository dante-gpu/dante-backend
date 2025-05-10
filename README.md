# Dante GPU Platform - Backend Services

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Welcome to the central backend repository for the **Dante GPU Platform**. This project implements a microservices architecture to manage distributed GPU resources for AI training and other compute-intensive tasks.

---

## Architecture Overview

The backend is designed as a collection of independent services communicating via APIs (REST, gRPC) and message queues (NATS). This approach promotes scalability, resilience, and maintainability.

**Core Components:**

1.  **API Gateway (`api-gateway/`)**: 
    *   **Language:** Go
    *   **Description:** The single entry point for all external client requests. Responsible for request routing, JWT authentication/authorization, rate limiting, CORS handling, service discovery (via Consul), load balancing, and reverse proxying to appropriate backend services.
    *   **Status:** **Active Development** (Core proxying, auth, NATS publishing implemented)

2.  **Authentication Service (`auth-service/`)**:
    *   **Language:** Python (FastAPI)
    *   **Description:** Manages user accounts (providers and requesters), including registration, credential verification (password hashing), and potentially profile management. Provides endpoints for the API Gateway to verify credentials.
    *   **Status:** **Active Development** (User model, schemas, CRUD, registration, login endpoints implemented; DB migrations setup)

3.  **Provider Registry Service (`provider-registry-service/`)**:
    *   **Language:** Go
    *   **Description:** Tracks currently connected GPU providers, their hardware specifications (GPU model, VRAM, drivers), real-time status (idle, busy), location, and utilization metrics. Features robust database error handling, secure credential management, and context-aware logging.
    *   **Status:** **Implemented** (Core functionality complete with PostgreSQL storage, Consul integration, error handling, and secure credential management)

4.  **Job Queue Service (`job-queue-service/`)**:
    *   **Language:** Integrated via NATS
    *   **Description:** Handles the queuing of AI job requests received from users via the API Gateway. NATS JetStream is used for persistence and reliable delivery.
    *   **Status:** **Partially Implemented** (NATS integration exists in API Gateway)

5.  **Scheduler/Orchestrator Service (`scheduler-orchestrator-service/`)**:
    *   **Language:** Go
    *   **Description:** The "brain" of the system. Dequeues jobs from NATS JetStream, queries the `provider-registry-service` to find suitable and available GPUs, assigns tasks via NATS to provider daemons, and tracks job progress through a persistent job store.
    *   **Status:** **Implemented** (Job consumption, provider matching, task dispatch, and database persistence implemented)

6.  **Storage Service (`storage-service/`)**:
    *   **Language:** Go
    *   **Description:** Provides an abstraction layer for storing user data, AI models, datasets, and job results. This service interfaces with an underlying object storage solution, with MinIO as the initial backend. It supports operations like upload, download, delete, list objects, and presigned URL generation.
    *   **Status:** **Implemented** (Core functionality with MinIO backend, Consul integration, and robust configuration management)

7.  **Monitoring & Logging Service (`monitoring-logging-service/`)**:
    *   **Language:** N/A (Configuration-based)
    *   **Description:** Aggregates logs and metrics from all services for monitoring and debugging. Potential stack: Prometheus for metrics, Grafana for visualization, ELK Stack (Elasticsearch, Logstash, Kibana) or Loki/Tempo for logging.
    *   **Status:** **Planned**

8.  **Billing & Payment Service (`billing-payment-service/`)**:
    *   **Language:** TBD (Likely Python or Go)
    *   **Description:** Tracks resource usage (GPU time, storage), generates invoices for requesters, and handles payouts to providers. Requires integration with payment gateways (e.g., Stripe, PayPal).
    *   **Status:** **Planned (Post-MVP)**

---

## Getting Started

### Prerequisites

*   Docker & Docker Compose
*   Go 1.22+ (for Go services)
*   Python 3.10+ (for `auth-service` development)
*   Consul (for service discovery)
*   NATS & NATS JetStream (for message queuing and persistence)
*   PostgreSQL (for database storage)
*   Make (optional, for running common commands)

### Running the System (Conceptual)

While a full `docker-compose.yml` for the entire system is pending, the general steps to run the currently implemented services involve:

1.  **Start Infrastructure:** Launch Consul, NATS, and PostgreSQL containers.
2.  **Build & Run `api-gateway`:** Navigate to `api-gateway/`, build the binary (`go build ...`), and run it. Ensure it can connect to Consul and NATS.
3.  **Setup & Run `auth-service`:** Navigate to `auth-service/`, create/activate a Python virtual environment, install dependencies (`pip install -r requirements.txt`), configure the `.env` file (especially `DATABASE_URL`), run database migrations (`alembic upgrade head`), and start the service (`uvicorn app.main:app ...`).
4.  **Setup & Run `provider-registry-service`:** Navigate to `provider-registry-service/`, build the binary (`go build ./cmd/main.go -o provider-registry`), and run it. The service will register with Consul and connect to PostgreSQL.
5.  **Setup & Run `scheduler-orchestrator-service`:** Navigate to `scheduler-orchestrator-service/`, build the binary (`go build ./cmd/main.go -o scheduler-orchestrator`), and run it. It will connect to NATS, PostgreSQL, and discover the provider registry service via Consul.

Refer to the `README.md` file within each service directory for detailed setup and execution instructions.

---

## Development

*   **Branching:** Please follow standard Gitflow or a similar branching model (e.g., feature branches off `main` or `develop`).
*   **Commits:** Use [Conventional Commits](https://www.conventionalcommits.org/) for clear and automated versioning/changelog generation.
*   **Code Style:** Adhere to standard Go formatting (`gofmt`) and Python formatting (e.g., Black, Ruff).
*   **Dependencies:** Manage dependencies using Go Modules (Go services) and `pip`/`requirements.txt` (`auth-service`).
*   **Error Handling:** Use the custom error packages in each service for consistent error handling patterns.

---

## What's Completed

- **Custom error handling system** with typed errors, proper wrapping, and error checking utilities
- **Context-aware logging** with correlation IDs for request tracing
- **Secure database credentials management** with environment variable and file-based secrets
- **Database operation retry mechanism** for handling transient errors
- **Provider Registry service** with full CRUD operations for GPU providers
- **Scheduler/Orchestrator service** with job consumption, provider matching, and task dispatch
- **Service discovery and registration** via Consul
- **Message passing infrastructure** using NATS and JetStream
- **Storage service** with MinIO backend, supporting uploads, downloads, presigned URLs, and bucket management.

## What Remains To Be Done

- **Provider daemon client** implementation that receives and executes tasks on provider machines
- **End-to-end job execution flow** testing and optimization
- **Metrics collection** for job performance and provider status
- **User interface** for monitoring jobs and managing providers
- **Billing system** based on resource usage
- **Comprehensive test coverage** and integration tests
- **Deployment automation** with Docker Compose or Kubernetes
- **Advanced scheduling algorithms** based on job requirements and provider capabilities

---

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.