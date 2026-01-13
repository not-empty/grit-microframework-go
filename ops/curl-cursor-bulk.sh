#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost:8001}"
AUTH_CONTEXT="${AUTH_CONTEXT:-general}"
AUTH_TOKEN="${AUTH_TOKEN:-adm}"
AUTH_SECRET="${AUTH_SECRET:-d22337700548a5aa91adbb353e8bcb9968e112c8b03c2077bb94228ec5954245fe7459a1bf39d4e3b9b90a8d60efd2a4c1875755d51ee74175d639579c026fb7}"

# How many IDs to sample for bulk tests
SAMPLE_SIZE="${SAMPLE_SIZE:-40}"

# If your /example/bulk supports pagination cursor, keep this 1. If not, set 0.
BULK_SUPPORTS_PAGINATION="${BULK_SUPPORTS_PAGINATION:-1}"

# Optional: fields to request (keep it small but includes order_by + id is enforced server-side anyway)
FIELDS_LIST="${FIELDS_LIST:-id}"   # list baseline can use id only
FIELDS_BULK="${FIELDS_BULK:-id}"   # bulk stress can use id only too

function print_custom_headers() {
  local headers_file="$1"
  grep -iE '^x-(token|expires|request-id|page-cursor):' "$headers_file" | tr -d '\r' || true
  echo
}

function require_bin() {
  command -v "$1" >/dev/null 2>&1 || { echo "❌ missing dependency: $1"; exit 1; }
}

require_bin curl
require_bin jq
require_bin shuf
require_bin sort
require_bin comm
require_bin wc
require_bin mktemp

echo "🔐 Requesting token..."
AUTH_HEADERS=$(mktemp)
curl -s -D "$AUTH_HEADERS" -o /dev/null -X POST "$BASE_URL/auth/generate" \
  -H "Content-Type: application/json" \
  -H "Context: $AUTH_CONTEXT" \
  -d "{\"token\": \"$AUTH_TOKEN\", \"secret\": \"$AUTH_SECRET\"}"

print_custom_headers "$AUTH_HEADERS"
TOKEN=$(grep -i '^x-token:' "$AUTH_HEADERS" | awk '{print $2}' | tr -d '\r')
EXPIRES=$(grep -i '^x-expires:' "$AUTH_HEADERS" | awk '{print $2}' | tr -d '\r')
rm -f "$AUTH_HEADERS"

if [[ -z "${TOKEN:-}" ]]; then
  echo "❌ Failed to retrieve token"
  exit 1
fi
echo "✅ Token acquired. Expires: $EXPIRES"
echo

# ------------------------------------------------------------------------------
# Helpers
# ------------------------------------------------------------------------------

# GET list pages and output IDs (one per line)
fetch_list_ids() {
  local order_by="${1:-id}"
  local order="${2:-asc}"

  local page_cursor=""
  while :; do
    local url="$BASE_URL/example/list?fields=$FIELDS_LIST&order=${order}&order_by=${order_by}"
    if [[ -n "$page_cursor" ]]; then
      url+="&page_cursor=$page_cursor"
    fi

    local headers="$(mktemp)"
    local body="$(mktemp)"

    curl -s -D "$headers" -o "$body" -X GET "$url" \
      -H "Authorization: Bearer $TOKEN" \
      -H "Context: $AUTH_CONTEXT" \
      -H "Accept: application/json" >/dev/null

    # Print IDs from this page
    jq -r '.[].id' "$body"

    local next_cursor
    next_cursor="$(grep -i '^x-page-cursor:' "$headers" | awk '{print $2}' | tr -d '\r' || true)"

    rm -f "$headers" "$body"

    if [[ -z "${next_cursor:-}" ]]; then
      break
    fi
    page_cursor="$next_cursor"
  done
}

