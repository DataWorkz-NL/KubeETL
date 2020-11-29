# Task

A `Task` captures all configuration necessary to perform a part of an ETL workflow. E.g. a transformation `Task` would have contains all information required to read from a data source, transform that information, store the relevant lineage & metadata and pass the data to a next stage or store the data in a data sink.

A proposal for the `Task` kind could be the following:

```yml
apiversion: t.b.d.
kind: DataSource
metadata:
  name: TransformTransactionsToFact
spec:
  spec: # PodSpec
    - name: #
    - image: #
    - env:
        - name: MYSQL_CONNECTION_URL
        - value: {{ datasource.connection.url }}
  dataSource:
    fromRef: TransactionTable
  dataSink:
   fromRef: TransactionFact
  schema: # inline Connection to store the schema
    url: s3://schema-storage/finance/transaction_fact
  metadata:
    fromRef: MetadataDB
  # ... additional required configuration
``` 

A templating engine will be used by KubeETL to allow for injection of variables into a container based on `DataSource` definitions.