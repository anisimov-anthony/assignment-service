#!/bin/bash

export BASE_URL="${BASE_URL:-http://localhost:8080}"

SCENARIO="${1:-smoke}"

case "$SCENARIO" in
smoke) SCRIPT="src/scenarios/01_smoke.js" ;;
average) SCRIPT="src/scenarios/02_average_load.js" ;;
stress) SCRIPT="src/scenarios/03_stress.js" ;;
spike) SCRIPT="src/scenarios/04_spike.js" ;;
soak) SCRIPT="src/scenarios/05_soak.js" ;;
chaos) SCRIPT="src/scenarios/06_chaos_reassign.js" ;;
realistic) SCRIPT="src/scenarios/07_realistic_mix.js" ;;
*) echo "Usage: $0 [smoke|average|stress|spike|soak|chaos|realistic]" && exit 1 ;;
esac

echo "Running scenario: $SCENARIO"
echo "BASE_URL = $BASE_URL"
echo "----------------------------------------"

k6 run "$SCRIPT"
