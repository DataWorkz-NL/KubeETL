#!/usr/bin/env bash
set -euf -o pipefail

kind create cluster --wait 200s

kubectl cluster-info --context kind-kind
kubectl config use-context kind-kind

helm repo add jetstack https://charts.jetstack.io && helm repo update

helm install \
  cert-manager jetstack/cert-manager \
  --namespace cert-manager \
  --create-namespace \
  --version v1.4.0 \
  --set installCRDs=true