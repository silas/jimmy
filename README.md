# Jimmy

Jimmy is a Google Spanner schema migrations tool.

## Usage

```
Usage:
  jimmy [command]

Available Commands:
  init        Initialize migrations
  create      Create a new migration
  add         Add to an existing migration
  upgrade     Run all schema upgrades
  templates   Show templates
  help        Help about any command

Flags:
  -c, --config string     configuration file (default ".jimmy.yaml")
  -d, --database string   set Spanner database ID
      --emulator          set whether to enable emulator mode (default automatically detected)
  -h, --help              help for jimmy
  -i, --instance string   set Spanner instance ID
  -p, --project string    set Google project ID
```
