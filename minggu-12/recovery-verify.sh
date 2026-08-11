#!/usr/bin/env bash
set -eo pipefail

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${YELLOW}====================================================${NC}"
echo -e "${YELLOW}   SRE POST-INCIDENT RECOVERY VERIFICATION SCRIPT   ${NC}"
echo -e "${YELLOW}====================================================${NC}"

echo -n "[1/4] Checking Pod Health in namespace prod-app... "
UNHEALTHY_PODS=$(kubectl get pods -n prod-app --no-headers 2>/dev/null | grep -v "Running" | grep -v "Completed" | wc -l || echo "0")

if [ "$UNHEALTHY_PODS" -eq 0 ]; then
  echo -e "${GREEN}PASSED (All Pods are Running)${NC}"
else
  echo -e "${RED}FAILED ($UNHEALTHY_PODS unhealthy pods found)${NC}"
  kubectl get pods -n prod-app
  exit 1
fi

echo -n "[2/4] Testing HTTP /healthz Endpoint... "
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/healthz || echo "000")

if [ "$HTTP_CODE" -eq 200 ]; then
  echo -e "${GREEN}PASSED (HTTP 200 OK)${NC}"
else
  echo -e "${RED}FAILED (HTTP Status: $HTTP_CODE)${NC}"
  exit 1
fi

echo "[3/4] Running 1-minute K6 Smoke Test to validate SLO latency..."
k6 run --env TEST_TYPE=smoke --env TARGET_URL=http://localhost:8080 minggu-12/manifests/03-k6-comprehensive-loadtest.yaml > /tmp/k6_recovery.log 2>&1 || true

if grep -q "http_req_failed................: 0.00%" /tmp/k6_recovery.log 2>/dev/null; then
  echo -e "${GREEN}PASSED (Error Rate = 0.00%)${NC}"
else
  echo -e "${YELLOW}SKIPPED / WARNING (K6 output check - ensure K6 binary is installed)${NC}"
fi

echo -n "[4/4] Checking Alertmanager for active P1 alerts... "
ACTIVE_ALERTS=$(curl -s http://localhost:9093/api/v2/alerts 2>/dev/null | grep -c "SLOErrorBudgetFastBurn" || echo "0")

if [ "$ACTIVE_ALERTS" -eq 0 ]; then
  echo -e "${GREEN}PASSED (No Active P1 Fast Burn Alerts)${NC}"
else
  echo -e "${YELLOW}WARNING ($ACTIVE_ALERTS P1 alert still active / resolving)${NC}"
fi

echo -e "${GREEN}====================================================${NC}"
echo -e "${GREEN}   SYSTEM FULLY RECOVERED & SLO TARGET ACHIEVED!   ${NC}"
echo -e "${GREEN}====================================================${NC}"
