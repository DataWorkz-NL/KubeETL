# Workflows

A `Workflow` is a collection of tasks that are executed as part of an ETL workflow. A `Workflow` can reference one or more tasks, and the `Workflow` is responsible for ensuring the `Tasks` are executed in the correct order. The order should be captured by maintaining the dependencies of individual `Tasks` in the `Workflow`. A proposal for the `Workflow` kind could be the following:

```yml
apiversion: t.b.d.
kind: Workflow
metadata:
  name: TransactionTransformationWorkflow
spec:
  tasks:
    - name: load_data
      fromRef: MySQLDataSourceTask
      input:
        dataSource: FinanceDB
        dataSink: EphemeralTransactions
    - name:
      fromRef: #...
      dependsOn: load_data
      # etc.
```
