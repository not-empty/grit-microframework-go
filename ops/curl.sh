#!/bin/bash

set -e

BASE_URL="http://localhost:8001"
AUTH_CONTEXT="general"
AUTH_TOKEN="adm"
AUTH_SECRET="d22337700548a5aa91adbb353e8bcb9968e112c8b03c2077bb94228ec5954245fe7459a1bf39d4e3b9b90a8d60efd2a4c1875755d51ee74175d639579c026fb7"
DOMAIN="example"
DATA='{
  "name": "Example Name",
  "age": 22
}'
EDIT_DATA='{
  "name": "New Edited Name",
  "age": 99,
  "last_login": "2025-04-28 23:45:12"
}'
NOT_RAW='{
  "query": "unknown"
}'
COUNT_RAW='{
  "query": "count"
}'


function print_custom_headers() {
  local headers_file="$1"
  echo "🔐 Custom headers:"
  grep -iE '^x-(token|expires|request-id|page-cursor):' "$headers_file" | tr -d '\r'
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

# 2. Create a new data
echo "➕ Creating data..."
HEADERS_FILE=$(mktemp)
ADD_RESPONSE=$(curl -s -D "$HEADERS_FILE" -X POST "$BASE_URL/$DOMAIN/add" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Context: $AUTH_CONTEXT" \
  -H "Content-Type: application/json" \
  -d "$DATA") 

print_custom_headers "$HEADERS_FILE"
rm -f "$HEADERS_FILE"

example_ID=$(echo "$ADD_RESPONSE" | jq -r '.id')
if [[ "$example_ID" == "null" || -z "$example_ID" ]]; then
  echo "❌ Failed to get example ID from add response"
  echo "$ADD_RESPONSE"
  exit 1
fi
echo "✅ example created with ID: $example_ID"

# 3. Create a new data 2
echo "➕ Creating data..."
HEADERS_FILE=$(mktemp)
ADD_RESPONSE_2=$(curl -s -D "$HEADERS_FILE" -X POST "$BASE_URL/$DOMAIN/add" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Context: $AUTH_CONTEXT" \
  -H "Content-Type: application/json" \
  -d "$DATA")

print_custom_headers "$HEADERS_FILE"
rm -f "$HEADERS_FILE"

example_ID_2=$(echo "$ADD_RESPONSE_2" | jq -r '.id')
if [[ "$example_ID" == "null" || -z "$example_ID" ]]; then
  echo "❌ Failed to get data ID from add response"
  echo "$ADD_RESPONSE"
  exit 1
fi
echo "✅ data created with ID: $example_ID"


# 4. Get initial detail
echo "🔎 Fetching data detail (before edit)..."
HEADERS_FILE=$(mktemp)
curl -s -D "$HEADERS_FILE" -X GET "$BASE_URL/$DOMAIN/detail/$example_ID" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Context: $AUTH_CONTEXT" \
  -H "Accept: application/json" | jq
print_custom_headers "$HEADERS_FILE"
rm -f "$HEADERS_FILE"

# 4. Edit the data
echo "✏️ Updating data $example_ID..."
HEADERS_FILE=$(mktemp)
curl -s -D "$HEADERS_FILE" -X PATCH "$BASE_URL/$DOMAIN/edit/$example_ID" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Context: $AUTH_CONTEXT" \
  -H "Content-Type: application/json" \
  -d "$EDIT_DATA" | jq
print_custom_headers "$HEADERS_FILE"
rm -f "$HEADERS_FILE"

# 5. Get updated detail
echo "🔎 Fetching data detail (after edit)..."
HEADERS_FILE=$(mktemp)
curl -s -D "$HEADERS_FILE" -X GET "$BASE_URL/$DOMAIN/detail/$example_ID" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Context: $AUTH_CONTEXT" \
  -H "Accept: application/json" | jq
print_custom_headers "$HEADERS_FILE"
rm -f "$HEADERS_FILE"

# 6. Delete the data
echo "❌ Deleting data $example_ID..."
HEADERS_FILE=$(mktemp)
curl -s -D "$HEADERS_FILE" -X DELETE "$BASE_URL/$DOMAIN/delete/$example_ID" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Context: $AUTH_CONTEXT" \
  -H "Accept: application/json" | jq
echo -e "\n🗑️  data deleted."
print_custom_headers "$HEADERS_FILE"
rm -f "$HEADERS_FILE"

# 7. Get dead detail
echo "🕵️ Getting /$DOMAIN/dead_detail/$example_ID..."
HEADERS_FILE=$(mktemp)
curl -s -D "$HEADERS_FILE" -X GET "$BASE_URL/$DOMAIN/dead_detail/$example_ID" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Context: $AUTH_CONTEXT" \
  -H "Accept: application/json" | jq
print_custom_headers "$HEADERS_FILE"
rm -f "$HEADERS_FILE"

# 8. List deleted data
echo "📋 Listing /$DOMAIN/dead_list..."
HEADERS_FILE=$(mktemp)
curl -s -D "$HEADERS_FILE" -X GET "$BASE_URL/$DOMAIN/dead_list" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Context: $AUTH_CONTEXT" \
  -H "Accept: application/json" | jq
print_custom_headers "$HEADERS_FILE"
rm -f "$HEADERS_FILE"

# 9. List active data (final)
echo "📥 Listing examples (after delete)..."
HEADERS_FILE=$(mktemp)
curl -s -D "$HEADERS_FILE" -X GET "$BASE_URL/$DOMAIN/list" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Context: $AUTH_CONTEXT" \
  -H "Accept: application/json" | jq
print_custom_headers "$HEADERS_FILE"
rm -f "$HEADERS_FILE"

# # 10. List one active data
echo "📥 Listing one example..."
HEADERS_FILE=$(mktemp)
curl -s -D "$HEADERS_FILE" -X GET "$BASE_URL/$DOMAIN/list_one?filter=age:eql:22" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Context: $AUTH_CONTEXT" \
  -H "Accept: application/json" | jq
print_custom_headers "$HEADERS_FILE"
rm -f "$HEADERS_FILE"

# 11. Bulk fetch by ID
echo "📦 Bulk fetching data $example_ID..."
HEADERS_FILE=$(mktemp)
curl -s -D "$HEADERS_FILE" -X POST "$BASE_URL/$DOMAIN/bulk" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Context: $AUTH_CONTEXT" \
  -H "Content-Type: application/json" \
  -d '{"ids": ["'"$example_ID_2"'"]}' | jq
print_custom_headers "$HEADERS_FILE"
rm -f "$HEADERS_FILE"

# # 12. Not existent raw
echo "📥 Executing non existent raw query..."
HEADERS_FILE=$(mktemp)
curl -s -D "$HEADERS_FILE" -X POST "$BASE_URL/$DOMAIN/select_raw" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Context: $AUTH_CONTEXT" \
  -H "Content-Type: application/json" \
  -d "$NOT_RAW" | jq
print_custom_headers "$HEADERS_FILE"
rm -f "$HEADERS_FILE"

# # 13. Valid raw
echo "📥 Executing existent raw query..."
HEADERS_FILE=$(mktemp)
curl -s -D "$HEADERS_FILE" -X POST "$BASE_URL/$DOMAIN/select_raw" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Context: $AUTH_CONTEXT" \
  -H "Content-Type: application/json" \
  -d "$COUNT_RAW" | jq
print_custom_headers "$HEADERS_FILE"
rm -f "$HEADERS_FILE"
