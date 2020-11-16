# Integrating definition and usage of components

The current lifecycle of containers is messy.
Containers are defined using the Dockerfile DSL, builds and pushes are done in CI.
The container is pushed to a certain tag in a certain registry, which need to be configured as magic strings when used in worfklow orchestration.
This process requires a lot of manual effort and relies heavily on manually connecting various stages.

By offering an SDK that integrates the definition (container build instructions, image repository, image tag) and usage (task in workflow orchestration pipeline),
KubeETL streamlines the aforementioned process.
The language of choice for this SDK would be go, rather than Python as is more commonly seen for DSL offerings for similar concepts.
The reasoning here is that Go has much better integration with the container ecosystem, as many components are written in Go.
Furthermore, Go-modules offer interesting options for integrating KubeETL container builds in components that are located in separate repositories.
Containers could be defined in a small Go-module in a project and then imported in a KubeETL orchestration project.

The SDK would offer the following functionality:

* Build instructions
* Image metadata configuration (label, name, maintainer, etc)
* Optional build/push to registry
* Operational requirements

## Container build instructions

Written in Go using (a wrapper around?) [Moby Buildkit](https://github.com/moby/buildkit).
The examples from the buildkit repo are similar to what KubeETL could be, but the syntax and constructs are not immediately obvious (or documented). We could consider wrapping Buildkit.

This also allows for standardization of container build "parts", like with the project [gobuild](github.com/tonistiigi/llb-gobuild), which standardizes container build process for a Go application. The output is Buildkit LLB State which can either be built with Buildkit or used downstream.

We could also support plain Dockerfiles by just passing them through to Buildkit.
What we really need is a Go module that lists information about containers, including the build and or push process could be optional.
This would also help with adoption.

We could use Argo for the CI/CD process. [Blog article describing Argo for CI/CD](https://medium.com/axons/ci-cd-with-argo-on-kubernetes-28c1a99616a9).

## Image metadata

The Go-module for an image would contain the following metadata:

* Name
* Repository
* Tag
* Version
* Label

## Operational metadata

Another "category" of metadata is "operational metadata".
A container typically needs credentials and other inputs from its environment, either by passing command line arguments or setting environment variables.
We could streamline this aspect by declaring operational requirements.
This keeps the requirements for running an image explicit and defined at the source.
Examples include:

* Credentials/secrets, mounted as environment variables
* Input (stdin, file input)
* Output (stdout, file output)

