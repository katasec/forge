#!/usr/bin/env bash
# Exercise the forge-agent serve path end-to-end against a running server.
#
# Start the server first (see scripts/serve.ps1), then run this. No API key is
# needed here — we only curl localhost; the server holds the upstream key.
#
#   ./scripts/demo.sh                         # default http://localhost:8787/v1
#   ./scripts/demo.sh http://localhost:9000/v1
set -euo pipefail

BASE="${1:-http://localhost:8787/v1}"
PROMPT='Review a small Go HTTP server package and suggest the single next useful improvement.'

pp() { python3 -c 'import sys,json; d=json.load(sys.stdin); print(d.get("choices",[{}])[0].get("message",{}).get("content", d))'; }

echo "== GET $BASE/models =="
curl -s "$BASE/models" | python3 -m json.tool

for model in vanilla_reviewer forged_reviewer; do
  echo
  echo "== POST $BASE/chat/completions  (model=$model) =="
  curl -s "$BASE/chat/completions" \
    -H 'Content-Type: application/json' \
    -d "{\"model\":\"$model\",\"messages\":[{\"role\":\"user\",\"content\":\"$PROMPT\"}]}" \
    | pp
done

echo
echo "== streaming  (model=forged_reviewer, stream=true) =="
curl -sN "$BASE/chat/completions" \
  -H 'Content-Type: application/json' \
  -d "{\"model\":\"forged_reviewer\",\"messages\":[{\"role\":\"user\",\"content\":\"$PROMPT\"}],\"stream\":true}"
echo
