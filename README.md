# dante-backend

This repository contains the backend microservices for the Dante GPU platform.

## Services

*   **api-gateway**: The main entry point for all external requests (Go). Handles routing, authentication, rate limiting, etc.
*   **(Upcoming)** auth-service: Manages user authentication and authorization.
*   **(Upcoming)** provider-registry-service: Tracks available GPU providers.
*   **(Upcoming)** job-queue-service: Manages the queue for AI jobs.
*   **(Upcoming)** scheduler-orchestrator-service: Assigns jobs to providers.
*   **(Upcoming)** storage-service: Interface for data/model storage.
*   **(Upcoming)** monitoring-logging-service: Handles system monitoring and logging.
*   **(Upcoming)** billing-payment-service: Manages billing and payments.

See the README within each service directory for more details.