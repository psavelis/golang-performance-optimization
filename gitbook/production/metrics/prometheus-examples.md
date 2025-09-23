# Prometheus Examples

1. Start the stack:

  ```bash
  make start_observability_stack
  ```

2. Generate sample data and expose metrics:

  ```bash
  METRICS_ADDR=:2112 METRICS_HOLD_FOR=60s ./target/build/bin/generator-prometheus 100000 .docs/artifacts/test-data/test_prom.json
  ```

3. Load data and expose metrics:

  ```bash
  METRICS_ADDR=:2113 METRICS_HOLD_FOR=60s ./target/build/bin/loader-prometheus 'postgresql://test:test@localhost:5432/test?sslmode=disable' .docs/artifacts/test-data/test_prom.json
  ```

4. Explore metrics:

  Prometheus (http://localhost:9090):

  - events_generated_total
  - generation_duration_seconds
  - interviewing_golang_loader_events_loaded_total
  - interviewing_golang_loader_load_duration_seconds
  - interviewing_golang_loader_transaction_duration_seconds

  Grafana (http://localhost:3000): add a panel with PromQL queries above.
