<!-- markdownlint-disable MD010-->

# Workflows

Our current (draft) Workflow spec extends the Argo Workflow spec with a `connections` property.
To view the Go types associated with this document, see `workflow_types.go`

An example Workflow and the subsequently generated objects are available in `workflow.yaml`

```yaml
apiVersion: etl.dataworkz.nl/v1alpha1
kind: Workflow
metadata:
  generateName: connections-
spec:
  entrypoint: use-connection
  connections:
    - name: mysql-prod
      alias: mysql
      injectable:
        - name: connectionstring
          env: MYSQL_CONNECTION_STRING
          value: "{{.user}}@{{.host}}:{{.port}}/{{.database}}"
        - name: connectionfile
          file: "myconnection.txt"
          content: |
            mysql://{{user}}:{{password}}@{{host}}:{{port}}/{{database}}
  templates:
    - name: inject-by-name
      inject:
        # only injects the `connectionstring` injectable
        - connection: mysql
          name: connectionstring
      container:
        image: alpine:3.7
        command: [sh, -c]
        args:
          - echo "value from env: $MYSQL_CONNECTION_STRING"

    - name: injection-all
      inject:
        # injections without name specified will inject all injectables
        - connection: mysql
      container:
        image: alpine:3.7
        command: [sh, -c]
        args:
          - |
            echo "value from env: $MYSQL_CONNECTION_STRING"
            echo "secret from file: `cat /connections/mountpath/myconnection.txt`"
```

## Connection Injection

```go

// Contains a reference to a `Connection` and the desired injections
// If no InjectDefinitions are specified, credentials will be injected as environment variables
//+kubebuilder:object:generate:=true
type ConnectionInjection struct {
	// Name of the `Connection` that is being injected here
	//+required
	ConnectionName string `json:"name,"`

	// Optional alias for consuming templates
	//+optional
	Alias string `json:"alias,omitempty"`

	// If true, all the InjectDefinitions will be applied to all ContainerTemplates in this workflow.
	// If false, consuming templates must specifically request this ConnectionInjection
	//+optional
	Global bool `json:"global,omitempty"`

	// A list of injections
	//+optional
 InjectDefinitions []InjectDefinition `json:"injectable,"`
}

// InjectDefinition specifies how a connection will be injected
type InjectDefinition struct {
	// Identifier for this injection in case of selective injections
	//+optional
	Name string `json:"name,omitempty"`

	// Name of the injected environment variable
	//+optional
	Key string `json:"key,omitempty"`

	// Path where value will be mounted as a file
	//+optional
	Path string `json:"path,omitempty"`

	// Go template that will be rendered using the connection fields as data
	// Example: mysql://{{.user}}:{{.password}}@{{.host}}:{{.port}}/{{.database}}
	//+required
	Value string `json:"value,"`
}
```

At the top level of a `Workflow`, a list of connections and methods of injection are specified.
The `name` property refers to an existing `Connection`.
The `alias` field can be used by templates in this workflow to refer to this connection.

### Inject Definitions

_TODO: Not sure about the name. "Injectable" could also work_

Inject Definitions specify how the values of a connection will be injected into a target container.
Connections can be injected as environment variables and mounted files.
Specifying `env` or `file` determines the method of injection.
`value` contains a Go-template that is rendered using the `Connection` fields.
This allows more elaborate/tailored injections, such as constructing a connectionstring.

```yaml
apiVersion: etl.dataworkz.nl/v1alpha1
kind: Workflow
metadata:
  generateName: connections-
spec:
  connections:
    - name: mysql-prod
      alias: mysql
      injectable:
        - name: connectionstring
          env: MYSQL_CONNECTION_STRING
          value: "{{.user}}@{{.host}}:{{.port}}/{{.database}}"
        - name: connectionfile
          file: "myconnection.txt"
          content: |
            mysql://{{user}}:{{password}}@{{host}}:{{port}}/{{database}}
```

### Template Injection

