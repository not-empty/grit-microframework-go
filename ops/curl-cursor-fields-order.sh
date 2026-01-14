#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost:8001}"
AUTH_CONTEXT="${AUTH_CONTEXT:-general}"
AUTH_TOKEN="${AUTH_TOKEN:-adm}"
AUTH_SECRET="${AUTH_SECRET:-d22337700548a5aa91adbb353e8bcb9968e112c8b03c2077bb94228ec5954245fe7459a1bf39d4e3b9b90a8d60efd2a4c1875755d51ee74175d639579c026fb7}"

MAX_PAGES="${MAX_PAGES:-200}" # safety guard against infinite loops

function print_custom_headers() {
  local headers_file="$1"
  grep -iE '^x-(token|expires|request-id|page-cursor):' "$headers_file" | tr -d '\r' || true
  echo
}

function require_jq() {
  command -v jq >/dev/null 2>&1 || {
    echo "❌ jq is required for this script."
    exit 1
  }
}

function get_token() {
  echo "🔐 Requesting token..."
  local auth_headers
  auth_headers=$(mktemp)

  curl -s -D "$auth_headers" -o /dev/null -X POST "$BASE_URL/auth/generate" \
    -H "Content-Type: application/json" \
    -H "Context: $AUTH_CONTEXT" \
    -d "{\"token\": \"$AUTH_TOKEN\", \"secret\": \"$AUTH_SECRET\"}"

  print_custom_headers "$auth_headers"

  local token
  token=$(grep -i '^x-token:' "$auth_headers" | awk '{print $2}' | tr -d '\r' || true)
  local expires
  expires=$(grep -i '^x-expires:' "$auth_headers" | awk '{print $2}' | tr -d '\r' || true)

  rm -f "$auth_headers"

  if [[ -z "${token:-}" ]]; then
    echo "❌ Failed to retrieve token"
    exit 1
  fi

  echo "✅ Token acquired. Expires: $expires"
  echo "$token"
}

# Assertions:
# - body is array
# - if array non-empty, all objects contain "id"
# - if required_field provided, all objects contain it too
function assert_page_body() {
  local body_file="$1"
  local required_field="${2:-}"

  # must be JSON array
  jq -e 'type=="array"' "$body_file" >/dev/null || {
    echo "❌ Response is not a JSON array:"
    cat "$body_file"
    exit 1
  }

  # if empty array, ok (end condition)
  if jq -e 'length==0' "$body_file" >/dev/null; then
    return 0
  fi

  # ensure every item has id
  jq -e 'all(.[]; has("id"))' "$body_file" >/dev/null || {
    echo "❌ Assertion failed: not all items contain \"id\""
    jq . "$body_file" || cat "$body_file"
    exit 1
  }

  # ensure every item has required_field (when provided)
  if [[ -n "$required_field" ]]; then
    jq -e --arg f "$required_field" 'all(.[]; has($f))' "$body_file" >/dev/null || {
      echo "❌ Assertion failed: not all items contain \"$required_field\""
      jq . "$body_file" || cat "$body_file"
      exit 1
    }
  fi
}

function run_case() {
  local label="$1"
  local base_query="$2"          # everything after ? excluding page_cursor
  local order_by_for_assert="$3" # empty or field name to assert present when non-empty pages

  echo
  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
  echo "🧪 CASE: $label"
  echo "    query: $base_query"
  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

  local page_cursor=""
  local seen_cursors=""
  local i=0

  while :; do
    i=$((i+1))
    if (( i > MAX_PAGES )); then
      echo "❌ Hit MAX_PAGES=$MAX_PAGES (possible infinite pagination loop)"
      exit 1
    fi

    local url="$BASE_URL/example/list?$base_query"
    if [[ -n "$page_cursor" ]]; then
      url="${url}&page_cursor=${page_cursor}"
    fi

    echo "➡️  GET $url"
    local headers body
    headers=$(mktemp)
    body=$(mktemp)

    curl -s -D "$headers" -o "$body" -X GET "$url" \
      -H "Authorization: Bearer $TOKEN" \
      -H "Context: $AUTH_CONTEXT" \
      -H "Accept: application/json"

    # Pretty print
    jq . "$body" || cat "$body"
    echo

    echo "🔑 Response headers:"
    print_custom_headers "$headers"

    # Assertions
    # If order_by is set and not id, ensure it exists in each item
    local required_field=""
    if [[ -n "$order_by_for_assert" && "$order_by_for_assert" != "id" ]]; then
      required_field="$order_by_for_assert"
    fi
    assert_page_body "$body" "$required_field"

    # next cursor
    local next_cursor
    next_cursor=$(grep -i '^x-page-cursor:' "$headers" | awk '{print $2}' | tr -d '\r' || true)

    rm -f "$headers"

    # stop conditions: no cursor OR empty array
    if [[ -z "${next_cursor:-}" ]]; then
      echo "⚑ done: no X-Page-Cursor returned"
      rm -f "$body"
      break
    fi

    if jq -e 'length==0' "$body" >/dev/null; then
      echo "⚑ done: empty array"
      rm -f "$body"
      break
    fi

    # cursor must advance (avoid loops)
    if [[ "$seen_cursors" == *"|$next_cursor|"* ]]; then
      echo "❌ Cursor repeated (pagination loop detected): $next_cursor"
      rm -f "$body"
      exit 1
    fi
    seen_cursors="${seen_cursors}|${next_cursor}|"

    page_cursor="$next_cursor"
    echo "🔄 next cursor: $page_cursor"
    rm -f "$body"
    echo
  done

  echo "✅ CASE PASSED: $label"
}

# ──────────────────────────────────────────────────────────────────────────────
# Main
# ──────────────────────────────────────────────────────────────────────────────

require_jq
TOKEN="$(get_token)"

# 1) fields + NO order (no id in fields)
run_case \
  "fields + no order (no id in fields)" \
  "fields=name" \
  ""

# 2) fields + NO order (id included in fields)
run_case \
  "fields + no order (id included)" \
  "fields=id,name" \
  ""

# 3) fields + order (neither id nor order_by field included)
run_case \
  "fields + order (missing id and order_by field)" \
  "fields=age&order=asc&order_by=name" \
  "name"

# 4) fields + order (id and order_by field included)
run_case \
  "fields + order (id + order_by field included)" \
  "fields=id,name,age&order=asc&order_by=name" \
  "name"

echo
echo "🎉 All pagination field/cursor cases passed."
