# Project Structure

The repository is organized as follows:

```plaintext
k8stool/
├── cmd/                # Main command-line entry point
├── internal/          # Internal packages
│   ├── k8s/          # Kubernetes client wrappers
│   └── cli/          # Command-line interface components
├── pkg/              # Public packages
├── docs/             # Documentation
├── images/           # Project images and assets
├── .github/          # GitHub workflows and templates
├── go.mod           # Go module definition
├── go.sum           # Go module checksums
├── .goreleaser.yml  # GoReleaser configuration
├── .golangci.yml    # GolangCI-Lint configuration
├── .air.toml        # Air live reload configuration
├── LICENSE          # Project license
└── README.md        # Project documentation
```

## Directory Details

### Core Code
- `cmd/`: Contains the main application entry point
- `internal/`: Private packages used only within this project
  - `k8s/`: Kubernetes client implementations
  - `cli/`: Command-line interface logic
- `pkg/`: Public packages that could be used by external projects

### Documentation
- `docs/`: MkDocs documentation files
- `images/`: Project images and screenshots
- `README.md`: Project overview and quick start

### Configuration
- `.github/`: GitHub-specific configurations and workflows
- `.goreleaser.yml`: Release automation configuration
- `.golangci.yml`: Linter configuration
- `.air.toml`: Development live reload settings

### Dependencies
- `go.mod`: Go module dependencies
- `go.sum`: Dependency checksums