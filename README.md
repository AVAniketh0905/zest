# Zest

## About

A cross-platform CLI tool that launches project-specific workspaces, automating the opening of browser tabs, editors, terminals, containers, notes, and other apps via a simple, user-defined YAML config.

Designed to eliminate repetitive setup and enable seamless development environments with a single command.

![Zest CLI Tests](https://github.com/AVAniketh0905/zest/actions/workflows/cli-test.yml/badge.svg?branch=main)

## Zest Configuration

Zest configuration and state directory structure:

```bash
$HOME/.zest/
├── zest.yaml                        // Global configuration file (user editable)
├── workspace/                       // Per-workspace configuration (user editable)
│   ├── [name of wsp].yaml           // Config for each workspace
├── state/                           // Internal state files (NOT user editable)
│   ├── workspace.json               // Overall state of all workspaces
│   ├── other_future_cmds.json       // Additional future commands/state
│   └── workspace/                   // Per-workspace state files
│       ├── [name of wsp].json       // State for each workspace
```
