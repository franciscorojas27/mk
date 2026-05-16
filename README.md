# mk

`mk` is a small Go CLI that creates files and directories from a path.

## Usage

```bash
mk path/to/file.txt
```

It also supports simple brace expansion:

```bash
mk "src/{api,cmd}/main.go"
```

## Releases

Tagged pushes like `v1.0.0` trigger GitHub Actions to build Windows and Linux binaries and publish them as a GitHub Release.
