# Connections

A Connection captures all the relevant information to connect with a data source or sink. A connection can be referenced from a `DataSource` and should provide all the relevant information to the Task that uses the `DataSource`. A proposal for a `Connection` kind could be the following:

```yml
apiversion: t.b.d.
kind: Connection
metadata:
  name: MySQLConnection
type: MySQL
spec:
  healthCheck: Interval/None/WhenUsed
  credentials:
    url:
      value:
    username:
      value:
    password:
      valueFrom:
        secretKeyRef:
          # ...
    host:
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
type: MySQL
metadata:
  name: MySQLConnection
spec:
  credentials:
    fromSecret: ...
status:
```

Opaque secret:

```yml
apiversion: t.b.d.
kind: Connection
type: Opaque
metadata:
  name: MyGenericConnection
spec:
  credentials:
    - name: MY_ENV_VAR
      value: MY_ENV_VALUE
    - mountPath: /tmp/my_secret_file.ext
      value: |
        supersecretlongtext
status:
```

The `type` field can be used for a dynamic determination of what source is being connected to.
The behavior of this field should mirror that of the `type` field in `Secrets`.
This can be used in a `Task` to determine e.g. what image should be used and what environment variables should be available.
For well-known types the supplied credentials can be validated using a predefined schema.
Mounting connections in a standardized manner - either as predefined environment variables or files at a predefined location - promotes consistent conventions across the pipeline.

For unknown connection types, the `Opaque`-type should be used, like it is in secrets.
Healthchecks are disabled for `Opaque`.
The credentials for `Opaque` are specified in a similar way to environment variables and volumemounts in a `Pod`.
Entries with the `name` field are provided as environment variables, entries with `mountPath` are mounted as volumemounts.

The `Connection` contains all relevant information to achieve a `Connection`.
It is assumed the `Task` utilising the `Connection` is able to use that information to achieve a connection.
The `status` field can be used to track usage of the `Connection` and even to achieve locking functionality if that is desired at some point.

By occasionally (and optionally) performing health checks for a `Connection` we can proactively prevent or influence execution of pipelines to prevent failures due to one or more instances of a faulty `Connection`.
The timing of healthchecks should be configurable to an interval, when it is needed or disabled entirely.

Credentials are key-value pairs that can be supplied in various ways:

- Reference to existing `ConfigMap`/`Secret`
- Inline keys
  - Value from `ConfigMap`/`Secret`
  - Inline values

## Security

Usage of `Connections` can be limited using Kubernetes RBAC. KubeETL must respect this.
Additionally KubeETL should provide a way to securely store the credentials for a connection,
leveraging the Kubernetes `Secrets` or alternatively supporting referencing secrets from a secrets management system such as HashiCorp Vault.
