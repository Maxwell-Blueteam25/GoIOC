# GoIOC

A portable, agentless IOC scanner for Windows (and cross-platform compatible).

This is a complete rewrite of [BluewaveSweeper](https://github.com/Maxwell-Blueteam25/Bluewave-IR-Toolkit/tree/main/Scripts/Bluewave-Sweeper) in Go. It sweeps endpoints for specific file names, hashes, registry keys, and processes defined in a JSON profile.

The logic is derived from the "Collector" methodology outlined in *Incident Response & Computer Forensics* (Mandia). It implements that concept using a static binary to automate the search for specific artifacts across a fleet without requiring an EDR agent or .NET dependencies.

## Capabilities

* **Static Binary:** Compiles to a single `.exe`. No PowerShell version requirements, no .NET dependencies, and no external libraries on the target.
* **Memory Safe:** Uses streaming I/O for file hashing. Can hash multi-gigabyte files without spiking RAM.
* **Context-Aware:** Implements "Constraint Logic." If a path is defined in the profile, the engine targets that directory directly (O(1)) rather than crawling the entire disk.
* **Transport:** Supports local output, SMB copy, or direct upload to Azure Blob Storage via REST API (SAS Token).

## Repository Structure

| File | Description |
| :--- | :--- |
| `main.go` | Entry point. Handles config parsing, validation, and module dispatch. |
| `go.mod` | Go module definition and dependency tracking. |
| `engines/` | Core logic packages (`hash.go`, `reg.go`, `process.go`, `file.go`). |
| `sweeper_config.json` | Sample configuration file defining the IOCs. |
| `sweeper-schema.json` | JSON schema used to validate the config file before execution. |

## Supported Indicators

The engine supports five logic modules:

* **file_name**: Scans for specific filenames.
* **file_hash**: Scans for SHA256 hashes. Supports path constraints to limit scope.
* **registry_key**: Checks for the existence of specific keys.
* **registry_value**: Checks for specific values (data) within a key.
* **process_name**: Checks for running processes and captures PIDs/Command Lines.

## Workflow

### 1. Build
Compile the tool into a static binary.
```bash
go mod tidy
go build -o GoIOC.exe .
```

### 2. Configure

Define target indicators in `sweeper_config.json`.

JSON

```
{
  "sweeps": [
    {
      "type": "process_name",
      "value": "nc.exe"
    },
    {
      "type": "file_hash",
      "value": "E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855",
      "constraints": { "path": "C:\\Windows\\Temp" }
    }
  ]
}
```

### 3. Deploy and Sweep

Copy `GoIOC.exe` and `sweeper_config.json` to the target machine.

**Mode: Local (Default)** Saves the JSON report to the current directory.

PowerShell

```
.\GoIOC.exe sweeper_config.json Local
```

**Mode: SMB** Saves the report locally, then copies it to a specified share.

PowerShell

```
.\GoIOC.exe sweeper_config.json SMB "\\Server\EvidenceShare"
```

**Mode: Cloud (Azure Blob)** Saves the report locally, then uploads it via REST API (PUT) to an Azure Blob Container.

PowerShell

```
.\GoIOC.exe sweeper_config.json Cloud "" "https://<account>.blob.core.windows.net/<container>?<sastoken>"
```

## Output Format

Reports are generated as JSON files: `Report_<HOSTNAME>_<TIMESTAMP>.json`.

**Sample Output:**

JSON

```
[
  {
    "Hostname": "DESKTOP-01",
    "Timestamp": "2026-01-19T14:00:00Z",
    "Type": "process_name",
    "Value": "svchost.exe",
    "Status": "DIRTY",
    "Details": "PID: 1044 | CMD: C:\\Windows\\System32\\svchost.exe -k netsvcs"
  }
]
```
