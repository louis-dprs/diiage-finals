#!/usr/bin/env bash
set -e

echo "Installing ArgoCD via Helm..."

helm repo add argo https://argoproj.github.io/argo-helm
helm repo update

kubectl create namespace argocd --dry-run=client -o yaml | kubectl apply -f -

helm install argocd argo/argo-cd \
  --namespace argocd \
  --set server.service.type=NodePort \
  --set server.insecure=true \
  --set configs.secret.argocdServerAdminPassword='$2a$10$6Mpy/pHxJjRuBQqNkEHeJelFb0ZDLmksr.klBkgbgctU3yXuPpBqK' \
  --set configs.secret.argocdServerAdminPasswordMtime="$(date -u +%Y-%m-%dT%H:%M:%SZ)"


echo "ArgoCD installation completed."