```go
// Contains the required keys to select an InjectDefinition
type InjectDefinitionRef struct {
       // The connection name (or alias if defined)
	//+required
	ConnectionKey string `json:"connection"`

	//+optional
	Name string `json:"name,omitempty"`
}

```

```yaml
apiVersion: etl.dataworkz.nl/v1alpha1
kind: Workflow
metadata:
  generateName: connections-
spec:
  templates:
    - name: inject-by-name
      inject:
        # only injects the `connectionstring` injectable
        - connection: mysql
          name: connectionstring
      container:
        image: alpine:3.7
        command: [sh, -c]
        args:
          - echo "value from env: $MYSQL_CONNECTION_STRING"

    - name: injection-all
      inject:
        # injections without name specified will inject all injectables
        - connection: mysql
      container:
        image: alpine:3.7
        command: [sh, -c]
        args:
          - |
            echo "value from env: $MYSQL_CONNECTION_STRING"
            echo "secret from file: `cat /connections/mountpath/myconnection.txt`"
```

Injections are referenced by templates using the `inject` field.
The only required field is `connection`, which refers to a `Connection` name or alias as specified in `connections`.
Omitting the `name` field results in all inject definitions being injected in the container.
Specifying the `name` field results only selects a specific inject definition.

### Injection Secrets

Since we are injecting processed values, simply mounting the `Connection`-fields as environment variables does not suffice.
There are two main ways to provide templating functionality and injecting the results as environment variables.

- Process injection ([Good writeup](https://banzaicloud.com/blog/inject-secrets-into-pods-vault-revisited/))
- Secrets as carriers

Process Injection happens at container runtime by injecting an executable that loads environment variables before launching the original executable.
This method avoids unnecessary duplication of secret data.
The [implementation seems quite managable](https://github.com/banzaicloud/bank-vaults/blob/master/cmd/vault-env/main.go).
Still, it feels like we're "hijacking" the process in some way. It's debatable whether this is a bad thing, but it's something to consider.

Using `Secrets` as carriers for the rendered template values is another approach.
Before a to-be injected container runs, an init container (or initial `Workflow` step) runs.
The `Connection`-fields are mounted to the initcontainer, which renders the template with those values.
The initcontainer then writes these to a `Secret` which is used by the targetcontainer to obtain the rendered value.

_TODO: investigate secret/workflow lifetime_

The injection definitions as specified above would result in the following `Secret` being created:

```yaml
apiVersion: core/v1
kind: Secret
metadata:
  name: connections-9tfdw-mysql # format: ${workflow_name}-${connection name/alias}
data:
  connectionstring: | # key: inject def name
    mysql://admin:p4ssw0rd@foo.bar:3306/baz
  connectionfile: |
    admin:p4ssw0rd@foo.bar:3306/baz
```

Where to populate the carrier `Secret` is an open question.
We could use init containers but this would result in a lot of redundant work, since the same carrier `Secret` is used in the entire `Workflow`.
Another method is inserting a workflow step at the start of the `Workflow` which initializes the `Secret`

Regardless of when this happens, the container definition would look something like the example below.
Note that this only takes care of 1 connection injection.

```yaml
container:
  image: kubeetlinjector
  args:
    - --secret-name={{workflow.parameters.injection-secret-mysql}}
    - --inject
    - connectionstring=mysql://{{user}}:{{password}}@{{host}}:{{port}}/{{database}}
    - --inject
    - connectionfile=mysql://{{user}}:{{password}}@{{host}}:{{port}}/{{database}}
  env:
    - name: inject__user
      valueFrom:
        secretKeyRef:
          name: mysql-cred-secret
          key: user
    - name: inject__password
      valueFrom:
        secretKeyRef:
          name: mysql-cred-secret
          key: password
    - name: inject__host
      valueFrom:
        secretKeyRef:
          name: mysql-cred-secret
          key: host
    - name: inject__port
      valueFrom:
        secretKeyRef:
          name: mysql-cred-secret
          key: port
    - name: inject__database
      valueFrom:
        secretKeyRef:
          name: mysql-cred-secret
          key: database
```
