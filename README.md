# 🛰️ Harair — Harbor Air-Gap Synchronization CLI

**Harair** is a Go-based command-line tool designed to **mirror container images and Helm charts between Harbor registries**, even across **air-gapped networks**.  
It automates `skopeo` operations, manages configuration via YAML, and supports parallel transfers, network-isolated Docker execution, and rule-based project synchronization.

---

## ✨ Key Features

- 🔁 **Registry Mirroring** — Sync projects, repositories, and tags between two Harbor instances.
- 🧱 **Air-Gap Mode** — Works fully offline using Docker-based Skopeo.
- ⚙️ **Rules-Based Filtering** — Define includes/excludes and tag patterns in a `rules.yaml` file.
- 🚀 **Parallel Copy** — Multi-threaded transfers with `--concurrency`.
- 🧩 **Docker Network Support** — Run `skopeo` inside an isolated Docker network (`--docker-network`).
- 🧾 **Dry-Run Mode** — Preview all copy operations before executing.
- 🗂️ **Simple Config** — Define multiple registries in a single `config.yaml`.
- 🪶 **Lightweight** — Built entirely in Go; no dependencies beyond Docker or Skopeo.

---

## 🧩 Architecture Overview

```text
+-----------------------------+
|           CLI (Go)          |
|  └── Commands:              |
|      login, ls, sync,       |
|      sync-direct            |
+-------------┬---------------+
              │
              ▼
+---------------------------------------------+
|         Docker Engine (local runtime)       |
|  └── Runs "quay.io/skopeo/stable" container |
+---------------------------------------------+
              │
   ┌──────────┴─────────────┐
   ▼                        ▼
Harbor #1 (source)     Harbor #2 (destination)
reg1:5000/demo/...     reg2:5000/demo/...
   │                        │
   └── Inside same Docker network ("airgap")
