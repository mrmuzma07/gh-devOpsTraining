#!/usr/bin/env bash
set -eo pipefail

CLUSTER_NAME="ha-cluster"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${YELLOW}====================================================${NC}"
echo -e "${YELLOW}   PROVISIONING 5-NODE k3s HA CLUSTER VIA K3D      ${NC}"
echo -e "${YELLOW}====================================================${NC}"

if ! command -v k3d &> /dev/null; then
  echo -e "${RED}ERROR: k3d CLI belum terinstall!${NC}"
  echo "Silakan install k3d terlebih dahulu: curl -s https://raw.githubusercontent.com/k3d-io/k3d/main/install.sh | bash"
  exit 1
fi

if k3d cluster list 2>/dev/null | grep -q "$CLUSTER_NAME"; then
  echo -e "${YELLOW}Mendeteksi cluster lama '$CLUSTER_NAME'. Menghapus...${NC}"
  k3d cluster delete "$CLUSTER_NAME"
fi

echo -e "${GREEN}[1/3] Creating 3 Server (Master) + 2 Agent (Worker) HA Cluster...${NC}"
k3d cluster create "$CLUSTER_NAME" \
  --servers 3 \
  --agents 2 \
  --port "8080:80@loadbalancer" \
  --port "8443:443@loadbalancer" \
  --k3s-arg "--disable=traefik@server:*" \
  --wait

echo -e "${GREEN}[2/3] Verifying Node Status...${NC}"
kubectl get nodes -o wide

echo -e "${GREEN}[3/3] Checking Control Plane etcd Members...${NC}"
kubectl get endpoints kubernetes -n default

echo -e "${GREEN}====================================================${NC}"
echo -e "${GREEN}   HA CLUSTER SUCCESSFULLY PROVISIONED & READY!     ${NC}"
echo -e "${GREEN}====================================================${NC}"
