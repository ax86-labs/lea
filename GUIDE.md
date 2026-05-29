# Lea Usage Guide

This guide explains how to install, run, and explore `lea`, and includes diagrams that summarize the system workflows.

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/PizenLabs/lea.git
cd lea

# Build the binary
make build

# Install to your GOPATH/bin
make install
```

## Quick Start

### 1. Index a Repository
Build the structural graph for the current project.

```bash
lea index .
```

### 2. Launch the MCP Server
Expose your repository to AI agents via the Model Context Protocol.

```bash
lea mcp
```

### 3. Browse with the TUI
Explore symbols and relationships interactively.

```bash
lea tui
```

## Common Commands

| Command | Purpose | Example |
| --- | --- | --- |
| `index` | Build/update the structural graph | `lea index .` |
| `tui` | Launch the interactive explorer | `lea tui` |
| `mcp` | Start the MCP server | `lea mcp` |
| `context` | Generate LLM-ready symbol context | `lea context "type:internal/storage/sqlite:Store"` |
| `trace` | Follow call graph paths | `lea trace "func:internal/cli/commands:Execute"` |
| `flow` | Inspect ordered call flow | `lea flow "func:internal/cli/commands:Execute"` |
| `neighbors` | Show immediate dependencies | `lea neighbors AuthService` |
| `impact` | Analyze change impact | `lea impact TokenService` |
| `violations` | Check architectural constraints | `lea violations --config arch.yaml` |
| `watch` | Watch filesystem changes | `lea watch .` |

## Usage Patterns

### AI-Agent Workflow (MCP)

1. **Find the target symbol** to get exact file and symbol coordinates.
2. **Expand context** to pull immediate neighbors (`CALLS`, `USES`, `IMPLEMENTS`).
3. **Trace execution** to map the ordered call graph for the change.
4. **Check boundaries** against architecture rules before committing updates.

### Developer Workflow (CLI/TUI)

- Use `lea tui` for fuzzy symbol search and graph browsing.
- Use `lea context` to generate prompt-ready context for web LLMs.
- Use `lea flow` and `lea trace` to understand execution order and impact.

## Diagrams

### System Architecture

```text
       ┌────────────────────────────────────────────────────────┐
       │                 Local Filesystem Event                 │
       └───────────────────────────┬────────────────────────────┘
                                   │ (fsnotify)
                                   ▼
       ┌────────────────────────────────────────────────────────┐
       │ Incremental Parser Layer (Native Go AST / Tree-sitter) │
       └───────────────────────────┬────────────────────────────┘
                                   │ (Extracted Symbols)
                                   ▼
       ┌────────────────────────────────────────────────────────┐
       │   Storage Layer: SQLite Graph Engine (Recursive CTEs)  │
       └───────────────────────────┬────────────────────────────┘
                                   │
                    ┌──────────────┴──────────────┐
                    ▼                             ▼
       ┌────────────────────────┐    ┌──────────────────────────┐
       │   Integration Layer    │    │     Retrieval Engine     │
       │   (Bubble Tea TUI)     │    │   (MCP Server for AIs)   │
       └────────────────────────┘    └──────────────────────────┘
```

### MCP Query Flow

```mermaid
sequenceDiagram
    autonumber
    actor Agent as AI Agent
    participant Lea as lea (MCP Server)
    database DB as SQLite (Graph Engine)

    Agent->>Lea: find_symbol(name)
    Lea->>DB: Query exact Symbol URI/File
    DB-->>Lea: Return Node
    Lea-->>Agent: Return Symbol Coordinates

    Agent->>Lea: get_neighbors(URI)
    Lea->>DB: Traverse Edges (CALLS, USES, etc.)
    DB-->>Lea: Return Subgraph
    Lea-->>Agent: Return Markdown Context


### Incremental Indexing Loop

```text
[Filesystem Event: Modify/Create/Delete]
                   │
                   ▼
       ┌───────────────────────┐
       │   Debounce & Batch    │ (Gathers changes over 100-300ms)
       └───────────┬───────────┘
                   │
                   ▼
       ┌───────────────────────┐
       │   Invalidation Stage  │ (Deletes affected Nodes & Edges in SQLite)
       └───────────┬───────────┘
                   │
                   ▼
       ┌───────────────────────┐
       │  Incremental Parsing  │ (Native go/ast or Tree-sitter AST extraction)
       └───────────┬───────────┘
                   │
                   ▼
       ┌───────────────────────┐
       │     Graph Commit      │ (Atomic SQL Transaction: Inserts new entities)
       └───────────────────────┘
```

## Troubleshooting

- If the graph is empty, re-run `lea index .` and confirm your repository path is correct.
- For architecture checks, ensure your rules file (for example `arch.yaml`) is present and valid.
- If symbols are missing, confirm the target language parser is supported and available.
