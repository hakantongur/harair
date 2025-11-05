# Harair

A Go-based CLI tool for mirroring Harbor container registries across air-gapped environments.

## Features
- Sync projects, repos, and tags between registries
- Run `skopeo` automatically in Docker
- Rule-based sync with `rules.yaml`
- Parallel copy (`--concurrency`)
- Works fully offline
