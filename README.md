# ctxd

**Structural context operating system for AI-native software engineering.**

`ctxd` is a terminal-first structural memory system designed to help AI models and developers navigate and understand large codebases with minimal context and maximum precision.

Unlike traditional RAG systems that treat code as flat text, `ctxd` models repositories as living structural graphs, preserving the deterministic relationships (calls, dependencies, implementations) that define software systems.

---

##  The Vision

Modern AI coding systems suffer from context window limitations, token inflation, and "context entropy." Most rely on probabilistic semantic chunking (embeddings), which often loses the architectural "big picture."

**Software is symbolic, not just semantic.** `ctxd` focuses on:
1. **Structural Retrieval First**: Symbols, dependencies, call graphs, and architectural boundaries.
2. **Semantic Retrieval Second**: Natural language understanding on top of structural certainty.

---

##  Key Features

- **Multi-Language AST Indexing**: Built-in support for **Go** (native `go/ast`) and **Python**, **TypeScript**, and **Rust** via Tree-sitter.
- **Structural Graph Engine**: Models your codebase as a graph of functions, structs, interfaces, and their relationships (`CALLS`, `IMPLEMENTS`, `USES`).
- **AI Context Compiler**: Generates high-signal, markdown-optimized context for LLMs (Claude, GPT, Gemini).
- **Model Context Protocol (MCP)**: Expose your codebase structure directly to AI agents via a standardized protocol.
- **Interactive TUI**: A rich, terminal-based explorer for fuzzy symbol navigation and dependency browsing.
- **Control Flow & Architecture Violations**: Ordered execution paths and boundary checks against architecture constraints.
- **Incremental & Reactive**: Real-time graph updates using `fsnotify` without re-indexing the entire repository.
- **Local-First**: Powered by an embedded SQLite database. Works offline and over SSH.

---

##  Installation

```bash
# Clone the repository
git clone https://github.com/ax86-labs/lea.git
cd ctxd

# Build the binary
go build -o ctxd ./cmd/ctxd/main.go

# (Optional) Move to your path
mv ctxd /usr/local/bin/
```

---

##  Usage

### 1. Index your project
Initialize the structural graph for your repository.
```bash
ctxd index .
```

### 2. Interactive Exploration
Launch the TUI to browse symbols and relationships.
```bash
ctxd tui
```

### 3. Trace Execution
Follow the call graph starting from a specific function.
```bash
ctxd trace "func:internal/cli/commands:Execute"
```

### 4. Control Flow Ordering
Inspect the ordered call flow within a symbol.
```bash
ctxd flow "func:internal/cli/commands:Execute"
```

### 5. Architecture Violations
Detect boundary violations using an architecture config.
```bash
ctxd violations --config .ctxd/architecture.yaml
```

### 6. AI Context Generation
Generate high-signal markdown for your LLM prompts.
```bash
ctxd context "type:internal/storage/sqlite:Store"
```

### 7. MCP Server
Connect your favorite AI agent directly to `ctxd`.
```bash
ctxd mcp
```

---

##  Architecture

`ctxd` is built with a modular, performance-oriented architecture:
- **Parser Layer**: Pluggable parsers using native ASTs and Tree-sitter.
- **Structural Graph Engine**: High-performance relationship modeling.
- **Storage Layer**: SQLite-backed graph storage with recursive CTE support.
- **Retrieval Engine**: Advanced queries for neighbors, impact analysis, and call tracing.
- **Integration Layer**: MCP and TUI for human/AI interaction.

---

##  Roadmap

- [x] **Phase 1: MVP**: Go parser, SQLite storage, basic graph queries.
- [x] **Phase 2: AI Context Layer**: High-signal markdown generation.
- [x] **Phase 1.5: Incremental Updates**: File watching (`fsnotify`) and partial re-indexing.
- [x] **Phase 3: MCP Integration**: Expose `ctxd` as a Model Context Protocol server.
- [x] **Phase 4: Interactive TUI**: Fuzzy navigation and visual dependency trees.
- [x] **Phase 5: Multi-Language Support**: Tree-sitter integration for Python, TypeScript, and Rust.
- [x] **Phase 6: Advanced Retrieval**: Control flow analysis and architecture violation detection.

---

## Contributing

Contributions are welcome! Please read [CONTRIBUTING.md](CONTRIBUTING.md) for setup, linting, and PR guidelines.

---

## License

This project is licensed under the [MIT License](LICENSE).

---

Built for the future of AI-native engineering. 🦾
