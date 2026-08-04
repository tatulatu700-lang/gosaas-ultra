#!/usr/bin/env bash
set -euo pipefail

echo "=============================================================="
echo "    GOSAAS-ULTRA — AUTOMATED AGENT PRODUCTION DEPLOYMENT      "
echo "=============================================================="

# 1. Check Environmental Prerequisites
for cmd in go python3 systemctl; do
    if ! command -v "$cmd" &> /dev/null; then
        echo "[-] Error: Missing required system dependency: $cmd" >&2
        exit 1
    fi
done

# 2. Synchronize Go Module Infrastructure
echo "[*] Synchronizing backend package footprints..."
go mod tidy

# 3. Native Architecture Hardened Compile Pass
echo "[*] Compiling hyper-optimized static production binary..."
CGO_ENABLED=0 go build -ldflags="-s -w" -o go_saas_core main.go

# 4. Inject Native Systemd Service Configuration
echo "[*] Registering persistent background systemd service..."
sudo tee /etc/systemd/system/gosaas_core.service > /dev/null << SERVICE_EOF
[Unit]
Description=GoSaaS-Ultra Production Engine
After=network.target

[Service]
Type=simple
User=$USER
WorkingDirectory=$(pwd)
ExecStart=$(pwd)/go_saas_core
Restart=always
RestartSec=5
Environment=PORT=8080

[Install]
WantedBy=multi-user.target
SERVICE_EOF

# 5. Kickstart Background Service
echo "[*] Launching system daemon matrices..."
sudo systemctl daemon-reload
sudo systemctl enable --now gosaas_core.service

# 6. Verify Active Interface State
echo "[*] Validating network port allocation..."
sleep 1
if curl -s -I http://127.0.0 2>&1 | grep -q "405"; then
    echo -e "\n\033[1;32m[SUCCESS] GoSaaS-Ultra is live and accepting traffic on port 8080!\033[0m"
else
    echo -e "\n\033[1;31m[-] Deployment warning: Check status with 'sudo systemctl status gosaas_core.service'\033[0m"
fi
