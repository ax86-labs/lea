# `ctxd` Operational Architecture & Workflow

This document details the internal lifecycle, workflow mechanics, and structural context routing of `ctxd`. It illustrates how the system transitions from raw source code to a real-time deterministic knowledge graph, serving high-signal context to AI Agents and developers.

---

## 1. Core Reflex Loop (The Underlying Engine)
`ctxd` operates on a continuous, event-driven reflex loop powered by `fsnotify`. It ensures that the structural graph stored in SQLite remains synchronized with your filesystem in real-time, operating with bounded execution latency. ```text
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

---

## 2. Automated AI-Agent Workflow (Via MCP Server)

This is the primary automated pipeline. Instead of forcing human intervention, `ctxd` acts as a **Structural Language Server for LLMs**, allowing AI agents (e.g., Claude Code, Aider) to query the codebase deterministically through the Model Context Protocol (MCP).

```
   ┌──────────┐             ┌──────────┐             ┌──────────┐
   │ AI Agent │             │   ctxd   │             │  SQLite  │
   └────┬─────┘             └───┬──────┘             ────┬──────
        │                       │                        │
        │ 1. find_symbol(name)  │                        │
        ├──────────────────────>│                        │
        │                       │─── 2. Query exact ────>│
        │                       │    Symbol URI/File     │
        │                       │<── 3. Return Node ─────┤
        │<── 4. Return Symbol ──┤                        │
        │                       │                        │
        │ 5. get_neighbors(URI) │                        │
        ├──────────────────────>│                        │
        │                       │─── 6. Traverse Edges ─>│
        │                       │    (CALLS, USES, etc.) │
        │                       │<── 7. Return Graph ────┤
        │<── 8. Return Context ─┤                        │
        │                       │                        │

```

### Step-by-Step AI Execution Path:

1. **Symbol Discovery (`find_symbol`)**: The user instructs the AI to modify a feature. The AI initiates by requesting the exact coordinates (File, Line, Symbol Type) of the target component.
2. **Context Expansion (`get_symbol_neighbors`)**: Instead of swallowing whole directories, the AI requests adjacent dependencies. `ctxd` exposes exactly what interfaces the symbol implements and what structures it uses.
3. **Execution Trace (`trace_execution_path`)**: The AI maps the execution control flow. `ctxd` leverages SQLite Recursive CTEs to output an ordered hierarchy of function call paths.
4. **Guardrail Check (`find_architecture_violations`)**: Before committing modifications to disk, the AI runs a compliance check against `.ctxd/architecture.yaml` to ensure no architectural boundary constraints are broken.

---

## 3. Human Developer Workflow (CLI & TUI)

Designed for developer onboarding, rapid codebase exploration, and manual context synthesis for web-based LLMs.

### Scenario A: Codebase Exploration via TUI

```bash
ctxd tui

```

* **Fuzzy Match**: The developer inputs a partial symbol string into the Bubble Tea TUI.
* **Graph Exploration**: Using keyboard navigations, the developer expands nodes to reveal structural execution branches and dependency vectors visually without ever opening separate files.

### Scenario B: Manual Context Synthesis for Web LLMs

```bash
ctxd context "type:internal/storage/sqlite:Store" > prompt_context.md

```

* **Entropy Minimization**: `ctxd` extracts the requested target, its immediate dependencies, its interface mapping, and execution constraints into a highly dense, markdown-optimized layout.
* **Token Optimization**: The developer drops `prompt_context.md` into Claude/GPT. The LLM instantly gains system architect-level precision while using **90% fewer tokens** than standard flat-text codebase dumps.

---

## 4. Context Compilation Blueprint

When a retrieval request is executed, `ctxd` builds a **High-Signal Markdown Context**. This payload is structured syntactically to match the token-attention mechanisms of modern LLMs:

```markdown
### Symbol: [Symbol URI]
- Type: [Struct | Function | Interface | Method]
- Source: `path/to/file.go`

#### Structural Synapses (Edges)
- [IMPLEMENTS] -> Interface: `Writer`
- [USES] -> Struct: `Config` (via field `cfg`)

#### Execution Topography (Call Graph)
- Inbound Calls: `cli.Execute()` -> `main.go`
- Outbound Calls: `sqlite.Connect()` -> `store.go`

#### Architectural Boundaries
- [Constraint] Boundary: `storage`
- [Status] Compliant (No restricted packages imported)

```

---

## 5. System Characteristics

* **Deterministic over Probabilistic**: Zero semantic guessing during indexing. Code relationships are extracted using exact AST nodes, offering 100% mathematical certainty.
* **Stateful Edge Resolution**: When a single file updates, `ctxd` maintains inbound edges from unmodified files, preserving global graph integrity without requiring a full repository re-index.
* **Zero Cloud Footprint**: Runs fully local, processing files out-of-the-box, ensuring total source privacy and zero network overhead.

