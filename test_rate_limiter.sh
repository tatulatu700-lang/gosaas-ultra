#!/usr/bin/env bash
set -euo pipefail
echo "=============================================================="
echo "    LAUNCHING VOLUMETRIC RATE-LIMITER AUDIT PASS             "
echo "    Target Matrix: http://127.0.0.1:8080/api/v1/auth/register"
echo "    Burst Profile: 15 Rapid Sequential Hits"
echo "=============================================================="
success_count=0; blocked_count=0
for ((i=1; i<=15; i++)); do
    STATUS_CODE=$(curl -s -o /dev/null -w "%{http_code}" -X POST -H "Content-Type: application/json" -d '{"email":"flood_test_'$i'@test.local","password":"secure_entropy_pass"}' "http://127.0.0.1:8080/api/v1/auth/register")
    if [ "$STATUS_CODE" -eq 201 ]; then
        echo "[Request $i] Status: $STATUS_CODE -> ALLOWED"
        success_count=$((success_count + 1))
    elif [ "$STATUS_CODE" -eq 429 ]; then
        echo -e "[Request $i] Status: \033[1;31m$STATUS_CODE\033[0m -> \033[1;31mBLOCKED BY SECURITY CEILING\033[0m"
        blocked_count=$((blocked_count + 1))
    else
        echo "[Request $i] Status: $STATUS_CODE -> UNKNOWN_STATE"
    fi
done
echo "=============================================================="
echo "                   AUDIT SUMMARY DATA                         "
echo "=============================================================="
echo "Total Dispatched Requests: 15"
echo "Allowed Transactions:      $success_count (Token Burst Cap)"
echo "Intercepted Exceptions:    $blocked_count (Rate Deflected)"
echo "=============================================================="
