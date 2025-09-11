# Tutorial: Capacity Forecasting MCP

Objective: Implement `cross.capacity_forecasting_engine` for proactive scaling decisions.

Steps:
1. Inputs: metric, lookback_hours, forecast_hours, method
2. Query time-series (Prometheus range query) covering lookback
3. Resample to uniform step (e.g., 1m) & fill gaps (linear or carry forward)
4. Apply model:
   - holtwinters: if seasonality present
   - fallback linear regression if variance low
5. Extract forecast peak & confidence interval
6. Compare against threshold (e.g., 0.75 CPU) -> classify risk

Go Pseudocode:
```go
series := fetchRange(promURL, expr, lookback)
clean := resample(series, step)
model := fitHoltWinters(clean)
forecast := model.Forecast(forecastH)
peak := max(forecast.Values)
ci := forecast.CI90()
```

Edge Cases:
- Sparse data -> revert to linear trend
- Sudden spike in last 5% points -> optionally dampen

Extensions:
- Multi-series aggregation (per shard) -> sum & headroom
- Export Grafana panel JSON with forecast overlay
