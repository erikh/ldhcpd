#!bash

printf 'package version\nconst Version = "%s"' "${VERSION}" > generated.go
