# Grafana Dashboard Author MCP

Name: grafana.dashboard_author

Problem:
Creating consistent dashboards is slow & error-prone; this MCP encodes patterns and style guardrails.

Inputs:
```json
{
  "service": "generator",
  "panels": [
    {"kind":"latency","histogram":"http_request_duration_seconds","quantiles":[0.5,0.95,0.99]},
    {"kind":"error_ratio","numerator":"http_requests_total{code=~'5..'}","denominator":"http_requests_total"},
    {"kind":"resource","metric":"process_cpu_seconds_total"}
  ],
  "theme": "light",
  "folder": "Services/Generator",
  "uid_prefix": "gen"
}
```

Behavior:
- Generates JSON model (Grafana dashboard schema) with consistent tags, templating for environment, datasource inference.

Output (truncated):
```json
{
  "title":"Generator Service Overview",
  "uid":"gen-overview-abc123",
  "tags":["service:generator","tier:backend"],
  "templating":{"list":[{"name":"env","query":"label_values(up,env)"}]},
  "panels":[ {"type":"timeseries","title":"Latency p50/p95","targets":[...] } ]
}
```

Extensions:
- Add golden signals preset
- Add annotation streams (deploy events)
