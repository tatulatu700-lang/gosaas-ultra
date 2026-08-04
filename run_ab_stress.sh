#!/usr/bin/env bash
set -euo pipefail
TARGET_URL="http://127.0.0.1:8080/api/v1/auth/register"
PAYLOAD_FILE="$HOME/gosaas_ultra/ab_payload.json"
echo '{"email":"ab_stress_node@test.local","password":"secure_entropy_pass"}' > "$PAYLOAD_FILE"
ab -k -n 500 -c 5 -p "$PAYLOAD_FILE" -T "application/json" "$TARGET_URL"
rm -f "$PAYLOAD_FILE"
