#!/usr/bin/env bash
set -euo pipefail

# Default configuration
SHOW_FILES_IN_TREE=false # by default, show directories only
ROOT="."
OUT="Structure.md"

usage() {
  cat <<'EOF'
Usage:
  ./structure.sh [--withfile] [--root <path>] [--out <file>]

Options:
  --withfile         Include filenames in the tree structure (does not print content).
  --root <path>      Project root (default: .)
  --out <file>       Output markdown file (default: Structure.md)
  -h, --help         Show help

Examples:
  ./structure.sh             # Directories only
  ./structure.sh --withfile  # Directories + Files
  ./structure.sh --root . --out docs/Structure.md
EOF
}

# --- arg parsing ---
while [[ $# -gt 0 ]]; do
  case "$1" in
    --withfile)
      SHOW_FILES_IN_TREE=true
      shift
      ;;
    --root)
      ROOT="${2:-}"
      shift 2
      ;;
    --out)
      OUT="${2:-}"
      shift 2
      ;;
    --max-bytes)
      # Kept for compatibility, but ignored
      shift 2
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "Unknown arg: $1" >&2
      usage >&2
      exit 1
      ;;
  esac
done

ROOT="${ROOT%/}"

# --- helpers ---
is_excluded_dir() {
  case "$1" in
    */.git|*/.git/*) return 0 ;;
    */vendor|*/vendor/*) return 0 ;;
    */node_modules|*/node_modules/*) return 0 ;;
    */dist|*/dist/*) return 0 ;;
    */build|*/build/*) return 0 ;;
    */bin|*/bin/*) return 0 ;;
    */tmp|*/tmp/*) return 0 ;;
    */coverage|*/coverage/*) return 0 ;;
    */.idea|*/.idea/*) return 0 ;;
    */.vscode|*/.vscode/*) return 0 ;;
    */.DS_Store) return 0 ;;
    *) return 1 ;;
  esac
}

is_excluded_file() {
  local p="$1"
  local b
  b="$(basename "$p")"

  # skip output itself + this script
  [[ "$p" == "$OUT" || "$b" == "structure.sh" ]] && return 0

  # skip common sensitive files (optional logic for tree view)
  case "$b" in
    .env|.env.*) return 0 ;;
    .DS_Store) return 0 ;;
  esac

  return 1
}

relpath() {
  local p="$1"
  if [[ "$p" == "$ROOT" ]]; then
    echo "."
  else
    echo "${p#"$ROOT"/}"
  fi
}

indent_for_rel() {
  local rel="$1"
  if [[ "$rel" == "." ]]; then
    echo ""
    return
  fi
  # depth = number of segments - 1
  local depth
  depth="$(awk -F'/' '{print NF-1}' <<< "$rel")"
  local spaces=$((depth * 2))
  printf '%*s' "$spaces" ""
}

# --- collect entries ---
entries_tmp="$(mktemp)"
files_tmp="$(mktemp)"
dirs_tmp="$(mktemp)"
trap 'rm -f "$entries_tmp" "$files_tmp" "$dirs_tmp"' EXIT

# Find directories
while IFS= read -r d; do
  if is_excluded_dir "$d"; then
    continue
  fi
  echo "$d" >> "$dirs_tmp"
done < <(find "$ROOT" -type d 2>/dev/null)

# Find files (only if needed for tree)
if $SHOW_FILES_IN_TREE; then
  while IFS= read -r f; do
    if is_excluded_dir "$f"; then
      continue
    fi
    echo "$f" >> "$files_tmp"
  done < <(find "$ROOT" -type f 2>/dev/null)
fi

# Build entries list for tree
# 1. Dirs
while IFS= read -r d; do
  [[ "$d" == "$ROOT" ]] && continue
  rel="$(relpath "$d")"
  echo "${rel}/" >> "$entries_tmp"
done < "$dirs_tmp"

# 2. Files (if enabled)
if $SHOW_FILES_IN_TREE; then
  while IFS= read -r f; do
    rel="$(relpath "$f")"
    [[ "$rel" == "." ]] && continue
    if is_excluded_file "$rel"; then
      continue
    fi
    echo "$rel" >> "$entries_tmp"
  done < "$files_tmp"
fi

LC_ALL=C sort -u "$entries_tmp" -o "$entries_tmp"

# --- write output ---
now="$(date '+%Y-%m-%d %H:%M:%S')"
mkdir -p "$(dirname "$OUT")" 2>/dev/null || true

{
  echo "# Project Structure"
  echo
  echo "_Generated: ${now}_"
  echo
  echo "Root: \`$ROOT\`"
  echo
  echo '```text'
  echo "."
  while IFS= read -r e; do
    # e is like "internal/" or "internal/app.go"
    rel="$e"
    # indentation based on rel
    rel_no_slash="${rel%/}"
    ind="$(indent_for_rel "$rel_no_slash")"
    name="$(basename "$rel_no_slash")"
    if [[ "$rel" == */ ]]; then
      echo "${ind}- ${name}/"
    else
      echo "${ind}- ${name}"
    fi
  done < "$entries_tmp"
  echo '```'
  echo
} > "$OUT"

echo "✅ Generated: $OUT"
echo "   - tree: directories$( $SHOW_FILES_IN_TREE && echo " + files" || echo "" )"
echo "   - contents: excluded (structure only)"