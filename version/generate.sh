#!bash

printf '//nolint\npackage version\nconst Version = "%s"' "${VERSION}" > generated.go
