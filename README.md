
# GoIOC

**A portable, agentless forensic sweep and live-response tool for high-velocity incident response.**

GoIOC (rewrite of [BluewaveSweeper](https://github.com/Maxwell-Blueteam25/Bluewave-IR-Toolkit/tree/main/Scripts/Bluewave-Sweeper) in Go) is a static binary designed to IOC sweeps and IR triage, specifically for systems without EDR agents.

Inspired by Mandiant's Redline to conduct IOC Deployment's across systems in a network.

## Capabilities

- **Zero Dependency:** Compiles to a single, static `.exe`. Runs natively on the target.
    
- **Dispatch Architecture:** Uses strict JSON schema validation to route indicators to modular engines (Hash, Registry, Process, File).
    
- **Constraints:** Explicit path targeting to avoid expensive full-disk recursion.
    
- **Flash Triage (New):** Optional "Live Collection" engine captures volatile state (Process Tree, Active Network Connections, Services, Persistence Keys) for context.
    
- **Transport Agnostic:** Supports local JSON output, SMB replication, or direct Azure Blob upload via REST API (SAS Token) to decouple reporting from local disk I/O.
    

## Repository Structure

Plaintext

```
.
├── collections/       # Live Response Engine (Process, Net, Svc, Reg)
├── configs/           # Sample JSON IOC profiles
├── engines/           # Core IOC Logic (Hash, Registry, Process, File)
├── main.go            # Entry Point & Dispatcher
├── go.mod             # Dependencies (gopsutil, gojsonschema)
├── go.sum             # Checksums
├── .gitignore         # Build artifacts
└── README.md          # Documentation
```

## Supported Indicators

The `engines/` package supports five detection modules:

- **`file_name`**: Target specific filenames (e.g., `mimikatz.exe`).
    
- **`file_hash`**: Streaming SHA256 verification (memory safe via `io.Copy`). Supports path constraints.
    
- **`registry_key`**: Validates existence of keys (e.g., `HKLM\..\Run`).
    
- **`registry_value`**: Scans for specific malicious data payload strings.
    
- **`process_name`**: Enumerates running processes, capturing PIDs and Command Lines.
    

## Workflow

### 1. Build

Compile the tool into a static binary. (Flags strip debug symbols for smaller footprint).

Bash

```
go mod tidy
go build -ldflags "-s -w" -o GoIOC.exe .
```

### 2. Configure

Define target indicators in `configs/sweeper_config.json`.

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
      "value": "E3B0C44298FC...B855",
      "constraints": { "path": "C:\\Windows\\Temp" }
    }
  ]
}
```

### 3. Deploy & Execute

Copy `GoIOC.exe` and your config to the target. Arguments are **positional** and strict.

Syntax:

GoIOC.exe <config> <Mode> <SMBPath> <SasURL> <Collections>

#### Scenario A: Local Scan + Live Collections

Saves JSON to disk. Enables telemetry (Network/Process/Service lists).

> **Note:** Use `"none"` as a placeholder for unused middle arguments.

PowerShell

```
.\GoIOC.exe configs/sweeper_config.json Local "none" "none" true
```

#### Scenario B: SMB Exfiltration (Stealth)

Copies report to a network share. Collections disabled.

PowerShell

```
.\GoIOC.exe configs/sweeper_config.json SMB "\\10.1.1.50\Evidence" "none" false
```

#### Scenario C: Cloud Direct (Azure Blob)

Uploads directly via REST API. Useful for off-network endpoints.

PowerShell

```
.\GoIOC.exe configs/sweeper_config.json Cloud "none" "https://<account>.blob.core.windows.net/<container>?<sastoken>" true
```

## Data Model

Reports are generated as structured JSON (`Report_<HOSTNAME>_<TIMESTAMP>.json`).

JSON

```
{
  "metadata": {
    "hostname": "WORKSTATION-01",
    "timestamp": "2026-01-19T14:00:00Z",
    "scan_id": "WORKSTATION-01_20260119_140000"
  },
  "ioc_matches": [
    {
      "Type": "process_name",
      "Value": "svchost.exe",
      "Status": "DIRTY",
      "Details": "PID: 1044 | CMD: C:\\Windows\\Temp\\svchost.exe"
    }
  ],
  "collections": {
    "processes": [ ... ],
    "network": [
      {
        "protocol": "tcp",
        "local_address": "192.168.1.105:49812",
        "remote_address": "104.21.55.2:443",
        "state": "ESTABLISHED",
        "pid": 1044
      }
    ],
    "services": [ ... ],
    "registry": [ ... ]
  }
}
```

## Requirements

- **Build:** Go 1.21+
    
- **Target:** Windows 10/11 or Server 2012+ (No agent required).
    
- **Privileges:** Admin rights recommended for full Process/Registry visibility.
