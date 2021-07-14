#!/usr/bin/env bash

helm repo add jetstack https://charts.jetstack.io && helm repo update

helm install \
  cert-manager jetstack/cert-manager \
  --namespace cert-manager \
  --create-namespace \
  --version v1.4.0 \
  --set installCRDs=true