# KubeETL

## Installation

KubeETL provide quick-start files in the `manifests/` folder. If you want to further customize your configuration we recommend creating your own kustomize overlay.

For the default installation, execute the following commands:

```console
kubectl create namespace kubeetl
kubectl apply -n kubeetl -f https://raw.githubusercontent.com/DataWorkz-NL/KubeETL/manifests/quick-start.yaml
```
