#!/usr/bin/env bash
set -euo pipefail

BASE=${BASE:-http://localhost:8080/api}
EMAIL=${EMAIL:-seeduser@example.com}
PASS=${PASS:-changeme123}
UNIT=${UNIT:-101}

echo "Seeding Cul-de-Chat API at $BASE"

token=""

login() {
  local resp token
  resp=$(curl -sS -X POST -H 'Content-Type: application/json' \
    "$BASE/auth/login" \
    -d "{\"email\":\"$EMAIL\",\"password\":\"$PASS\"}") || resp=""
  token=$(printf '%s' "$resp" | python3 - <<'PY'
import sys, json
try:
    d = json.load(sys.stdin)
    print(d.get('token', ''))
except Exception:
    print('')
PY
)
  printf '%s' "$token"
}

register() {
  curl -sS -X POST -H 'Content-Type: application/json' \
    "$BASE/auth/register" \
    -d "{\"email\":\"$EMAIL\",\"unit_number\":\"$UNIT\",\"password\":\"$PASS\"}" || true
}

echo "Attempting login..."
token=$(login)
if [ -z "$token" ]; then
  echo "Registering user..."
  register >/dev/null || true
  echo "Retrying login..."
  token=$(login)
fi

if [ -z "$token" ]; then
  echo "ERROR: Unable to obtain token" >&2
  exit 1
fi

echo "Token acquired (len=${#token})"

echo "Creating boards (idempotent)..."
curl -sS -X POST -H 'Content-Type: application/json' -H "Authorization: Bearer $token" \
  "$BASE/boards" -d '{"name":"General"}' >/dev/null || true
curl -sS -X POST -H 'Content-Type: application/json' -H "Authorization: Bearer $token" \
  "$BASE/boards" -d '{"name":"For Sale"}' >/dev/null || true

boards=$(curl -sS "$BASE/boards")
gen_id=$(printf '%s' "$boards" | python3 - <<'PY'
import sys, json
try:
    arr = json.load(sys.stdin)
    print(next((b.get('id', '') for b in arr if b.get('name') == 'General'), ''))
except Exception:
    print('')
PY
)

if [ -z "$gen_id" ]; then
  echo "ERROR: Could not resolve General board id" >&2
  exit 1
fi

echo "General board id: $gen_id"

echo "Creating a sample post..."
post_resp=$(curl -sS -X POST -H 'Content-Type: application/json' -H "Authorization: Bearer $token" \
  "$BASE/posts" \
  -d "{\"board_id\":\"$gen_id\",\"title\":\"Welcome to General\",\"content\":\"This is a seed post for smoke testing.\",\"bulletin\":false}")
post_id=$(printf '%s' "$post_resp" | python3 - <<'PY'
import sys, json
try:
    d = json.load(sys.stdin)
    print(d.get('id', ''))
except Exception:
    print('')
PY
)

echo "Post created id: ${post_id:-none}"

echo "Listing posts in General..."
curl -sS "$BASE/posts/board/$gen_id"
echo

echo "Seeding complete."


