# GCE Machine Database

This package (gcedb) acts as a static "knowledge database" for GCE machine
types. It provides a `GetMachineInfo() (MachineInfo, error)` method which
returns:

```go
struct MachineInfo {
  CPUCores              int
  // Amount of memory in GB (10^9 bytes).
  MemoryGB              float64
  // Network bandwidth in Gbps.
  NetworkGbps           float64
  // Allowed number of local SSDs; empty if local SSD not supported.
  AllowedLocalSSDCount  []int
  // List of allowed storage types, e.g. "pd-ssd", "hyperdisk-balanced".
  StorgeTypes          []string
}
```

For example, for `n2-standard-32`, `MachineInfo()` would return:

```go
MachineInfo{
  CPUCores:             32,
  MemoryGB:             128,
  NetworkGbps:          32,
  AllowedLocalSSDCount: []int{4, 8, 16, 24},
  StorageTypes:         []string{"pd-standard", "pd-balanced", "pd-ssd", "pd-extreme", "hyperdisk-extreme", "hyperdisk-throughput"},
}
```

## Generation

To generate or update the code, use the most up-to-date information available
from Google pages, starting at https://docs.cloud.google.com/compute/docs/ 

Aggregate information from the following regions:
 - asia-southeast1
 - us-central1
 - us-east1
 - us-west2
 - southamerica-east1
 - europe-west1
 - europe-west2
 - europe-west3

If there is any discrepancy between regions, choose the superset of supported
features. We want the database to reflect the most capable configuration of each
machine type.

Based on the gathered information, generate a succinct implementation of
`MachineInfo()` that synthesizes this information (likely having a big `switch`
on the machine type). Try to find formulas that makes the code simpler, for
example many machine types have a fixed ratio of CPU cores to memory which can
be used instead of hardcoding every possibility. Quote relevant documentation
in code, so that a reader can easily verify the accuracy.

The code should support custom machine types using the
information in the machine type itself: for example "n2-custom-16-32768" has 16
CPU cores and 32 GB RAM. Populate the rest of the fields with reasonable
defaults based on the family.

If code is already present, make a reasonable attempt to keep the 
same organization and minimize the code differences.
