#!/usr/bin/env sh
set -eu

SRC="${1:-.env}"
DST="${2:-.env.example}"

# Tag yang value-nya tetap dipertahankan
KEEP_TAGS="${KEEP_TAGS:-COMMON DOMAIN ROUTE TIME}"

awk -v keep_tags="$KEEP_TAGS" '
BEGIN{
  n=split(keep_tags, a, /[[:space:]]+/)
  for(i=1;i<=n;i++) if(a[i]!="") keep[toupper(a[i])]=1
  current=""
}

function set_tag(line){
  sub(/^[[:space:]]*#[[:space:]]*/, "", line)

  # dukung "# TAG <NAME>"
  if (toupper(substr(line,1,3))=="TAG") {
    sub(/^[Tt][Aa][Gg][[:space:]]+/, "", line)
  }

  # ambil token pertama sebagai tag
  if (match(line, /^[A-Za-z0-9_-]+/)) {
    current = toupper(substr(line, RSTART, RLENGTH))
  }
}

# comment line (bisa jadi header tag)
 /^[[:space:]]*#/ {
  set_tag($0)
  print $0
  next
}

# blank line
/^[[:space:]]*$/ { print $0; next }

{
  line=$0

  # baris KEY=... (optional "export ")
  if (match(line, /^[[:space:]]*(export[[:space:]]+)?[A-Za-z_][A-Za-z0-9_]*=/)) {
    # kalau tag termasuk yang dikecualikan, pertahankan value
    if (keep[current]) { print line; next }

    eq = index(line, "=")
    lhs = substr(line, 1, eq)      # sampai "="
    rest = substr(line, eq+1)

    # pertahankan inline comment kalau formatnya " ... # comment"
    comment = ""
    if (match(rest, /[[:space:]]#/)) {
      comment = substr(rest, RSTART)
    }

    print lhs comment
    next
  }

  # selain itu, biarkan apa adanya
  print line
}
' "$SRC" > "$DST"

echo "OK: $SRC -> $DST (keep tags: $KEEP_TAGS)"
