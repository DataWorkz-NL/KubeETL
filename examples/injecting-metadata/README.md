# Injecting & Templating Metadata

KubeETL provides the ability to inject metadata into Workflows. This allows you to remove any metadata & configuration from your code.

As a simple example we will use metadata stored in a Dataset to inject configuration into a sample Workflow.

First we create a Kubernetes Secret to store the a password:

```console
kubernetes create secret generic db-secret --from-literal=password='S!B\*d$zDsb='
```

Next we define a Connection that contains the username & password to connect to the MySQL database (see connection.yaml):

```console
kubectl apply -f connection.yaml
```

The Connection can contain inline credentials (`value`) or values from a references (such as secrets and configmaps with `valueFrom`):

```yaml
apiVersion: etl.dataworkz.nl/v1alpha1
kind: Connection
metadata:
  name: mysql-connection
spec:
  type: mysql
  credentials:
    username:
      value: some-username
    password:
      valueFrom:
        secretKeyRef:
          name: db-secret
          key: password
```

Next we create a DataSet that references the Connection and provides additional metadata (see dataset.yaml):


```console
kubectl apply -f dataset.yaml
```

The DataSet gives you the ability to add some extra metadata beyond connection information. In this case we add the database name where to find this specific dataset:

```yaml
apiVersion: etl.dataworkz.nl/v1alpha1
kind: DataSet
metadata:
  name: sessions-dataset
spec:
  connection:
    connectionFrom:
      name: mysql-connection
  metadata:
    host:
      value: mysql
    port:
      value: 3000
    database:
      value: mydatabase
```

To use this information, we create a Workflow that uses the DataSet to inject metadata in some of it's steps (see workflow.yaml):

```console
kubectl apply -f workflow.yaml
```

The workflow defines injectable values, which define how to inject the values (from a Dataset or Connection) and whether to inject the values as Environment Variables or mount them as a file.

The `injectInto` section defines in which steps of the workflow the `injectables` should be injected.

```yaml
apiVersion: etl.dataworkz.nl/v1alpha1
kind: Workflow
metadata:
  name: basic-workflow
spec:
  injectInto:
    - name: bash-template
      injectedValues:
        - injectable-connection
  injectable:
    - name: injectable-connection
      datasetRef:
        name: sessions-dataset
      content: mysql://{{connection.user}}:{{connection.password}}@{{metadata.host}}:{{metadata.port}}/{{metadata.database}}
      envName: MYSQL_URL
  entrypoint: bash-template
  templates:
  - name: bash-template
    inputs:
      parameters:
      - name: args
    container:
      image: busybox:latest
      command: [sh, -c]
      args: ["echo", "$MYSQL_URL"]
```

The templating language allows you to combine information from e.g. a Dataset or a Connection into a single environment variable or file. In this example we utilise this feature to combine the information into a single MySQL connection string.
