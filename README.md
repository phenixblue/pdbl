# pdbl

A utility to lookup workloads associated with a Kubernetes Pod Disruption Budget (PDB)

```
A utility to lookup workloads associated with a Kubernetes Pod Disruption Budget (PDB)

Usage:
  pdbl [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  list        List the Pod Disruption Bidget (PDB) resources
  lookup      Lookup the pods assocaited with a target Pod Disruption Bidget (PDB) resource

Flags:
      --context string      The name of the kubeconfig context to use
  -h, --help                help for pdbl
      --kubeconfig string   Kubernetes configuration file
  -n, --namespace string    The Namespace to use when listing Pods
  -t, --toggle              Help message for toggle

Use "pdbl [command] --help" for more information about a command.
```

## List Command

```
List the Pod Disruption Bidget (PDB) resources

Usage:
  pdbl list [flags]

Flags:
  -h, --help   help for list

Global Flags:
      --context string      The name of the kubeconfig context to use
      --kubeconfig string   Kubernetes configuration file
  -n, --namespace string    The Namespace to use when listing Pods
  ```

  ## Lookup Command

  ```
  Lookup the pods assocaited with a target Pod Disruption Bidget (PDB) resource

Usage:
  pdbl lookup [flags]

Flags:
  -b, --blocking                   Filter for blocking PDB's only
  -t, --blocking-threshold int16   Set the threshold for blocking PDB's. This number is the upper bound for "Allowed Disruptions" for a PDB (Default: 0)
  -h, --help                       help for lookup
      --json                       Output in JSON format
      --no-headers                 Output without column headers

Global Flags:
      --context string      The name of the kubeconfig context to use
      --kubeconfig string   Kubernetes configuration file
  -n, --namespace string    The Namespace to use when listing Pods
  ```