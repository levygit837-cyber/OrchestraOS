#!/usr/bin/env sh
set -eu

repo_root="$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd)"

export GOPATH="${ORCHESTRAOS_GOPATH:-$repo_root/.go}"
export GOMODCACHE="${ORCHESTRAOS_GOMODCACHE:-$GOPATH/pkg/mod}"
export GOCACHE="${ORCHESTRAOS_GOCACHE:-$repo_root/.go/cache}"
export GOTOOLCHAIN="${GOTOOLCHAIN:-local}"

if [ -x "$repo_root/.tools/go/bin/go" ]; then
  export GOROOT="$repo_root/.tools/go"
  export PATH="$GOROOT/bin:$PATH"
elif [ -x "/usr/local/go/bin/go" ]; then
  export PATH="/usr/local/go/bin:$PATH"
fi

mkdir -p "$GOPATH" "$GOMODCACHE" "$GOCACHE"
