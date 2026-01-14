#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost:8001}"
AUTH_CONTEXT="${AUTH_CONTEXT:-general}"
AUTH_TOKEN="${AUTH_TOKEN:-adm}"
AUTH_SECRET="${AUTH_SECRET:-d22337700548a5aa91adbb353e8bcb9968e112c8b03c2077bb94228ec5954245fe7459a1bf39d4e3b9b90a8d60efd2a4c1875755d51ee74175d639579c026fb7}"

# Optional:
# LIMIT="${LIMIT:-25}"

print_custom_headers() {
  local headers_file="$1"
  grep -iE '^x-(token|expires|request-id|page-cursor):' "$headers_file" | tr -d '\r' || true
  echo
}

get_token() {
  echo "🔐 Requesting token..."
  local auth_headers
  auth_headers="$(mktemp)"
  curl -s -D "$auth_headers" -o /dev/null -X POST "$BASE_URL/auth/generate" \
    -H "Content-Type: application/json" \
    -H "Context: $AUTH_CONTEXT" \
    -d "{\"token\": \"$AUTH_TOKEN\", \"secret\": \"$AUTH_SECRET\"}"

  print_custom_headers "$auth_headers"
  TOKEN="$(grep -i '^x-token:' "$auth_headers" | awk '{print $2}' | tr -d '\r' || true)"
  EXPIRES="$(grep -i '^x-expires:' "$auth_headers" | awk '{print $2}' | tr -d '\r' || true)"
  rm -f "$auth_headers"

  if [[ -z "${TOKEN:-}" ]]; then
    echo "❌ Failed to retrieve token"
    exit 1
  fi
  echo "✅ Token acquired. Expires: ${EXPIRES:-<none>}"
  echo
}

# Fetch all pages for a given query and write ALL ids to a file (one per line).
# Also prints progress and verifies we never repeat an ID across pages (basic sanity).
fetch_all_ids_for_query() {
  local label="$1"
  local query="$2"
  local out_ids_file="$3"

  : > "$out_ids_file"

  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
  echo "🧪 CASE: $label"
  echo "    query: $query"
  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

  local page_cursor=""
  local pages=0
  local total_ids=0

  while :; do
    local url="$BASE_URL/example/list?$query"
    if [[ -n "$page_cursor" ]]; then
      url+="&page_cursor=$page_cursor"
    fi

    pages=$((pages + 1))
    echo "➡️  GET $url"

    local headers body
    headers="$(mktemp)"
    body="$(mktemp)"

    curl -s -D "$headers" -o "$body" -X GET "$url" \
      -H "Authorization: Bearer $TOKEN" \
      -H "Context: $AUTH_CONTEXT" \
      -H "Accept: application/json"

    # Print a compact view (avoid huge dumps)
    if jq -e 'type=="array"' "$body" >/dev/null 2>&1; then
      local len
      len="$(jq 'length' "$body")"
      echo "📦 items in page: $len"
      # Append IDs
      jq -r '.[].id // empty' "$body" >> "$out_ids_file"
      total_ids=$((total_ids + len))
    else
      echo "❌ Response is not an array:"
      cat "$body"
      rm -f "$headers" "$body"
      exit 1
    fi

    echo
    echo "🔑 Response headers:"
    print_custom_headers "$headers"

    local next_cursor
    next_cursor="$(grep -i '^x-page-cursor:' "$headers" | awk '{print $2}' | tr -d '\r' || true)"

    rm -f "$headers" "$body"

    if [[ -z "$next_cursor" ]]; then
      echo "⚑ done: no X-Page-Cursor returned"
      break
    fi

    page_cursor="$next_cursor"
    echo "🔄 next cursor: $page_cursor"
    echo
  done

  # Basic sanity: detect duplicates (means pagination may have repeated rows)
  local tmp_sorted tmp_uniq
  tmp_sorted="$(mktemp)"
  tmp_uniq="$(mktemp)"
  sort "$out_ids_file" > "$tmp_sorted"
  uniq "$tmp_sorted" > "$tmp_uniq"

  local sorted_count uniq_count
  sorted_count="$(wc -l < "$tmp_sorted" | tr -d ' ')"
  uniq_count="$(wc -l < "$tmp_uniq" | tr -d ' ')"

  if [[ "$sorted_count" -ne "$uniq_count" ]]; then
    echo "❌ Duplicate IDs detected across pages for: $label"
    echo "   sorted_count=$sorted_count uniq_count=$uniq_count"
    echo "   showing first duplicates:"
    uniq -d "$tmp_sorted" | head -n 50
    rm -f "$tmp_sorted" "$tmp_uniq"
    exit 1
  fi

  rm -f "$tmp_sorted" "$tmp_uniq"

  echo "✅ CASE DONE: $label (pages=$pages, total_ids=$uniq_count)"
  echo
}

# Compare two newline-separated ID files for exact equality (as sets).
compare_id_files_as_set() {
  local label_a="$1"
  local file_a="$2"
  local label_b="$3"
  local file_b="$4"

  local count_a count_b
  count_a="$(wc -l < "$file_a" | tr -d ' ')"
  count_b="$(wc -l < "$file_b" | tr -d ' ')"

  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
  echo "🔎 COMPARISON"
  echo "  A: $label_a (count=$count_a)"
  echo "  B: $label_b (count=$count_b)"
  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

  if [[ "$count_a" -ne "$count_b" ]]; then
    echo "❌ Counts differ: $count_a vs $count_b"
    echo "   (This can happen if pagination is incorrect for one of the sorts.)"
    return 1
  fi

  local sa sb
  sa="$(mktemp)"
  sb="$(mktemp)"

  sort "$file_a" > "$sa"
  sort "$file_b" > "$sb"

  if ! diff -q "$sa" "$sb" >/dev/null; then
    echo "❌ ID sets differ."
    echo "   First differences (up to 120 lines):"
    diff "$sa" "$sb" | head -n 120 || true
    rm -f "$sa" "$sb"
    return 1
  fi

  rm -f "$sa" "$sb"
  echo "✅ Same count and same ID set."
  echo
  return 0
}

main() {
  command -v jq >/dev/null 2>&1 || { echo "❌ jq is required"; exit 1; }

  get_token

  # We want to verify that changing order_by type (string vs int) does NOT change
  # the set of IDs we can retrieve through pagination.

  local ids_name ids_age
  ids_name="$(mktemp)"
  ids_age="$(mktemp)"

  # CASE A: order_by=name (string)
  # Using fields=age so we can reproduce your scenario where requested fields differ,
  # but the API should still paginate consistently.
  fetch_all_ids_for_query \
    "order_by=name (string) | fields=age" \
    "fields=age&order=asc&order_by=name" \
    "$ids_name"

  # CASE B: order_by=age (int)
  # Ensure your API supports ordering by age; if not allowed, it will fallback to id
  # and the comparison will fail (which is still useful).
  fetch_all_ids_for_query \
    "order_by=age (int) | fields=name" \
    "fields=name&order=asc&order_by=age" \
    "$ids_age"

  # Compare final ID sets
  compare_id_files_as_set \
    "order_by=name (string)" "$ids_name" \
    "order_by=age (int)" "$ids_age"

  rm -f "$ids_name" "$ids_age"

  echo "🎉 Type/order pagination test finished."
}

main "$@"
