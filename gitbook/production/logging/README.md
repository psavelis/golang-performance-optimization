# Centralized Logging (Fluentd + Loki)

This section covers a lightweight, production-ready logging path using Fluentd for ingestion and Grafana Loki for storage and query.

- Fluentd: Receives logs via HTTP (9880) or Forward (24224) and forwards to Loki
- Loki: Stores logs with labels for fast query via Grafana Explore
- Grafana: Loki datasource provisioned, query with LogQL
