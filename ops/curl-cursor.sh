#!/usr/bin/env bash
set -e

BASE_URL="http://localhost:8001"
AUTH_CONTEXT="general"
AUTH_TOKEN="adm"
AUTH_SECRET="d22337700548a5aa91adbb353e8bcb9968e112c8b03c2077bb94228ec5954245fe7459a1bf39d4e3b9b90a8d60efd2a4c1875755d51ee74175d639579c026fb7"

function print_custom_headers() {
  local headers_file="$1"
  grep -iE '^x-(token|expires|request-id|page-cursor):' "$headers_file" | tr -d '\r'
  echo
}

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

if [[ -z "$TOKEN" ]]; then
  echo "❌ Failed to retrieve token"
  exit 1
fi
echo "✅ Token acquired. Expires: $EXPIRES"

# ──────────────────────────────────────────────────────────────────────────────
# 📥 Loop through /example/list pages until we get back an empty JSON array
# ──────────────────────────────────────────────────────────────────────────────

PAGE_CURSOR=""
while :; do
  # build URL with optional page_cursor param
  URL="$BASE_URL/example/list?order=asc&order_by=name"
  if [[ -n "$PAGE_CURSOR" ]]; then
    URL+="&page_cursor=$PAGE_CURSOR"
  fi

  echo "➡️  GET $URL"
  HEADERS=$(mktemp)
  BODY=$(mktemp)
  curl -s -D "$HEADERS" -o "$BODY" -X GET "$URL" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Context: $AUTH_CONTEXT" \
    -H "Accept: application/json"

  # pretty‐print the JSON body
  jq . "$BODY" || cat "$BODY"
  echo

  # show our important headers
  echo "🔑 Response headers:"
  print_custom_headers "$HEADERS"

  # grab the next cursor
  NEXT_CURSOR=$(grep -i '^x-page-cursor:' "$HEADERS" | awk '{print $2}' | tr -d '\r')
  rm -f "$HEADERS"

  # stop if there's no cursor header, or the body is an empty array
  if [[ -z "$NEXT_CURSOR" ]]; then
    echo "⚑ no more pages (no X-Page-Cursor returned)"
    rm -f "$BODY"
    break
  fi
  if jq -e 'type=="array" and length==0' "$BODY" >/dev/null; then
    echo "⚑ no more pages (empty array)"
    rm -f "$BODY"
    break
  fi

  # otherwise, continue with the new cursor
  PAGE_CURSOR="$NEXT_CURSOR"
  echo "🔄 next cursor: $PAGE_CURSOR"
  rm -f "$BODY"
  echo
done
