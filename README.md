<div align="center">

  <img src="logo-wheel.svg" alt="helm" width="48" height="48"/>

  # helm

  **AI agent orchestration for backend development.**

  <br/>

  [![Go](https://img.shields.io/badge/Go-1.26-00ADD8?style=flat-square&logo=go&logoColor=white)](https://golang.org)
  [![License](https://img.shields.io/badge/license-MIT-green?style=flat-square)](LICENSE)
  [![Platform](https://img.shields.io/badge/platform-linux%20%7C%20macOS-lightgrey?style=flat-square)](https://github.com/getkaze/helm)

  <br/>

  [What is Helm](#what-is-helm) · [Install](#install) · [Usage](#usage) · [Pipeline](#pipeline) · [Agents](#agents) · [Session Management](#session-management) · [Build](#build)

</div>

---

## What is Helm

**Helm** (the helm of a ship — the wheel that steers direction) is a CLI tool that manages the lifecycle of AI agent pipeline sessions. It provides visibility into pipeline progress, session continuity across Claude Code conversations, and state checkpointing.

Helm is a **state manager, not an executor**. Agent logic lives in markdown definitions and runs inside Claude Code. The CLI manages session state, displays progress, and bridges the gap between conversations.

```bash
helm init
helm status
helm resume
helm save
```

---

## Install

```bash
curl -fsSL https://getkaze.dev/helm/install.sh | bash
```

This installs the latest release to `~/.local/bin/helm`. To update later, run `helm update`.

### From source

```bash
git clone https://github.com/getkaze/helm.git
cd helm
make build
```

Binary is output to `bin/helm`. Add it to your PATH or run directly.

### Within Claude Code

Use the `/helm` slash command to activate the orchestrator. The CLI complements Claude Code — use it standalone to check status and checkpoint state.

---

## Usage

```bash
# Initialize a new project
helm init
helm init --name my-api --type brownfield --lang pt-BR
helm init --force                    # reinitialize existing project

# Check pipeline status
helm status                          # colored dashboard
helm status --short                  # one-line summary
helm status --json                   # machine-readable output

# Resume from where you left off
helm resume                          # show context + next steps
helm resume --json                   # structured output

# Checkpoint session state
helm save                            # validate + checkpoint
helm save --message "before refactor"
helm save --force                    # skip recent checkpoint warning

# Update to latest version
helm update

# Other
helm version
helm help
```

### Global Flags

| Flag | Description |
|------|-------------|
| `--no-color` | Disable colored output |

Color is automatically disabled when piping output or when `NO_COLOR` environment variable is set.

---

## Pipeline

Helm guides projects through a structured pipeline of AI agents:

```
DISCOVER    →    PLAN    →    BUILD    →    QUALITY    →    DEPLOY
```

### Greenfield Flow (new project)
```
scout → research → planning → architect → roadmap → breakdown → review → build → verify → ship
```

### Brownfield Flow (existing codebase)
```
survey → research → planning → architect → roadmap → breakdown → review → build → verify → ship
```

### Status Dashboard

```
  Helm v0.1.0

  Project:  my-api
  Type:     brownfield
  Phase:    build
  Profile:  guided
  Language: en-US

  Pipeline:
    [done]  survey       100%
    [done]  research     100%
    [done]  planning     100%
    [done]  architect    100%
    [done]  roadmap      100%
    [done]  breakdown    100%
    [done]  review       100%
    [>>  ]  build        in progress
    [    ]  verify       pending
    [    ]  ship         pending
```

---

## Agents

| Agent | Phase | Role |
|-------|-------|------|
| **scout** | Discover | Explore greenfield project requirements |
| **survey** | Discover | Analyze existing codebase |
| **research** | Discover | Deep research on problems and constraints |
| **planning** | Plan | Write product requirements document |
| **architect** | Plan | Design system architecture |
| **roadmap** | Plan | Define phases and milestones |
| **breakdown** | Plan | Decompose into atomic tasks |
| **review** | Quality | Validate plan traceability (95% gate) |
| **build** | Build | Implement code from task breakdown |
| **verify** | Quality | Test, SAST, code review (95% gate) |
| **ship** | Deploy | Git operations, PR, deployment |

Agent definitions live in `agents/`. Governance rules in `rules/governance.md`.

---

## Session Management

### How it works

1. `helm init` creates `.helm/session.yaml` and `helm.yaml`
2. Agents run inside Claude Code via `/helm` and update session state
3. `helm save` checkpoints state for safe session handoff
4. `helm resume` shows where to pick up in a new Claude Code session

### File Structure

```
.helm/                    # Runtime state (gitignored)
  session.yaml            # Current session state
  session.yaml.backup     # Auto-backup before every write
  handoffs/               # Agent-to-agent handoff documents
  artifacts/              # Per-agent output (reports, specs)
  checkpoints/            # Session snapshots (max 5, FIFO rotation)

helm.yaml                 # Project config (committed)
```

### Checkpoints

`helm save` validates session integrity before checkpointing:
- Checks required fields in session.yaml
- Verifies handoff files exist for completed agents
- Verifies artifact directories exist for completed agents
- Rotates old checkpoints (keeps last 5)
- Auto-backs up session.yaml before writing

---

## Build

```bash
make build                # current platform
make build-linux          # Linux amd64
make build-linux-arm64    # Linux arm64
make build-darwin         # macOS amd64
make build-darwin-arm64   # macOS arm64
make build-all            # all targets
```

Binaries are output to `bin/`.

### Development

```bash
make build
bash install-dev.sh       # install dev build to ~/.local/bin
make test                 # run tests with race detection
```

### Stack

| Layer | Technology |
|-------|-----------|
| Language | Go 1.26 |
| CLI Framework | spf13/cobra |
| YAML | gopkg.in/yaml.v3 |
| Color | fatih/color |
| Build | Make + ldflags version injection |

---

## Star History

[![Star History Chart](https://api.star-history.com/image?repos=getkaze/helm&type=date&legend=top-left)](https://www.star-history.com/?repos=getkaze%2Fhelm&type=date&legend=top-left)

---

## License

MIT — see [LICENSE](LICENSE).
