# Advanced ELK Setup

This guide adds Elasticsearch, Logstash, and Kibana to the stack for powerful log analytics.

## What’s included

- Elasticsearch (single-node, security disabled for local dev)
- Logstash (HTTP and TCP inputs, JSON pipeline, index per-day)
- Kibana (pre-configured to the ES node)

## Start the stack

```bash
make start_observability_stack
```

This brings up Elasticsearch (9200), Logstash (8080/5000/9600), and Kibana (5601) alongside the existing observability services.

## Emit logs

Use the sample CLI to post JSON logs to Logstash’s HTTP input:

```bash
make run_logger_elk
```

## Verify ingestion

- Elasticsearch: http://localhost:9200 (should return cluster info JSON)
- Kibana: http://localhost:5601 → Discover
  - Create a data view for index pattern `interviewing-golang-*`
  - Query for fields like `app: "interviewing-golang" and env: "dev"`

## Pipeline details

`env/logstash/pipeline/logstash.conf`:
- Inputs: `http` (8080), `tcp` (5000 json_lines)
- Filter: Adds `[@metadata][index] = interviewing-golang`
- Output: Elasticsearch to `interviewing-golang-YYYY.MM.dd` and stdout (rubydebug)

## Production notes

- Enable security and TLS for ES/Kibana/Logstash in real deployments
- Set proper index lifecycle policies for retention
- Consider Beats or OTEL logs exporter for app integration at scale
- Use structured logs with consistent fields (app, env, trace/span IDs)
