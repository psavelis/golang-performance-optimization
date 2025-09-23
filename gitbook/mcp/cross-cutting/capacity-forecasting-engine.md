# Capacity Forecasting Engine MCP

Name: cross.capacity_forecasting_engine

Problem:
Reactive scaling decisions waste budget or cause brownouts.

Inputs:
```json
{
  "metric": "rate(process_cpu_seconds_total[5m])",
  "lookback_hours": 72,
  "forecast_hours": 24,
  "method": "holtwinters"
}
```

Algorithm:
1. Retrieve time-series
2. Fit selected model (fallback to linear)
3. Forecast + confidence interval
4. Compare forecast vs threshold (e.g., 75% CPU)

Output:
```json
{
  "current_value": 0.62,
  "forecast_peak": 0.78,
  "ci90": [0.71,0.82],
  "threshold_risk":"moderate",
  "recommendations":["Plan +1 replica before peak window"]
}
```

Extensions:
- Add capacity per shard aggregation
