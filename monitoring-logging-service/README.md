# Monitoring & Logging Service (Dante Backend)

This component of the Dante GPU Platform is responsible for aggregating logs and metrics from all backend services to provide observability, facilitate debugging, and monitor system health and performance.

Unlike other backend components that are custom-coded services, the Monitoring & Logging Service is primarily a **configuration-based setup** of existing, powerful open-source tools.

---

## Objectives

-   **Centralized Logging:** Collect logs from all microservices in a single, searchable location.
-   **Metrics Collection:** Gather key performance indicators (KPIs) and operational metrics from services (e.g., request rates, error rates, latency, resource utilization).
-   **Visualization & Dashboards:** Provide dashboards to visualize metrics and log data for easy monitoring and analysis.
-   **Alerting:** (Future) Set up alerts based on predefined thresholds for metrics or specific log patterns to notify administrators of potential issues.

---

## Chosen Stack (Initial Proposal)

-   **Metrics Collection:** [Prometheus](https://prometheus.io/)
    -   **Description:** An open-source systems monitoring and alerting toolkit. Services will need to expose a `/metrics` endpoint in a Prometheus-compatible format.
-   **Metrics Visualization:** [Grafana](https://grafana.com/)
    -   **Description:** An open-source platform for monitoring and observability. Grafana will use Prometheus as a data source to build dashboards.
-   **Log Aggregation:** [Loki](https://grafana.com/oss/loki/)
    -   **Description:** A horizontally scalable, highly available, multi-tenant log aggregation system inspired by Prometheus. It's designed to be cost-effective and easy to operate.
-   **Log Shipping:** [Promtail](https://grafana.com/docs/loki/latest/clients/promtail/)
    -   **Description:** An agent that ships the logs from local files or systemd journal to a private Loki instance or Grafana Cloud.

**Alternative Log Aggregation (Considered):**

-   **ELK Stack (Elasticsearch, Logstash, Kibana):** A very powerful but potentially more resource-intensive option.

---

## Integration Strategy

1.  **Metrics:**
    *   Individual Go services (API Gateway, Provider Registry, Scheduler, Storage) will be instrumented using a Go client library for Prometheus (e.g., `prometheus/client_golang`).
    *   Each service will expose an HTTP endpoint (typically `/metrics`) that Prometheus can scrape.
    *   The Python-based `auth-service` can use a Python Prometheus client (e.g., `prometheus-client`).
    *   Prometheus will be configured to discover and scrape these endpoints (potentially via Consul service discovery or static configuration).

2.  **Logging:**
    *   Services will continue to output structured logs (e.g., JSON format with Zap in Go services).
    *   Promtail will be configured to run as an agent (e.g., sidecar in Docker/Kubernetes, or a host agent) to collect these logs.
    *   Promtail will tail log files or read from container stdout/stderr and forward them to Loki, adding appropriate labels (e.g., service name, instance ID).
    *   Loki will store and index these logs.
    *   Grafana will use Loki as a data source for log querying and visualization (e.g., in conjunction with metrics dashboards).

---

## Setup (Conceptual)

This service will be deployed using Docker Compose. The `docker-compose.yml` file in this directory will define services for:

-   `prometheus`
-   `grafana`
-   `loki`
-   `promtail` (Promtail might be configured more globally or per-service, but a base config can reside here)

Configuration files for each tool (`prometheus.yml`, `loki-config.yml`, Grafana provisioning files) will be volume-mounted into the respective containers.

### Running the Stack

```bash
# Navigate to monitoring-logging-service directory
docker compose up -d
```

-   **Prometheus UI:** Typically `http://localhost:9090`
-   **Grafana UI:** Typically `http://localhost:3000`
-   **Loki API:** (Not usually accessed directly by users)

---

## Next Steps

1.  Create placeholder configuration files for Prometheus, Grafana, Loki, and Promtail.
2.  Develop the `docker-compose.yml` to launch the stack.
3.  Instrument existing Go services to expose a `/metrics` endpoint.
4.  Configure Promtail to collect logs from example service(s).
5.  Build basic Grafana dashboards for metrics and logs. 