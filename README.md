# Cov

Cov is a tool to analyze the go coverage output file.

## Install

```console
go get -u go.aporeto.io/cov
```

## Usage

```console
Usage of cov:
      --branch string       The branch to use to check the patch coverage against. Example: master
      --coverage strings    The coverage files to use. [required]
      --filter strings      The filters to use for coverage lookup
      --log-format string   Log format (default "console")
      --log-level string    Log level (default "info")
      --name string         Meaning full name to use for output (default "Project")
      --quiet               Do not print details just the verdict
      --target int          The target of coverage in percent that is requested
  -v, --version             Display the version
```

When the `--branch BASE` is used a diff will be done between your current branch and the branch passed as base to identify the files you changed.

You can pass several coverage files they all will be merged.

You can also filter for a given package or any substring.

When `--target X` is set, the output will be colored given that target and the return will be 1 if target is not reached.

## Examples

Show coverage for one package and unit test coverage:
```console
cov --coverage unit_coverage.cov --filter yuna

[17%] go.aporeto.io
└── [17%] backend
    └── [17%] srv
        └── [17%] yuna
            └── [17%] internal
                ├── [88%] constraints
                │   └── [88%] constraints.go
                └── [8%] processors
                    ├── [6%] discoverymode.go
                    ├── [0%] export.go
                    ├── [0%] import.go
                    └── [12%] importreferences.go

Project coverage: 17%
```

Show coverage for one package and unitest test plus integration test coverage:

```console
cov --coverage unit_coverage.cov --coverage integration_coverage.cov --filter yuna

[85%] go.aporeto.io
└── [85%] backend
    └── [85%] srv
        └── [85%] yuna
            ├── [85%] internal
            │   ├── [80%] configuration
            │   │   └── [80%] configuration.go
            │   ├── [88%] constraints
            │   │   └── [88%] constraints.go
            │   ├── [50%] db
            │   │   └── [50%] db.go
            │   ├── [100%] errors
            │   │   └── [100%] errors.go
            │   ├── [75%] importing
            │   │   └── [75%] import.go
            │   └── [85%] processors
            │       ├── [88%] discoverymode.go
            │       ├── [80%] export.go
            │       ├── [100%] import.go
            │       └── [81%] importreferences.go
            └── [89%] main.go

Project coverage: 85%
```


Show coverage for a pull request against master:

```console
cov --coverage unit_coverage.cov --coverage integration_coverage.cov --branch master

[60%] go.aporeto.io
└── [60%] backend
    └── [60%] srv
        ├── [58%] vargid
        │   └── [58%] internal
        │       ├── [49%] common
        │       │   └── [49%] common.go
        │       ├── [52%] configuration
        │       │   └── [52%] configuration.go
        │       ├── [59%] mode
        │       │   └── [59%] scheduler_mode.go
        │       └── [71%] notifications
        │           └── [71%] policies.go
        ├── [94%] yeul
        │   └── [94%] internal
        │       └── [94%] configuration
        │           └── [94%] configuration.go
        └── [81%] yuffie
            └── [81%] main.go

Project coverage: 60%
```

> Note: Your pull branch **must** be aligned with latest change of the base branch
