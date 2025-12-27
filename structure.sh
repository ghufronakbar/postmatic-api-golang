#!/usr/bin/env bash
set -euo pipefail

WITH_FILE=false
ROOT="."
OUT="Structure.md"
MAX_BYTES=200000 # 200KB per file (default)
SHOW_FILES_IN_TREE=false # by default, without --withfile we show directories only

usage() {
  cat <<'EOF'
Usage:
  ./structure.sh [--withfile] [--root <path>] [--out <file>] [--max-bytes <n>]

Options:
  --withfile         Include file list + file contents in output markdown.
  --root <path>      Project root (default: .)
  --out <file>       Output markdown file (default: Structure.md)
  --max-bytes <n>    Max bytes per file to inline (default: 200000)
  -h, --help         Show help

Examples:
  ./structure.sh
  ./structure.sh --withfile
  ./structure.sh --root . --out docs/Structure.md
  ./structure.sh --withfile --max-bytes 500000
EOF
}

# --- arg parsing ---
while [[ $# -gt 0 ]]; do
  case "$1" in
    --withfile)
      WITH_FILE=true
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
      MAX_BYTES="${2:-}"
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

  # skip output itself + this script (optional)
  [[ "$p" == "$OUT" || "$b" == "Structure.md" ]] && return 0

  # skip common sensitive / secret-ish files
  case "$b" in
    .env|.env.*) return 0 ;;
  esac
  case "$p" in
    *.pem|*.key|*.p12|*.pfx|*.crt|*.cer|*.der) return 0 ;;
    *id_rsa*|*id_ed25519*|*secrets*|*secret*|*credentials*|*credential*) return 0 ;;
  esac

  return 1
}

file_size_bytes() {
  # macOS: stat -f%z ; Linux: stat -c%s
  local p="$1"
  if stat -f%z "$p" >/dev/null 2>&1; then
    stat -f%z "$p"
  else
    stat -c%s "$p"
  fi
}

is_text_file() {
  # Use file(1) mime type detection; fallback by extension if needed.
  local p="$1"
  local mime
  mime="$(file -b --mime-type "$p" 2>/dev/null || true)"
  if [[ "$mime" == text/* ]]; then
    return 0
  fi
  case "$mime" in
    application/json|application/xml|application/x-yaml|application/javascript) return 0 ;;
  esac

  # fallback: extensions likely text
  case "$p" in
    *.go|*.mod|*.sum|*.sql|*.md|*.txt|*.yaml|*.yml|*.json|*.toml|*.ini|*.env|*.sh|*.bash|*.zsh|*.dockerfile|*Dockerfile|Makefile|*.mk|*.proto|*.graphql|*.html|*.css|*.js|*.ts|*.tsx|*.jsx) return 0 ;;
  esac

  return 1
}

lang_hint() {
  local p="$1"
  case "$p" in
    *.go) echo "go" ;;
    *.sql) echo "sql" ;;
    *.md) echo "md" ;;
    *.yaml|*.yml) echo "yaml" ;;
    *.json) echo "json" ;;
    *.sh|*.bash|*.zsh) echo "bash" ;;
    *.js) echo "javascript" ;;
    *.ts) echo "typescript" ;;
    *.tsx) echo "tsx" ;;
    *.jsx) echo "jsx" ;;
    *.html) echo "html" ;;
    *.css) echo "css" ;;
    *.proto) echo "proto" ;;
    *) echo "" ;;
  esac
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
# We'll build a sorted list of "rel/" for dirs and "rel" for files.
entries_tmp="$(mktemp)"
files_tmp="$(mktemp)"
dirs_tmp="$(mktemp)"
trap 'rm -f "$entries_tmp" "$files_tmp" "$dirs_tmp"' EXIT

# Find directories (excluding pruned ones)
while IFS= read -r d; do
  # skip excluded dirs
  if is_excluded_dir "$d"; then
    continue
  fi
  echo "$d" >> "$dirs_tmp"
done < <(find "$ROOT" -type d 2>/dev/null)

# Find files (excluding under excluded dirs)
while IFS= read -r f; do
  # skip if any excluded dir is in its path
  if is_excluded_dir "$f"; then
    continue
  fi
  echo "$f" >> "$files_tmp"
done < <(find "$ROOT" -type f 2>/dev/null)

# Build entries list for tree
# Include dirs (except root)
while IFS= read -r d; do
  [[ "$d" == "$ROOT" ]] && continue
  rel="$(relpath "$d")"
  echo "${rel}/" >> "$entries_tmp"
done < "$dirs_tmp"

# Include files in tree only if WITH_FILE (or SHOW_FILES_IN_TREE)
if $SHOW_FILES_IN_TREE; then
  while IFS= read -r f; do
    rel="$(relpath "$f")"
    [[ "$rel" == "." ]] && continue
    # avoid excluded file names in tree too
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
    # indentation based on rel (strip trailing slash for depth calc)
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

  if $WITH_FILE; then
    echo "## File contents"
    echo
    echo "> Catatan: file sensitif (.env, *.key, *.pem, dll) dan file besar (>${MAX_BYTES} bytes) akan di-skip otomatis."
    echo

    # sort files for deterministic output
    LC_ALL=C sort -u "$files_tmp" -o "$files_tmp"

    while IFS= read -r f; do
      # skip excluded dirs already handled; now check file exclude
      rel="$(relpath "$f")"
      if is_excluded_file "$rel"; then
        continue
      fi

      # size limit
      sz="$(file_size_bytes "$f" || echo 0)"
      if [[ "$sz" -gt "$MAX_BYTES" ]]; then
        echo "### \`$rel\`"
        echo
        echo "_Skipped (too large: ${sz} bytes > ${MAX_BYTES})_"
        echo
        continue
      fi

      # text check
      if ! is_text_file "$f"; then
        echo "### \`$rel\`"
        echo
        echo "_Skipped (non-text/binary)_"
        echo
        continue
      fi

      lh="$(lang_hint "$f")"
      echo "### \`$rel\`"
      echo
      if [[ -n "$lh" ]]; then
        echo "\`\`\`$lh"
      else
        echo "\`\`\`"
      fi
      cat "$f"
      echo
      echo "\`\`\`"
      echo
    done < "$files_tmp"
  fi
} > "$OUT"

echo "✅ Generated: $OUT"
echo "   - tree: directories$( $WITH_FILE && echo " + files" || echo "" )"
echo "   - contents: $( $WITH_FILE && echo "included (with limits)" || echo "not included" )"
