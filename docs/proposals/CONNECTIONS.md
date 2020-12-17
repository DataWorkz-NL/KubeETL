# Connections

A Connection captures all the relevant information to connect with a data source or sink. A connection can be referenced from a `DataSource` and should provide all the relevant information to the Task that uses the `DataSource`. A proposal for a `Connection` kind could be the following:

```yml
apiversion: t.b.d.
kind: Connection
metadata:
  name: MySQLConnection
spec:
  url: localhost:3306
  protocol: MySQL
  credentials:
    - username:
        value:
    - password:
        valueFrom:
          secretKeyRef:
          # ...
    - host:
        valueFrom:
          configMapKeyRef:
          # ...
    # ...
status:
```

Credentials from Secret:

```yml
apiversion: t.b.d.
kind: Connection
metadata:
  name: MySQLConnection
spec:
  url: localhost:3306
  protocol: MySQL
  credentials:
    fromSecret: ...
status:
  
```

The `protocol` field can be used for a dynamic determination of what source is being connected to.
This can be used in a Task to determine e.g. what image should be used and what environment variables should be available.
For well-known protocols the supplied credentials can be validated using a predefined schema.
Mounting connections as standardized environment variables also promotes consistent conventions across the pipeline.

The `Connection` contains all relevant information to achieve a `Connection`.
It is assumed the `Task` utilising the `Connection` is able to use that information to achieve a connection.
The `status` field can be used to track usage of the `Connection` and even to achieve locking functionality if that is desired at some point.

By occasionally (and optionally) performing connectivity tests for a `Connection` we can proactively prevent or influence execution of pipelines to prevent failures due to one or more instances of a faulty `Connection`.

Credentials are key-value pairs that can be supplied in various ways:

* Reference to `ConfigMap`/`Secret`
* Inline keys
  * Value from `ConfigMap`/`Secret`
  * Inline values

## Security

Usage of `Connections` can be limited using Kubernetes RBAC. KubeETL must respect this. Additionally KubeETL should provide a way to securely store the credentials for a connection, leveraging the Kubernetes `Secrets` or alternatively supporting referencing secrets from a secrets management system such as HashiCorp Vault.
