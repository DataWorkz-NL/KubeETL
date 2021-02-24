# Workflows

Our current Workflow spec extends the Argo Workflow spec with a `connections` property.

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Workflow
metadata:
  generateName: connections-
  annotations:
    # Signals KubeETL to inject stuff into this workflow
    kubeetl.dataworkz.nl/inject: enabled
spec:
  entrypoint: use-connection
  connections:
    - name: mysql
      # Entries in `inject` are converted into secrets. Secret names are deterministic based on the injection definition.
      # This ensures minimal duplication of sensitive data.
      # One thing to note is that we should keep track of secrets and workflows using them.
      # A very tidy solution would be JITing secret creation before being used by workflows but that's a bit tough at this stage.
      inject:
        # Create secret based on `value` template and mount as environment variable.
        # Secret key is the connection-name suffixed with MD5 hash of the value/content template.
        - key: MYSQL_CONNECTION_STRING
          value: "mysql://{{user}}:{{password}}@{{host}}:{{port}}/{{database}}"
        # Create secret based on `path` and `content` template.
        # Naming is different for file-injections because filenames are based on secret keys.
        # Therefore, even if the `content` matches another injection (both env/file), a new secret must be created if the filename differs.
        # Secret names for mounted connections must be the MD5 of path/content combo.
        - path: "/connections/mountpath/myconnection.txt"
          content: |
            mysql://{{user}}:{{password}}@{{host}}:{{port}}/{{database}}
  #### INJECTED BY KUBEETL ####
  volumes:
    - name: mysql-vol
      secret:
        secretName: mysql-b395d97407bedeb88f7919f8ae760d95
  #### END INJECTED BY KUBEETL ####
  templates:
    - name: print-secret
      container:
        image: alpine:3.7
        command: [sh, -c]
        args: [
            '
            echo "value from env: $MYSQL_CONNECTION_STRING"
            echo "secret from file: `cat /connections/mountpath/myconnection.txt`"
            ',
          ]
        #### INJECTED BY KUBEETL ####
        env:
          - name: MYSQL_CONNECTION_STRING
            valueFrom:
              secretKeyRef:
                name: mysql-cdb539752cf19ae3cde53bb161493cbd
                key: value
        volumeMounts:
          - name: mysql-cdb539752cf19ae3cde53bb161493cbd
            mountPath: "/connections/mountpath"
        #### END INJECTED BY KUBEETL ####
```

## Old initial design

A `Workflow` is a collection of tasks that are executed as part of an ETL workflow. A `Workflow` can reference one or more tasks, and the `Workflow` is responsible for ensuring the `Tasks` are executed in the correct order. The order should be captured by maintaining the dependencies of individual `Tasks` in the `Workflow`. A proposal for the `Workflow` kind could be the following:

```yaml
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
