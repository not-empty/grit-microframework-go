#!/bin/bash

set -e

BASE_URL="http://localhost:8001"
AUTH_CONTEXT="general"
AUTH_TOKEN="adm"
AUTH_SECRET="d22337700548a5aa91adbb353e8bcb9968e112c8b03c2077bb94228ec5954245fe7459a1bf39d4e3b9b90a8d60efd2a4c1875755d51ee74175d639579c026fb7"

function print_custom_headers() {
  local headers_file="$1"
  echo "🔐 Custom headers:"
  grep -iE '^x-(token|expires|request-id):' "$headers_file" | tr -d '\r'
  echo
}

# 1. Get JWT Token from headers
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

# 2. Create a new example
echo "➕ Creating example..."
HEADERS_FILE=$(mktemp)
ADD_RESPONSE=$(curl -s -D "$HEADERS_FILE" -X POST "$BASE_URL/example/add" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Context: $AUTH_CONTEXT" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Alice Braga<script>alert(1)</script>",
    "age": 22
  }')

print_custom_headers "$HEADERS_FILE"
rm -f "$HEADERS_FILE"

example_ID=$(echo "$ADD_RESPONSE" | jq -r '.id')
if [[ "$example_ID" == "null" || -z "$example_ID" ]]; then
  echo "❌ Failed to get example ID from add response"
  echo "$ADD_RESPONSE"
  exit 1
fi
echo "✅ example created with ID: $example_ID"

# 3. Create a new example 2
echo "➕ Creating example..."
HEADERS_FILE=$(mktemp)
ADD_RESPONSE_2=$(curl -s -D "$HEADERS_FILE" -X POST "$BASE_URL/example/add" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Context: $AUTH_CONTEXT" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "João Castro<script>alert(1)</script>",
    "age": 19
  }')

print_custom_headers "$HEADERS_FILE"
rm -f "$HEADERS_FILE"

example_ID_2=$(echo "$ADD_RESPONSE_2" | jq -r '.id')
if [[ "$example_ID" == "null" || -z "$example_ID" ]]; then
  echo "❌ Failed to get example ID from add response"
  echo "$ADD_RESPONSE"
  exit 1
fi
echo "✅ example created with ID: $example_ID"


# 4. Get initial detail
echo "🔎 Fetching example detail (before edit)..."
HEADERS_FILE=$(mktemp)
curl -s -D "$HEADERS_FILE" -X GET "$BASE_URL/example/detail/$example_ID" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Context: $AUTH_CONTEXT" \
  -H "Accept: application/json" | jq
print_custom_headers "$HEADERS_FILE"
rm -f "$HEADERS_FILE"

# 4. Edit the example
echo "✏️ Updating example $example_ID..."
HEADERS_FILE=$(mktemp)
curl -s -D "$HEADERS_FILE" -X PATCH "$BASE_URL/example/edit/$example_ID" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Context: $AUTH_CONTEXT" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Alice Johnson",
    "age": 44
  }'
print_custom_headers "$HEADERS_FILE"
rm -f "$HEADERS_FILE"

# 5. Get updated detail
echo "🔎 Fetching example detail (after edit)..."
HEADERS_FILE=$(mktemp)
curl -s -D "$HEADERS_FILE" -X GET "$BASE_URL/example/detail/$example_ID" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Context: $AUTH_CONTEXT" \
  -H "Accept: application/json" | jq
print_custom_headers "$HEADERS_FILE"
rm -f "$HEADERS_FILE"

# 6. Delete the example
echo "❌ Deleting example $example_ID..."
HEADERS_FILE=$(mktemp)
curl -s -D "$HEADERS_FILE" -X DELETE "$BASE_URL/example/delete/$example_ID" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Context: $AUTH_CONTEXT" \
  -H "Accept: application/json"
echo -e "\n🗑️  example deleted."
print_custom_headers "$HEADERS_FILE"
rm -f "$HEADERS_FILE"

# 7. Get dead detail
echo "🕵️ Getting /example/dead_detail/$example_ID..."
HEADERS_FILE=$(mktemp)
curl -s -D "$HEADERS_FILE" -X GET "$BASE_URL/example/dead_detail/$example_ID" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Context: $AUTH_CONTEXT" \
  -H "Accept: application/json" | jq
print_custom_headers "$HEADERS_FILE"
rm -f "$HEADERS_FILE"

# 8. List deleted examples
echo "📋 Listing /example/dead_list..."
HEADERS_FILE=$(mktemp)
curl -s -D "$HEADERS_FILE" -X GET "$BASE_URL/example/dead_list" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Context: $AUTH_CONTEXT" \
  -H "Accept: application/json" | jq
print_custom_headers "$HEADERS_FILE"
rm -f "$HEADERS_FILE"

# 9. List active examples (final)
echo "📥 Listing examples (after delete)..."
HEADERS_FILE=$(mktemp)
curl -s -D "$HEADERS_FILE" -X GET "$BASE_URL/example/list" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Context: $AUTH_CONTEXT" \
  -H "Accept: application/json" | jq
print_custom_headers "$HEADERS_FILE"
rm -f "$HEADERS_FILE"

# 10. Bulk fetch by ID
echo "📦 Bulk fetching example $example_ID..."
HEADERS_FILE=$(mktemp)
curl -s -D "$HEADERS_FILE" -X POST "$BASE_URL/example/bulk" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Context: $AUTH_CONTEXT" \
  -H "Content-Type: application/json" \
  -d '{"ids": ["'"$example_ID_2"'"]}' | jq
print_custom_headers "$HEADERS_FILE"
rm -f "$HEADERS_FILE"