# POST bulk (single page, no pagination) and output IDs
fetch_bulk_ids_single() {
  local order_by="${1:-}"
  local order="${2:-}"
  local ids_json="${3:?missing ids_json}" # JSON array string: ["id1","id2",...]

  local qs="fields=$FIELDS_BULK"
  if [[ -n "${order_by:-}" ]]; then qs+="&order_by=${order_by}"; fi
  if [[ -n "${order:-}" ]]; then qs+="&order=${order}"; fi

  local url="$BASE_URL/example/bulk?$qs"

  local headers="$(mktemp)"
  local body="$(mktemp)"

  curl -s -D "$headers" -o "$body" -X POST "$url" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Context: $AUTH_CONTEXT" \
    -H "Accept: application/json" \
    -H "Content-Type: application/json" \
    -d "{\"ids\": $ids_json}" >/dev/null

  jq -r '.[].id' "$body"

  rm -f "$headers" "$body"
}

# POST bulk with pagination cursor loop (if supported) and output IDs
fetch_bulk_ids_paged() {
  local order_by="${1:-}"
  local order="${2:-}"
  local ids_json="${3:?missing ids_json}"

  local page_cursor=""
  while :; do
    local qs="fields=$FIELDS_BULK"
    if [[ -n "${order_by:-}" ]]; then qs+="&order_by=${order_by}"; fi
    if [[ -n "${order:-}" ]]; then qs+="&order=${order}"; fi
    if [[ -n "${page_cursor:-}" ]]; then qs+="&page_cursor=${page_cursor}"; fi

    local url="$BASE_URL/example/bulk?$qs"

    local headers="$(mktemp)"
    local body="$(mktemp)"

    curl -s -D "$headers" -o "$body" -X POST "$url" \
      -H "Authorization: Bearer $TOKEN" \
      -H "Context: $AUTH_CONTEXT" \
      -H "Accept: application/json" \
      -H "Content-Type: application/json" \
      -d "{\"ids\": $ids_json}" >/dev/null

    jq -r '.[].id' "$body"

    local next_cursor
    next_cursor="$(grep -i '^x-page-cursor:' "$headers" | awk '{print $2}' | tr -d '\r' || true)"

    rm -f "$headers" "$body"

    if [[ -z "${next_cursor:-}" ]]; then
      break
    fi
    page_cursor="$next_cursor"
  done
}

# Compare expected vs actual ID sets (both files contain one id per line)
compare_id_sets() {
  local expected_file="$1"
  local actual_file="$2"
  local label="$3"

  sort -u "$expected_file" > "${expected_file}.sorted"
  sort -u "$actual_file" > "${actual_file}.sorted"

  local expected_count actual_count
  expected_count="$(wc -l < "${expected_file}.sorted" | tr -d ' ')"
  actual_count="$(wc -l < "${actual_file}.sorted" | tr -d ' ')"

  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
  echo "🧪 $label"
  echo "  expected unique ids: $expected_count"
  echo "  actual   unique ids: $actual_count"

  if [[ "$expected_count" != "$actual_count" ]]; then
    echo "❌ COUNT MISMATCH"
  fi

  # IDs missing from actual
  local missing extra
  missing="$(comm -23 "${expected_file}.sorted" "${actual_file}.sorted" | head -n 20 || true)"
  extra="$(comm -13 "${expected_file}.sorted" "${actual_file}.sorted" | head -n 20 || true)"

  if [[ -n "$missing" ]]; then
    echo "❌ Missing IDs (first 20):"
    echo "$missing"
  fi

  if [[ -n "$extra" ]]; then
    echo "❌ Extra IDs (first 20):"
    echo "$extra"
  fi

  if [[ -z "$missing" && -z "$extra" && "$expected_count" == "$actual_count" ]]; then
    echo "✅ OK: same ID set"
  else
    echo "❌ FAIL: ID sets differ"
    exit 1
  fi

  rm -f "${expected_file}.sorted" "${actual_file}.sorted"
  echo
}

# ------------------------------------------------------------------------------
# 1) Baseline: get ALL ids from list (stable order)
# ------------------------------------------------------------------------------
echo "📥 Fetching baseline IDs from /example/list..."
BASELINE_ALL="$(mktemp)"
# Use deterministic order to reduce flakiness (id asc is safest)
fetch_list_ids "id" "asc" | sort -u > "$BASELINE_ALL"

BASELINE_COUNT="$(wc -l < "$BASELINE_ALL" | tr -d ' ')"
echo "✅ Baseline IDs collected: $BASELINE_COUNT"
echo

if [[ "$BASELINE_COUNT" -lt 1 ]]; then
  echo "❌ No baseline IDs found; aborting."
  exit 1
fi

# ------------------------------------------------------------------------------
# 2) Build a random sample subset for bulk
# ------------------------------------------------------------------------------
echo "🎯 Sampling $SAMPLE_SIZE IDs for bulk tests..."
SAMPLE_IDS_FILE="$(mktemp)"
shuf "$BASELINE_ALL" | head -n "$SAMPLE_SIZE" | sort -u > "$SAMPLE_IDS_FILE"
SAMPLE_COUNT="$(wc -l < "$SAMPLE_IDS_FILE" | tr -d ' ')"
echo "✅ Sample size (unique): $SAMPLE_COUNT"
echo

# Build JSON array
IDS_JSON="$(jq -R -s -c 'split("\n")[:-1]' "$SAMPLE_IDS_FILE")"

# ------------------------------------------------------------------------------
# 3) Bulk fetch variants and compare against expected sample set
# ------------------------------------------------------------------------------
run_bulk_case() {
  local label="$1"
  local order_by="${2:-}"
  local order="${3:-}"

  local actual_file="$(mktemp)"

  if [[ "$BULK_SUPPORTS_PAGINATION" == "1" ]]; then
    fetch_bulk_ids_paged "$order_by" "$order" "$IDS_JSON" > "$actual_file"
  else
    fetch_bulk_ids_single "$order_by" "$order" "$IDS_JSON" > "$actual_file"
  fi

  compare_id_sets "$SAMPLE_IDS_FILE" "$actual_file" "$label"

  rm -f "$actual_file"
}

run_bulk_case "BULK: no order params" "" ""
run_bulk_case "BULK: order_by=name (string) asc" "name" "asc"
run_bulk_case "BULK: order_by=age (int) asc" "age" "asc"

echo "🎉 Bulk vs list subset checks passed."

rm -f "$BASELINE_ALL" "$SAMPLE_IDS_FILE"
