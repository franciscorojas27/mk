# mk

`mk` is a small Go CLI that creates files and directories from a path.

## Usage

```bash
mk path/to/file.txt
```

Preview what would happen without writing anything:

```bash
mk -n src/{api,cmd}/main.go
```

It also supports simple brace expansion:

```bash
mk "src/{api,cmd}/main.go"
```
