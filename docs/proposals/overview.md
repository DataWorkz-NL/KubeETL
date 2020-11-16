# Overview

KubeETL aims to simplify and enhance the entire lifecycle of data processing on Kubernetes, similar to what Kubeflow has done for ML.

## Integrating definition and usage of components

The current lifecycle of containers is messy.
Containers are defined using the Dockerfile DSL, builds and pushes are done in CI.
The container is pushed to a certain tag in a certain registry, which need to be configured as magic strings when used in worfklow orchestration.
This process requires a lot of manual effort and relies heavily on manually connecting various stages.

By offering an SDK that integrates the definition (container build instructions, image repository, image tag) and usage (task in workflow orchestration pipeline),
KubeETL streamlines the aforementioned process.

## Proposed functionality

* Workflow orchestration
* Data lineage tracking
* Container lifecycle integration

