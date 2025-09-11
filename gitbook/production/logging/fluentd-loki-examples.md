# Fluentd → Loki: Examples

## Start the stack

```bash
make start_observability_stack
```

This includes Fluentd (HTTP: 9880, Forward: 24224) and Loki (3100).

## Emit logs

Use the provided CLI to send sample logs via Fluentd Forward:

```bash
make run_logger_fluentd
```

You can also send JSON via HTTP (root path accepts the payload with tag/record):

```bash
curl -X POST -H 'Content-Type: application/json' \
  -d '{"tag":"interviewing.golang","time":null,"record":{"message":"hello","app":"interviewing-golang","env":"dev"}}' \
  http://localhost:9880/
```

## Query logs in Grafana

- Open Grafana → Explore → Datasource: Loki
- Run a simple query:

```
{job="fluentd"}
```

- Or filter by our labels:

```
{app="interviewing-golang", env="dev"}
```

## Notes

- Fluentd config: `env/fluentd/fluent.conf`
- Loki config: `env/loki/config.yaml`
- Grafana datasource: `env/grafana/provisioning/datasources/loki.yml`
- For production, secure Fluentd inputs and consider structured logs with consistent labels.
