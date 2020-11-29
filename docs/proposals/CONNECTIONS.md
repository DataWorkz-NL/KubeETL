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
   username:
    fromSecret:
     # ...
   password:
    fromSecret:
     # ...
status:
  
```

The `protocol` field can be used for a dynamic determination of what source is being connected to. The `Connection` contains all relevant information to achieve a `Connection`. It is assumed the `Task` utilising the `Connection` is able to use that information to achieve a connection. The `status` field can be used to track usage of the `Connection` and even to achieve locking functionality if that is desired at some point.

# Security

Usage of `Connections` can be limited using Kubernetes RBAC. KubeETL must respect this. Additionally KubeETL should provide a way to securely store the credentials for a connection, leveraging the Kubernetes `Secrets` or alternatively supporting referencing secrets from a secrets management system such as HashiCorp Vault.
