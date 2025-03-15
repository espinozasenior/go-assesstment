#!/bin/bash

# Script to generate kubeconfig for GKE cluster

set -e

# Get variables from terraform output
CLUSTER_NAME=$(cd terraform && terraform output -raw cluster_name)
CLUSTER_REGION=$(cd terraform && terraform output -raw region 2>/dev/null || grep region terraform/terraform.tfvars | cut -d '=' -f2 | tr -d ' "')
PROJECT_ID=$(grep project_id terraform/terraform.tfvars | cut -d '=' -f2 | tr -d ' "')

# Check if variables are set
if [ -z "$CLUSTER_NAME" ] || [ -z "$PROJECT_ID" ] || [ -z "$CLUSTER_REGION" ]; then
  echo "❌ Error: Missing required variables. Make sure terraform has been applied and terraform.tfvars exists."
  exit 1
fi

# Generate kubeconfig
echo "🔄 Generating kubeconfig for cluster $CLUSTER_NAME in project $PROJECT_ID..."
gcloud container clusters get-credentials "$CLUSTER_NAME" \
  --region="$CLUSTER_REGION" \
  --project="$PROJECT_ID"

# Export kubeconfig to file
KUBECONFIG_PATH="$(pwd)/kubeconfig.yaml"
kubectl config view --minify --flatten > "$KUBECONFIG_PATH"

if [ -f "$KUBECONFIG_PATH" ]; then
  echo "✅ Kubernetes cluster accessed successfully."
  echo "📌 Kubeconfig file saved: $KUBECONFIG_PATH"
else
  echo "❌ Failed to generate kubeconfig file."
  exit 1
fi