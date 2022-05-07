# KubeETL

KubeETL is a Kubernetes based framework for managing datasets and creating data-driven pipelines that interact with those datasets. It aims to simplify tasks that commonly arise when managing a large number of datasets, such as:

- Maintaining metadata about the datasets.
- Checking the quality of the data.
- Connecting to the source that holds the dataset.
- Leveraging datasets created by other teams.
- Finding issues in datasets based on the lineage of that dataset.

Often these are the types of tasks that are pushed to the backlog in favor of connecting more data sources and providing more reporting to downstream consumers of data. However, in our experience we also know that if these tasks are not prioritised, eventually you will experience issues in the reliability of your workflows. Unreliable workflows lead to unreliable data, and that will affect the trust your end users have in your data.

KubeETL is available for your usage under the Apache 2.0 License.

## Installation

KubeETL provide quick-start files in the `manifests/` folder. If you want to further customize your configuration we recommend creating your own kustomize overlay.

For the default installation, execute the following commands:

```console
kubectl create namespace kubeetl
kubectl apply -n kubeetl -f https://raw.githubusercontent.com/DataWorkz-NL/KubeETL/manifests/quick-start.yaml
```

As KubeETL (currently) relies on Argo Workflows to execute Workflows you have to install Argo Workflows as well. See the [Argo Workflows quick start](https://argoproj.github.io/argo-workflows/quick-start/) guide for more details

## Examples

See the [examples](./examples/) directory for examples of how to use KubeETL. Each example provides a README that explains how to follow along.

## Concepts

KubeETL provides adds a set of Custom Resources (CRs) to your Kubernetes cluster to define and interact with datasets. These are:

- Connections (To define how to connect to a source or a sink)
- Datasets (To define a dataset)
- Workflows (To create a pipeline to interact with the Datasets or Connections)

Based on these resources KubeETL can simplify your interaction with the data by:

- Injecting connection details into your workflows and unifying access to sources and sinks.
- Automating data validation and preventing workflows from working with invalid data.
- Collecting valuable metadata about your data.

KubeETL leverages Kubernetes to provide all these mechanism. At it's core, KubeETL is a Kubernetes Operator, that can interact with Workflows running on Kubernetes.

## Features

Currently KubeETL provides the following features:

- Custom Resources for DataSets, Connections & Workflows
- DataSet & Connection metadata validation using Admission Webhooks
- Creating custom workflows to track DataSet health
- Automatically injecting Connection and DataSet information into a Workflow

## Roadmap

Currently we have the following main priorities:

- Integrating KubeETL with a metadata collection framework (such as Openmetadata or Openlineage).
- Decouple KubeETL from Argo Workflows, so it can work with other Workflow schedulers such as Airflow or Prefect.
- Improving our documentation and creating a documentation website.

If you want to contribute to the evolution of KubeETL, see the next section.

## Contributing

We gladly accept contributions to the project. We accept any kinds of improvements:

- Documentation improvements
- Bug reports
- New features
- Suggestions & Use cases

We would also love to hear where you would like to see the project evolve too. Feel free to open an issue on Github to share your ideas.

Make sure to check out our [contributing guide](CONTRIBUTING.md) before making a contribution to the project.
