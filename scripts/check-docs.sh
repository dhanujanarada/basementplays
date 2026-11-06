#!/bin/sh
# Check documentation links and Covers paths. POSIX shell only.
set -eu

ROOT=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
DOCS="$ROOT/docs"
fail=0
checked=0

check_path() {
  file=$1
  value=$2
  base=$3

  case "$value" in
    ''|'#'*|http://*|https://*|mailto:*|tel:*)
      return
      ;;
  esac

  clean=${value%%#*}
  clean=${clean%%\?*}

  case "$clean" in
    '') return ;;
    /*) full=$ROOT$clean ;;
    *) full=$(dirname "$file")/$clean ;;
  esac

  if [ "$base" = "root" ]; then
    full=$ROOT/$clean
  fi

  if [ ! -e "$full" ]; then
    printf 'BROKEN PATH: %s -> %s\n' "${file#$ROOT/}" "$value" >&2
    fail=1
  fi
}

check_links() {
  file=$1
  links=$(
    awk '{
      line = $0
      while (match(line, /\]\([^)]*\)/)) {
        print substr(line, RSTART + 2, RLENGTH - 3)
        line = substr(line, RSTART + RLENGTH)
      }
    }' "$file"
  )

  old_ifs=$IFS
  IFS='
'
  for target in $links; do
    target=${target%% *}
    check_path "$file" "$target" file
  done
  IFS=$old_ifs
}

check_covers() {
  file=$1
  paths=$(
    awk '
      /^[[:space:]]*[-*]?[[:space:]]*\*\*Covers:\*\*/ ||
      /^[[:space:]]*Covers:/ {
        line = $0
        while (match(line, /`[^`]*`/)) {
          print substr(line, RSTART + 1, RLENGTH - 2)
          line = substr(line, RSTART + RLENGTH)
        }
      }
    ' "$file"
  )

  old_ifs=$IFS
  IFS='
'
  for path in $paths; do
    check_path "$file" "$path" root
  done
  IFS=$old_ifs
}

if [ ! -d "$DOCS" ]; then
  echo "docs directory not found" >&2
  exit 1
fi

for file in $(find "$DOCS" -type f -name '*.md' | sort); do
  check_links "$file"
  check_covers "$file"
  checked=$((checked + 1))
done

if [ "$fail" -ne 0 ]; then
  printf 'Documentation check failed.\n' >&2
  exit 1
fi

printf 'Documentation check passed: %s files.\n' "$checked"
