# Data Sources

A `DataSource` captures all relevant information that is needed to use the `DataSource`. It could capture the following information:

- Connection: The connection used to connect to the `DataSource`.
- Location: location of the `DataSource` at the `Connection`. E.g. a table in a database or a file in a storage system.
- Schema: A schema of the data (e.g. a schema.json). Could also be provided as a link to another `DataSource` providing the schema.
- Other relevant metadata such as data quality, usage, lineage, etc. could all be captured as part of a `DataSource`.

A proposal for a `DataSource` kind could be the following:

```yml
apiversion: t.b.d.
kind: DataSource
metadata:
  name: TransactionTable
spec:
  connection:
    fromRef: MySQLConnection
  location:
    schema: finance
    table: transactions
  schema:
    url: s3://schema-storage/finance/transaction # inline connection reference
status:
  dataQuality:
    # to be determined
  usage:
   # to be determined
```

## Security

The Kubernetes RBAC mechanism should provide a mechanism to secure the access to `DataSource` based on the permissions given to a `ServiceAccount`. KubeETL must respect the RBAC permissions. Additionally the `ServiceAccount` should have access to any reference in the `DataSource`.
