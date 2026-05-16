# ctxd

**Structural context operating system for AI-native software engineering.**

`ctxd` is a terminal-first structural memory system designed to help AI models and developers navigate and understand large codebases with minimal context and maximum precision.

Unlike traditional RAG systems that treat code as flat text, `ctxd` models repositories as living structural graphs, preserving the deterministic relationships that define software systems.

---

##  The Vision

Modern AI coding systems suffer from context window limitations, token inflation, and "context entropy." Most rely on probabilistic semantic chunking (embeddings), which often loses the architectural "big picture."

**Software is symbolic, not just semantic.** `ctxd` focuses on:
1. **Structural Retrieval First**: Symbols, dependencies, call graphs, and architectural boundaries.
2. **Semantic Retrieval Second**: Natural language understanding on top of structural certainty.

---

##  Key Features

- **AST-Powered Indexing**: Compiler-grade parsing using `go/ast` (and soon Tree-sitter).
- **Structural Graph Engine**: Models your codebase as a graph of functions, structs, interfaces, and their relationships (`CALLS`, `IMPLEMENTS`, `USES`).
- **AI Context Compiler**: Generates high-signal, markdown-optimized context for LLMs (Claude, GPT, Gemini).
- **Local-First & Fast**: Powered by an embedded SQLite database. Works offline and over SSH.
- **Deterministic Traceability**: Map execution paths and impact zones with precision.

---

##  Installation

```bash
# Clone the repository
git clone https://github.com/andev0x/ctxd.git
cd ctxd

# Build the binary
go build -o ctxd ./cmd/ctxd/main.go

# (Optional) Move to your path
mv ctxd /usr/local/bin/
```

---

##  Usage

### 1. Index your project
Initialize the structural graph for the current directory.
```bash
ctxd index .
```

### 2. Explore Relationships
Find symbols directly related to a specific component.
```bash
ctxd neighbors "type:internal/storage/sqlite:Store"
```

### 3. Trace Execution
Follow the call graph starting from a specific function.
```bash
ctxd trace "func:internal/cli/commands:Execute"
```

### 4. Impact Analysis
Identify what depends on a specific package or symbol.
```bash
ctxd impact "pkg:internal/storage/sqlite"
```

### 5. AI Context Generation
Generate a high-signal markdown summary for an LLM prompt.
```bash
ctxd context "type:internal/storage/sqlite:Store"
```

---

##  Architecture

`ctxd` is built with a modular architecture:
- **Parser Engine**: Extracts symbols and relationships using native language tools.
- **Storage Layer**: SQLite-backed graph storage with recursive query support.
- **Context Compiler**: The bridge between the structural graph and AI prompts.
- **CLI/TUI**: Terminal-native interfaces for developers.

---

##  Roadmap

- [x] **Phase 1: MVP**: Go parser, SQLite storage, basic graph queries.
- [x] **Phase 2: AI Context Layer**: High-signal markdown generation.
- [x] **Phase 1.5: Incremental Updates**: File watching (`fsnotify`) and partial re-indexing.
- [x] **Phase 3: MCP Integration**: Expose `ctxd` as a Model Context Protocol server.
- [ ] **Phase 4: Interactive TUI**: Fuzzy navigation and visual dependency trees.
- [ ] **Phase 5: Multi-Language**: Tree-sitter integration for Rust, TypeScript, and Python.

---

##  License

This project is licensed under the [MIT License](LICENSE).

---

Built for the future of AI-native engineering. 🦾
