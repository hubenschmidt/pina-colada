# OpenAI Purge — Post-Implementation Audit

## Completed

- Backend: removed OpenAI client, API key config, and all GPT model routing in `orchestrator.go`, `config.go`, `agent_controller.go`
- Embedding service: replaced OpenAI embeddings with Voyage AI (`embedding_service.go`)
- Migrations: `109_voyage_embeddings.up.sql`, `110_default_provider_anthropic.up.sql`
- Frontend: removed GPT options from `AgentConfigMenu.jsx` and `usage/page.jsx`

## Missed

- `modules/client/app/automation/page.jsx` lines 95-96 — GPT 5.2 / GPT 5.1 in `AGENT_MODEL_OPTIONS` dropdown
- **Fixed**: removed both entries

## Intentionally Kept References

| File | Reference | Reason |
|---|---|---|
| `app/robots.js:9` | `GPTBot`, `ChatGPT-User` | Web crawler bot block list — unrelated to provider |
| `Chat.module.css:378` | `/* ChatGPT/Claude style */` | CSS comment — cosmetic |
| `docs/agent-config-sidebar.md` | Historical spec doc | Docs excluded per original plan |
| `app/page.jsx:151` | "OpenAI Agents SDK" | Framework name — factually accurate |
| `cost.go` | GPT pricing data | Historical cost records — kept per original plan |

## Verification

1. "Edit Crawler" dialog → Agent Model dropdown shows only Claude models
2. Existing crawlers with `gpt-5.1`/`gpt-5.2` in DB show raw value (not in dropdown); user must re-select a Claude model
