#!/bin/bash

set -e

BASE_URL="http://localhost:8001"
AUTH_CONTEXT="adm"
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

# 2. Create a new user
echo "➕ Creating user..."
HEADERS_FILE=$(mktemp)
ADD_RESPONSE=$(curl -s -D "$HEADERS_FILE" -X POST "$BASE_URL/user/add" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Context: $AUTH_CONTEXT" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Alice Braga</b>",
    "email": "alice@example.com"
  }')

print_custom_headers "$HEADERS_FILE"
rm -f "$HEADERS_FILE"

USER_ID=$(echo "$ADD_RESPONSE" | jq -r '.id')
if [[ "$USER_ID" == "null" || -z "$USER_ID" ]]; then
  echo "❌ Failed to get user ID from add response"
  echo "$ADD_RESPONSE"
  exit 1
fi
echo "✅ User created with ID: $USER_ID"

# 3. Create a new user 2
echo "➕ Creating user..."
HEADERS_FILE=$(mktemp)
ADD_RESPONSE_2=$(curl -s -D "$HEADERS_FILE" -X POST "$BASE_URL/user/add" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Context: $AUTH_CONTEXT" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "João Castro</b>",
    "email": "alice@example.com"
  }')

print_custom_headers "$HEADERS_FILE"
rm -f "$HEADERS_FILE"

USER_ID_2=$(echo "$ADD_RESPONSE_2" | jq -r '.id')
if [[ "$USER_ID" == "null" || -z "$USER_ID" ]]; then
  echo "❌ Failed to get user ID from add response"
  echo "$ADD_RESPONSE"
  exit 1
fi
echo "✅ User created with ID: $USER_ID"


# 4. Get initial detail
echo "🔎 Fetching user detail (before edit)..."
HEADERS_FILE=$(mktemp)
curl -s -D "$HEADERS_FILE" -X GET "$BASE_URL/user/detail/$USER_ID" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Context: $AUTH_CONTEXT" \
  -H "Accept: application/json" | jq
print_custom_headers "$HEADERS_FILE"
rm -f "$HEADERS_FILE"

# 4. Edit the user
echo "✏️ Updating user $USER_ID..."
HEADERS_FILE=$(mktemp)
curl -s -D "$HEADERS_FILE" -X PATCH "$BASE_URL/user/edit/$USER_ID" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Context: $AUTH_CONTEXT" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Alice Johnson",
    "email": "alice.johnson@example.com"
  }'
print_custom_headers "$HEADERS_FILE"
rm -f "$HEADERS_FILE"

# 5. Get updated detail
echo "🔎 Fetching user detail (after edit)..."
HEADERS_FILE=$(mktemp)
curl -s -D "$HEADERS_FILE" -X GET "$BASE_URL/user/detail/$USER_ID" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Context: $AUTH_CONTEXT" \
  -H "Accept: application/json" | jq
print_custom_headers "$HEADERS_FILE"
rm -f "$HEADERS_FILE"

# 6. Delete the user
echo "❌ Deleting user $USER_ID..."
HEADERS_FILE=$(mktemp)
curl -s -D "$HEADERS_FILE" -X DELETE "$BASE_URL/user/delete/$USER_ID" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Context: $AUTH_CONTEXT" \
  -H "Accept: application/json"
echo -e "\n🗑️  User deleted."
print_custom_headers "$HEADERS_FILE"
rm -f "$HEADERS_FILE"

# 7. Get dead detail
echo "🕵️ Getting /user/dead_detail/$USER_ID..."
HEADERS_FILE=$(mktemp)
curl -s -D "$HEADERS_FILE" -X GET "$BASE_URL/user/dead_detail/$USER_ID" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Context: $AUTH_CONTEXT" \
  -H "Accept: application/json" | jq
print_custom_headers "$HEADERS_FILE"
rm -f "$HEADERS_FILE"

# 8. List deleted users
echo "📋 Listing /user/dead_list..."
HEADERS_FILE=$(mktemp)
curl -s -D "$HEADERS_FILE" -X GET "$BASE_URL/user/dead_list" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Context: $AUTH_CONTEXT" \
  -H "Accept: application/json" | jq
print_custom_headers "$HEADERS_FILE"
rm -f "$HEADERS_FILE"

# 9. List active users (final)
echo "📥 Listing users (after delete)..."
HEADERS_FILE=$(mktemp)
curl -s -D "$HEADERS_FILE" -X GET "$BASE_URL/user/list" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Context: $AUTH_CONTEXT" \
  -H "Accept: application/json" | jq
print_custom_headers "$HEADERS_FILE"
rm -f "$HEADERS_FILE"

# 10. Bulk fetch by ID
echo "📦 Bulk fetching user $USER_ID..."
HEADERS_FILE=$(mktemp)
curl -s -D "$HEADERS_FILE" -X POST "$BASE_URL/user/bulk" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Context: $AUTH_CONTEXT" \
  -H "Content-Type: application/json" \
  -d '{"ids": ["'"$USER_ID_2"'"]}' | jq
print_custom_headers "$HEADERS_FILE"
rm -f "$HEADERS_FILE"