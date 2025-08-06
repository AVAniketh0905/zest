# Zest

[![Zest CLI Tests](https://github.com/AVAniketh0905/zest/actions/workflows/cli_os_test.yaml/badge.svg)](https://github.com/AVAniketh0905/zest/actions/workflows/cli_os_test.yaml)

Zest is a **cross-platform CLI tool** for managing isolated, project-specific **workspaces** with a single command.

Workspaces are defined using YAML config files and can automate the launch of:
- Browsers
- Editors
- Terminals
- Containers
- Notes
- Other custom apps

Ideal for developers, learners, and professionals who switch between multiple projects or contexts.

---

## Features

- Create and launch workspaces effortlessly
- Cross-platform support (Linux, macOS, Windows)
- Shell autocompletion support
- Easily extendable with templates

---

## Installation

You can install Zest using Go:

```bash
go install github.com/AVAniketh0905/zest@latest
```

### Prerequisites

- Go 1.18 or higher must be installed and properly configured.
- Ensure that $GOPATH/bin is in your system's PATH, so you can run zest from anywhere.

Once installed, you can verify it by running:

```bash
zest -v
```

---

## Usage

```bash
zest [command]

Available Commands:
  close       Close an existing or active workspace
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  init        Initialize a new workspace
  launch      Launch a workspace
  list        List all available workspaces
  status      Show the live status of one or more workspaces
```

---

## Zest Configuration

Zest configuration and state directory structure:

```bash
$HOME/.zest/
├── zest.yaml                        // Global configuration file (user editable)
├── workspace/                       // Per-workspace configuration (user editable)
│   ├── [name of wsp].yaml           // Config for each workspace
├── state/                           // Internal state files (NOT user editable)
│   ├── workspace.json               // Overall state of all workspaces
│   └── workspace/                   // Per-workspace state files
│       ├── [name of wsp].json       // State for each workspace
```
---

## Examples

1. Browsers:

Support for `brave`.

```yaml
brave:
  - tabs:
      - "https://example.com"
      - "https://github.com/your/repo"
    profile_dir: "path/to/profile"
    args:
      - "--no-first-run"
```

2. Code Editors:

Support for `vscode`.

```yaml
vscode:
  - path: path/to/project            # Optional; defaults to current working directory
    args: 
      - "--new-window"
```

3. Terminals:

Support for `powershell`.

```yaml
powershell:
  - path: path/to/project            # Optional; defaults to current working directory
    tabs:
      - "go version"
      - "git status"
    args:
      - "-NoExit"
```

4. Pdf Viewers:

Support for `sioyek`.

```yaml
sioyek:
  - files:
      - "path/to/file1.pdf"
      - "path/to/file2.pdf"
```

5. Custom App:

Support for launching any custom executable:

```yaml
custom:
  - name: brave
    cmd: cmd
    args:
      - /C
      - start
      - ""
      - brave.exe
      - --user-data-dir=path/to/profile
      - --no-first-run

  - name: powershell
    cmd: wt
    args:
      - -w
      - new-tab
      - powershell.exe
      - -NoExit
      - -Command
      - cd path/to/project
```

---

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

---
