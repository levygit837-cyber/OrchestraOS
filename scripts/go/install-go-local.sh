#!/usr/bin/env sh
set -eu

version="${GO_VERSION:-1.26.2}"
repo_root="$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd)"
tools_dir="$repo_root/.tools"
archive="$tools_dir/go${version}.linux-amd64.tar.gz"
url="https://go.dev/dl/go${version}.linux-amd64.tar.gz"

mkdir -p "$tools_dir"

if [ -x "$tools_dir/go/bin/go" ]; then
  "$tools_dir/go/bin/go" version
  exit 0
fi

echo "Downloading $url"
if command -v curl >/dev/null 2>&1; then
  curl -fsSL "$url" -o "$archive"
elif command -v wget >/dev/null 2>&1; then
  wget -q "$url" -O "$archive"
else
  echo "curl or wget is required to install Go locally" >&2
  exit 1
fi

rm -rf "$tools_dir/go"
tar -C "$tools_dir" -xzf "$archive"
"$tools_dir/go/bin/go" version
