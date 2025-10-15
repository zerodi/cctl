# cctl

`cctl` is a CLI tool for managing local Kind clusters and Cluster API workloads, with additional helpers for Proxmox, Talos, and Cilium workflows.

## Features

- Kind cluster lifecycle management (`up`, `down`, `reset`).
- Cluster API bootstrap and manifest deployment.
- Proxmox helpers for Talos images, schematics, and templates.
- Talos/Kubernetes secret retrieval utilities.
- Cilium Helm installation with kube-proxy-free defaults.

## Prerequisites

- Go 1.25 or newer.
- Docker (for kind).
- Access to required external CLIs when running specific commands:
  - `kubectl`, `helm`, `talosctl`, etc.

## Building

```bash
# Build for the current platform
make build

# Cross-compile for specific targets
make build-linux-amd64
make build-darwin-arm64
make build-darwin-amd64
make build-windows-amd64
```

Outputs are placed in the `dist/` directory. Override the binary name with `APP_NAME`:

```bash
make APP_NAME=mycctl build
```

## Usage

```bash
./dist/cctl --help
```

Key command groups:

- `cctl kind …` – manage Kind clusters.
- `cctl capi …` – bootstrap providers and apply manifests.
- `cctl proxmox …` – interact with Proxmox and Talos schematics.
- `cctl secrets …` – fetch kubeconfig/talosconfig secrets and Talos kubeconfigs.
- `cctl cilium …` – deploy Cilium with recommended settings.

Each subcommand exposes `--help` with full options.

## Configuration

`cctl` reads flags from CLI, configuration files, and environment variables (prefixed with `CCTL_`). Legacy environment variables such as `CLUSTER`, `NS`, `OUT_DIR`, etc., remain supported.

To use a config file:

```bash
cctl --config path/to/config.yaml …
```

## License

This project is licensed under the [MIT License](LICENSE).

Vibecoded with ChatGPT.
