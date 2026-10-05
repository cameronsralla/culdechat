#!/usr/bin/env bash
set -euo pipefail

BASE=${BASE:-http://localhost:8080/api}
ADMIN_EMAIL=${ADMIN_EMAIL:-admin@culdechat.local}
ADMIN_PASS=${ADMIN_PASS:-changeme123}
SEED_PASS=${SEED_PASS:-changeme123}

echo "Seeding Cul-de-Chat API at $BASE"

json_field() {
  python3 -c 'import sys, json
raw = sys.stdin.read().strip()
if not raw:
    raise SystemExit
try:
    d = json.loads(raw)
except Exception:
    raise SystemExit
print(d.get(sys.argv[1], "") or "")' "$1"
}

login() {
  local email="$1" password="$2"
  local resp
  resp=$(curl -sS -X POST -H 'Content-Type: application/json' \
    "$BASE/auth/login" \
    -d "{\"email\":\"$email\",\"password\":\"$password\"}") || resp=""
  printf '%s' "$resp" | json_field token
}

user_exists() {
  local email="$1"
  printf '%s' "$roster" | python3 -c '
import sys, json
email = sys.argv[1]
try:
    rows = json.load(sys.stdin)
except Exception:
    rows = []
print("yes" if any((r or {}).get("email") == email for r in rows) else "")
' "$email"
}

ensure_resident() {
  local email="$1" unit="$2" name="$3" directory="$4"
  local token=""
  if [ -n "$(user_exists "$email")" ]; then
    token=$(login "$email" "$SEED_PASS")
  else
    echo "Inviting $name ($email, unit $unit)..." >&2
    local invite reg_token passcode complete
    invite=$(curl -sS -X POST -H 'Content-Type: application/json' -H "Authorization: Bearer $admin_token" \
      "$BASE/auth/register" \
      -d "{\"email\":\"$email\",\"unit_number\":\"$unit\"}")
    reg_token=$(printf '%s' "$invite" | json_field registration_token)
    passcode=$(printf '%s' "$invite" | json_field passcode)
    if [ -z "$reg_token" ] || [ -z "$passcode" ]; then
      echo "ERROR: invite failed for $email: $invite" >&2
      exit 1
    fi
    complete=$(curl -sS -X POST -H 'Content-Type: application/json' \
      "$BASE/auth/complete-registration" \
      -d "{\"token\":\"$reg_token\",\"passcode\":\"$passcode\",\"password\":\"$SEED_PASS\",\"name\":\"$name\"}")
    token=$(printf '%s' "$complete" | json_field token)
  fi
  if [ -z "$token" ]; then
    echo "ERROR: Unable to obtain a session for $email" >&2
    exit 1
  fi
  curl -sS -X PATCH -H 'Content-Type: application/json' -H "Authorization: Bearer $token" \
    "$BASE/profile/me" \
    -d "{\"name\":\"$name\",\"directory_opt_in\":$directory}" >/dev/null
  printf '%s' "$token"
}

echo "Logging in as bootstrap admin..."
admin_token=$(login "$ADMIN_EMAIL" "$ADMIN_PASS")
if [ -z "$admin_token" ]; then
  echo "ERROR: Unable to obtain admin token. Is the API up and bootstrap admin configured?" >&2
  exit 1
fi

roster=$(curl -sS -H "Authorization: Bearer $admin_token" "$BASE/admin/users")

echo "Ensuring fake residents..."
maya_token=$(ensure_resident "maya@culdechat.local" "102" "Maya Chen" true)
jordan_token=$(ensure_resident "jordan@culdechat.local" "203" "Jordan Hale" true)
priya_token=$(ensure_resident "priya@culdechat.local" "305" "Priya Shah" true)
sam_token=$(ensure_resident "sam@culdechat.local" "410" "Sam Ortiz" true)
# Hidden from the directory so unit-number DMs can be tested later.
ensure_resident "riley@culdechat.local" "512" "Riley Nguyen" false >/dev/null
# Keep the original smoke-test account.
ensure_resident "seeduser@example.com" "101" "Seed User" true >/dev/null

echo "Creating boards (idempotent)..."
curl -sS -X POST -H 'Content-Type: application/json' -H "Authorization: Bearer $maya_token" \
  "$BASE/boards" -d '{"name":"General","description":"Community catch-all"}' >/dev/null || true
curl -sS -X POST -H 'Content-Type: application/json' -H "Authorization: Bearer $jordan_token" \
  "$BASE/boards" -d '{"name":"For Sale","description":"Buy and sell with neighbors"}' >/dev/null || true

boards=$(curl -sS -H "Authorization: Bearer $maya_token" "$BASE/boards")
gen_id=$(printf '%s' "$boards" | python3 -c '
import sys, json
try:
    arr = json.load(sys.stdin)
    print(next((b.get("id", "") for b in arr if b.get("name") == "General"), ""))
except Exception:
    print("")
')

if [ -z "$gen_id" ]; then
  echo "ERROR: Could not resolve General board id" >&2
  exit 1
fi

echo "General board id: $gen_id"

subscribed() {
  local token="$1" board_id="$2"
  curl -sS -H "Authorization: Bearer $token" "$BASE/boards" | python3 -c '
import sys, json
board_id = sys.argv[1]
try:
    rows = json.load(sys.stdin)
except Exception:
    rows = []
print("yes" if any((r or {}).get("id") == board_id and (r or {}).get("is_subscribed") for r in rows) else "")
' "$board_id"
}

for tok in "$jordan_token" "$priya_token" "$sam_token"; do
  if [ -z "$(subscribed "$tok" "$gen_id")" ]; then
    curl -sS -X POST -H "Authorization: Bearer $tok" "$BASE/boards/$gen_id/subscribe" >/dev/null || true
  fi
done

echo "Seeding sample posts from a few neighbors..."
curl -sS -X POST -H 'Content-Type: application/json' -H "Authorization: Bearer $maya_token" \
  "$BASE/boards/$gen_id/posts" \
  -d '{"title":"Anyone want a walking group?","content":"Thinking Saturday mornings around the courtyard. Ping me if you are in."}' >/dev/null || true
curl -sS -X POST -H 'Content-Type: application/json' -H "Authorization: Bearer $jordan_token" \
  "$BASE/boards/$gen_id/posts" \
  -d '{"title":"Spare folding table","content":"Free to a neighbor. Pickup from 203 this weekend."}' >/dev/null || true
curl -sS -X POST -H 'Content-Type: application/json' -H "Authorization: Bearer $priya_token" \
  "$BASE/boards/$gen_id/posts" \
  -d '{"title":"Book club restart","content":"First pick is a short one. Meeting on the patio Tuesday."}' >/dev/null || true
curl -sS -X POST -H 'Content-Type: application/json' -H "Authorization: Bearer $sam_token" \
  "$BASE/boards/$gen_id/posts" \
  -d '{"title":"Lost a blue water bottle","content":"Left it by the mailboxes last night. Unit 410 if you find it."}' >/dev/null || true

# Grab the newest posts so we can attach comments / reactions.
posts_json=$(curl -sS -H "Authorization: Bearer $maya_token" "$BASE/posts?limit=20")
post_id() {
  printf '%s' "$posts_json" | python3 -c '
import sys, json
title = sys.argv[1]
try:
    data = json.load(sys.stdin)
    posts = data.get("posts") or data or []
    print(next((p.get("id", "") for p in posts if p.get("title") == title), ""))
except Exception:
    print("")
' "$1"
}

walk_id=$(post_id "Anyone want a walking group?")
table_id=$(post_id "Spare folding table")
book_id=$(post_id "Book club restart")
bottle_id=$(post_id "Lost a blue water bottle")

echo "Seeding comments and reactions..."
if [ -n "$walk_id" ]; then
  curl -sS -X POST -H 'Content-Type: application/json' -H "Authorization: Bearer $jordan_token" \
    "$BASE/posts/$walk_id/comments" -d '{"content":"I am in for Saturdays. 8am work?"}' >/dev/null || true
  curl -sS -X POST -H 'Content-Type: application/json' -H "Authorization: Bearer $priya_token" \
    "$BASE/posts/$walk_id/comments" -d '{"content":"Yes! I can do every other week."}' >/dev/null || true
  curl -sS -X PUT -H 'Content-Type: application/json' -H "Authorization: Bearer $sam_token" \
    "$BASE/posts/$walk_id/reactions" -d '{"type":"like"}' >/dev/null || true
fi
if [ -n "$table_id" ]; then
  curl -sS -X POST -H 'Content-Type: application/json' -H "Authorization: Bearer $maya_token" \
    "$BASE/posts/$table_id/comments" -d '{"content":"Still available? I can grab it Sunday."}' >/dev/null || true
  curl -sS -X PUT -H 'Content-Type: application/json' -H "Authorization: Bearer $priya_token" \
    "$BASE/posts/$table_id/reactions" -d '{"type":"love"}' >/dev/null || true
fi
if [ -n "$book_id" ]; then
  curl -sS -X POST -H 'Content-Type: application/json' -H "Authorization: Bearer $maya_token" \
    "$BASE/posts/$book_id/comments" -d '{"content":"Count me in — what is the first book?"}' >/dev/null || true
  curl -sS -X PUT -H 'Content-Type: application/json' -H "Authorization: Bearer $jordan_token" \
    "$BASE/posts/$book_id/reactions" -d '{"type":"like"}' >/dev/null || true
fi
if [ -n "$bottle_id" ]; then
  curl -sS -X POST -H 'Content-Type: application/json' -H "Authorization: Bearer $priya_token" \
    "$BASE/posts/$bottle_id/comments" -d '{"content":"I think I saw a blue one near the recycling bins."}' >/dev/null || true
fi

echo "Seeding direct messages..."
riley_token=$(login "riley@culdechat.local" "$SEED_PASS")

# Maya <-> Jordan (visible neighbors)
curl -sS -X POST -H 'Content-Type: application/json' -H "Authorization: Bearer $maya_token" \
  "$BASE/messages" -d '{"unit_number":"203","content":"Hey Jordan — still have that folding table?"}' >/dev/null || true
curl -sS -X POST -H 'Content-Type: application/json' -H "Authorization: Bearer $jordan_token" \
  "$BASE/messages" -d '{"unit_number":"102","content":"Yep! Free all weekend. Want me to leave it by the lobby?"}' >/dev/null || true
curl -sS -X POST -H 'Content-Type: application/json' -H "Authorization: Bearer $maya_token" \
  "$BASE/messages" -d '{"unit_number":"203","content":"Lobby works — I will grab it Sunday afternoon. Thanks!"}' >/dev/null || true

# Maya <-> Priya
curl -sS -X POST -H 'Content-Type: application/json' -H "Authorization: Bearer $priya_token" \
  "$BASE/messages" -d '{"unit_number":"102","content":"Maya, want to walk together before book club Tuesday?"}' >/dev/null || true
curl -sS -X POST -H 'Content-Type: application/json' -H "Authorization: Bearer $maya_token" \
  "$BASE/messages" -d '{"unit_number":"305","content":"Absolutely. Meet by the courtyard gate at 6:30?"}' >/dev/null || true

# Maya <-> Riley (hidden: unit-only)
if [ -n "$riley_token" ]; then
  curl -sS -X POST -H 'Content-Type: application/json' -H "Authorization: Bearer $maya_token" \
    "$BASE/messages" -d '{"unit_number":"512","content":"Hey — package for 512 was sitting by the mailboxes. I brought it inside the vestibule."}' >/dev/null || true
  curl -sS -X POST -H 'Content-Type: application/json' -H "Authorization: Bearer $riley_token" \
    "$BASE/messages" -d '{"unit_number":"102","content":"Thank you!! Just grabbed it. Really appreciate it."}' >/dev/null || true
fi

echo
echo "Seed accounts (password: $SEED_PASS)"
printf '%-22s %-18s %-8s %s\n' "EMAIL" "NAME" "UNIT" "DIRECTORY"
printf '%-22s %-18s %-8s %s\n' "admin@culdechat.local" "Business Admin" "Admin" "n/a (admin)"
printf '%-22s %-18s %-8s %s\n' "maya@culdechat.local" "Maya Chen" "102" "listed"
printf '%-22s %-18s %-8s %s\n' "jordan@culdechat.local" "Jordan Hale" "203" "listed"
printf '%-22s %-18s %-8s %s\n' "priya@culdechat.local" "Priya Shah" "305" "listed"
printf '%-22s %-18s %-8s %s\n' "sam@culdechat.local" "Sam Ortiz" "410" "listed"
printf '%-22s %-18s %-8s %s\n' "riley@culdechat.local" "Riley Nguyen" "512" "hidden (unit-only DMs)"
printf '%-22s %-18s %-8s %s\n' "seeduser@example.com" "Seed User" "101" "listed"
echo
echo "Suggested test pairs (same password):"
echo "  Maya  <-> Jordan  (visible directory chat about the table)"
echo "  Maya  <-> Priya   (book club / walk)"
echo "  Maya  <-> Riley   (unit 512 — Riley is hidden from People)"
echo "  Open two browsers / profiles: log in as Maya in one, Jordan in the other."
echo
echo "Seeding complete."
