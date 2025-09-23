#!/usr/bin/env bash
set -euo pipefail

THRESHOLD_PERCENT=${THRESHOLD_PERCENT:-5}
REGRESSION_FOUND=0

# Usage: parse benchstat diff files and mark regression if delta positive > threshold
scan_file() {
  local file="$1"
  [ -f "$file" ] || return 0
  while read -r line; do
    # Expect pattern: name  old time/op  new time/op  delta
    if [[ "$line" =~ ([[:alnum:]_-]+)[[:space:]]+([0-9.]+)(ns|µs|ms|s)?[[:space:]]+([0-9.]+)(ns|µs|ms|s)?[[:space:]]+([+-][0-9.]+)% ]]; then
      delta=${BASH_REMATCH[6]}
      # Remove leading + or - convert to float
      sign=${delta:0:1}
      value=${delta:1}
      # If sign is + and value > THRESHOLD_PERCENT => regression
      greater=$(echo "$value > $THRESHOLD_PERCENT" | bc -l)
      if [[ "$sign" == "+" && "$greater" == "1" ]]; then
        echo "REGRESSION: $line (threshold ${THRESHOLD_PERCENT}%)" >&2
        REGRESSION_FOUND=1
      fi
    fi
  done <"$file"
}

for f in .docs/artifacts/benchdiff/bench_*.diff; do
  scan_file "$f" || true
done

if [[ "$REGRESSION_FOUND" == "1" ]]; then
  echo "Performance regression detected (>${THRESHOLD_PERCENT}%). Failing." >&2
  exit 2
fi

echo "No performance regressions over ${THRESHOLD_PERCENT}% detected."
