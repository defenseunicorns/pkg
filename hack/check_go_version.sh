#!/usr/bin/env bash

set -euo pipefail

first_version=""
# Find and iterate over all go.mod files, excluding the vendor directory
find . -name go.mod | while read -r mod; do
    current_version=$(grep '^go 1\.' "$mod" | cut -d ' ' -f 2)

    if [[ -z "$first_version" ]]; then
      first_version=$current_version
    elif [[ "$current_version" != "$first_version" ]]; then
      echo "Inconsistency found: $mod uses Go version $current_version, this differs from another found version $first_version."
      exit 1
    fi
done

echo "All modules use the same Go version: $first_version."
