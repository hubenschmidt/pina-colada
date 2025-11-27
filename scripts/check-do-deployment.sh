#!/bin/bash
# Check DigitalOcean App Platform deployment status

DO_TOKEN="${DO_API_TOKEN:-dop_v1_4dbe0ba968624f6e4ee6a443cc6e87db320834f1a0e4235618fad62d8578e7c2}"
DO_APP_ID="${DO_APP_ID:-78d80fa1-9b73-4c13-a20c-8553c8cf92d4}"

curl -s -H "Authorization: Bearer $DO_TOKEN" \
  "https://api.digitalocean.com/v2/apps/$DO_APP_ID/deployments" | python3 -c "
import json, sys
data = json.load(sys.stdin)
d = data['deployments'][0]
print(f\"Phase: {d['phase']}\")
print(f\"Created: {d['created_at']}\")
print(f\"Cause: {d.get('cause', 'N/A')}\")
for svc in d.get('spec', {}).get('services', []):
    img = svc.get('image', {})
    print(f\"Service: {svc.get('name')} -> {img.get('repository')}:{img.get('tag')}\")
"
