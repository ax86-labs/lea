# lea 

[![Go Version](https://img.shields.io/github/go-mod/go-v/PizenLabs/lea)](https://golang.org/doc/devel/release.html)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Build Status](https://img.shields.io/github/actions/workflow/status/PizenLabs/lea/ci.yml?branch=main)](https://github.com/PizenLabs/lea/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/PizenLabs/lea)](https://goreportcard.com/report/github.com/PizenLabs/lea)


**Structural context operating system for AI-native software engineering.**

`lea` is a terminal-first structural memory system designed to help AI models and developers navigate and understand large codebases with minimal context and maximum precision.

Unlike traditional RAG systems that treat code as flat text, `lea` models repositories as **living structural graphs**, preserving the deterministic relationships (calls, dependencies, implementations) that define software systems.

---

##  The Vision

Modern AI coding systems suffer from context window limitations, token inflation, and "context entropy." Most rely on probabilistic semantic chunking (embeddings), which often loses the architectural "big picture."

**Software is symbolic, not just semantic.** `lea` focuses on:
1.  **Structural Retrieval First**: Symbols, dependencies, call graphs, and architectural boundaries.
2.  **Semantic Retrieval Second**: Natural language understanding on top of structural certainty.

---

##  Key Features

-   **Multi-Language AST Indexing**: Native support for **Go** (`go/ast`) and **Python**, **TypeScript**, and **Rust** via Tree-sitter.
-   **Structural Graph Engine**: Models your codebase as a graph of functions, structs, interfaces, and their relationships (`CALLS`, `IMPLEMENTS`, `USES`).
-   **AI Context Compiler**: Generates high-signal, markdown-optimized context for LLMs (Claude, GPT, Gemini) using deterministic retrieval.
-   **Model Context Protocol (MCP)**: Expose your codebase structure directly to AI agents via a standardized protocol.
-   **Interactive TUI**: A rich, terminal-based explorer for fuzzy symbol navigation and dependency browsing.
-   **Control Flow & Architecture**: Trace execution paths and detect boundary violations against architectural constraints.
-   **Incremental & Reactive**: Real-time graph updates using `fsnotify` without re-indexing the entire repository.
-   **Local-First**: Powered by an embedded SQLite database. Works offline and over SSH.

---

##  Architecture

`lea` is built with a modular, performance-oriented architecture designed for local execution:

-   **Parser Layer**: Pluggable parsers using native ASTs and Tree-sitter for high-fidelity symbol extraction.
-   **Graph Engine**: A high-performance relationship model that treats your codebase as a first-class graph.
-   **Storage Layer**: SQLite-backed storage utilizing recursive CTEs for complex graph traversals.
-   **Integration Layer**: Built-in MCP server for AI agents and a Bubble Tea-powered TUI for humans.

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

##  Installation

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

---

##  Quick Start

### 1. Index your project
Initialize the structural graph for your repository.
```bash
lea index .
```

### 2. Start the MCP Server
Connect your favorite AI agent (like Claude Code or Aider) directly to your codebase.
```bash
lea mcp
```

### 3. Interactive Exploration
Launch the TUI to browse symbols and relationships visually.
```bash
lea tui
```

---

##  Command Guide

| Command | Description | Example |
| :--- | :--- | :--- |
| `index` | Build or update the structural graph | `lea index .` |
| `tui` | Open the interactive symbol explorer | `lea tui` |
| `mcp` | Start the Model Context Protocol server | `lea mcp` |
| `context` | Generate LLM-optimized context for a symbol | `lea context "type:internal/storage/sqlite:Store"` |
| `trace` | Follow the call graph from a specific function | `lea trace "func:internal/cli/commands:Execute"` |
| `flow` | Inspect ordered call flow within a symbol | `lea flow "func:internal/cli/commands:Execute"` |
| `neighbors` | Find immediate dependencies of a symbol | `lea neighbors AuthService` |
| `impact` | Analyze the impact of changing a symbol | `lea impact TokenService` |
| `violations` | Check for architectural boundary violations | `lea violations --config arch.yaml` |
| `watch` | Watch for file changes and update the graph | `lea watch .` |

---

##  Roadmap

- [x] **Phase 1: MVP**: Go parser, SQLite storage, basic graph queries.
- [x] **Phase 2: AI Context Layer**: High-signal markdown generation and context compilation.
- [x] **Phase 3: Incremental Updates**: Real-time file watching and partial re-indexing.
- [x] **Phase 4: MCP Integration**: Standardized protocol for AI agent connectivity.
- [x] **Phase 5: Interactive TUI**: Fuzzy navigation and visual dependency exploration.
- [x] **Phase 6: Multi-Language Support**: Tree-sitter integration for Python, Rust, and TypeScript.
- [x] **Phase 7: Advanced Retrieval**: Control flow analysis and architecture guardrails.

---

##  Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines on development, testing, and pull requests.

---

##  License

`lea` is licensed under the [MIT License](LICENSE).

---

Built for the future of AI-native engineering. 🦾
