# LLM Gateway

**One OpenAI-compatible endpoint. Every provider. Zero lock-in.**

Stop rewriting application code every time you switch LLM providers. LLM Gateway sits in front of OpenAI, Claude, Gemini, Groq, Ollama, and more — exposing a single `/v1/chat/completions` endpoint your app already knows how to talk to.

```bash
docker run -p 8080:8080 -v gateway-data:/data scutontech/llm-gateway
```

[![Docker Pulls](https://img.shields.io/docker/pulls/scutontech/llm-gateway?style=flat-square&logo=docker)](https://hub.docker.com/r/scutontech/llm-gateway)
[![GitHub Release](https://img.shields.io/github/v/release/scuton-technology/llm-gateway?style=flat-square)](https://github.com/scuton-technology/llm-gateway/releases)
[![Go 1.25](https://img.shields.io/badge/go-1.25-00ADD8?style=flat-square&logo=go)](https://golang.org)
[![License: MIT](https://img.shields.io/github/license/scuton-technology/llm-gateway?style=flat-square)](LICENSE)

<p align="center">
  <img src="docs/screenshots/dashboard-dark.png" alt="LLM Gateway — request dashboard" width="900">
</p>

---

## The Problem

Your app calls OpenAI today. Tomorrow you want to benchmark Claude. Next week you want Ollama in staging. Every switch means touching API client code, auth headers, request formats, and streaming parsers.

**LLM Gateway solves this once.** You call one endpoint. The gateway translates.

---

## Why LLM Gateway?

- **Provider-agnostic routing** — change the `model` field, nothing else
- **Real streaming** — OpenAI-format SSE across 11 providers, including format conversion for Anthropic and Gemini
- **Zero external dependencies** — Go binary + SQLite. No Redis. No Postgres. No sidecars.
- **API keys encrypted at rest** — AES-256-GCM with per-key random nonces
- **Built-in observability** — every request logged with provider, model, tokens, latency, cost, and client IP
- **Admin UI included** — settings, dashboard, and analytics out of the box
- **One-command deploy** — `docker run` and you're live in under a minute

---

## Drop-in OpenAI SDK Replacement

Already using the OpenAI SDK? Point it at the gateway. Change the model. Nothing else.

```python
from openai import OpenAI

client = OpenAI(
    base_url="http://localhost:8080/v1",
    api_key="any-string",               # gateway handles real auth
)

# This exact code works for every provider below
response = client.chat.completions.create(
    model="claude-sonnet-4-6",          # or gpt-4o, gemini-2.0-flash, llama3 …
    messages=[{"role": "user", "content": "Hello"}],
)
print(response.choices[0].message.content)
```

Works the same way in Node.js, Go, Ruby, or any other OpenAI SDK.

---

## Supported Providers

| Provider | Type | Streaming | Notes |
|---|---|:---:|---|
| OpenAI | Cloud | ✓ | gpt-4o, o1, o3-mini, … |
| Anthropic | Cloud | ✓ | claude-3-5, claude-sonnet-4, … |
| Google Gemini | Cloud | ✓ | gemini-2.0-flash, gemini-1.5-pro |
| Groq | Cloud | ✓ | llama-3.3-70b, mixtral-8x7b |
| Mistral | Cloud | ✓ | mistral-large, codestral |
| Cohere | Cloud | — | command-r, command-r-plus |
| xAI | Cloud | ✓ | grok-2, grok-2-mini |
| Perplexity | Cloud | ✓ | sonar-large, sonar-small |
| Together AI | Cloud | ✓ | meta-llama/Llama-3-70b, … |
| Ollama | Local | ✓ | any model you've pulled |
| LM Studio | Local | ✓ | any loaded model |
| vLLM | Local | ✓ | any deployed model |

### Automatic Model Routing

No configuration required. The gateway resolves the right provider from the model name:

| Model pattern | Routed to |
|---|---|
| `gpt-*`, `o1`, `o3-mini` | OpenAI |
| `claude-*` | Anthropic |
| `gemini-*` | Google Gemini |
| `llama*`, `mixtral-*` | Groq |
| `mistral-*`, `codestral*` | Mistral |
| `command-*` | Cohere |
| `grok-*` | xAI |
| `sonar-*` | Perplexity |
| `org/model` (slash in name) | Together AI |
| Any other name | Ollama → LM Studio → vLLM (first enabled) |

---

## Quick Start

### Docker (recommended)

```bash
docker run -p 8080:8080 -v gateway-data:/data scutontech/llm-gateway
```

1. Open `http://localhost:8080`
2. Set the admin password
3. Add your provider API keys in **Settings**
4. Send requests to `http://localhost:8080/v1/chat/completions`

> **Remote setup**: if you deploy on a server before setup, use the one-time token printed at startup:
> ```
> Remote setup URL: /admin/setup?token=<token>
> ```

### Build From Source

```bash
git clone https://github.com/scuton-technology/llm-gateway.git
cd llm-gateway
go build -o llm-gateway ./cmd/gateway
./llm-gateway
```

### Your First Request

```bash
curl http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4o",
    "messages": [{"role": "user", "content": "Write a haiku about Go"}]
  }'
```

Switch providers by changing `model` — the request format stays identical:

```bash
# Same request, different provider
curl http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-sonnet-4-6",
    "messages": [{"role": "user", "content": "Write a haiku about Go"}]
  }'
```

### Streaming

```bash
curl -N http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-sonnet-4-6",
    "stream": true,
    "messages": [{"role": "user", "content": "Count to five, slowly"}]
  }'
```

Anthropic and Gemini stream events are converted to OpenAI SSE format in real-time. Your client sees a uniform stream regardless of provider.

---

## What's Inside

### Proxy

- OpenAI-compatible `/v1/chat/completions` endpoint
- OpenAI-format SSE streaming across 11 providers
- Real-time format conversion for Anthropic and Gemini streams
- 2 MB request size limit, 429 rate limiting per IP
- Response headers: `X-LLM-Provider`, `X-LLM-Latency-Ms`

### Observability

- Every request logged: provider, model, prompt tokens, completion tokens, latency, status, error, client IP
- Dashboard: live request feed, per-model and per-provider breakdowns
- Analytics: daily and monthly trends, provider token usage, model cost breakdown
- CSV export

### Security

- Admin password hashed with **bcrypt cost 12**
- Provider API keys encrypted with **AES-256-GCM** (unique nonce per key)
- Session tokens stored as **SHA-256 hashes**, 24-hour expiry
- Proxy header trust (`X-Forwarded-For`) disabled by default
- Public `/v1` rate-limited at `PUBLIC_RATE_LIMIT_RPM` requests/minute

### Runtime

- Single Go binary — no runtime dependencies
- SQLite with WAL mode — crash-safe, indexed for analytics queries
- In-memory per-IP token bucket rate limiter
- Hot-reload of provider settings without restart

---

## Admin UI

### Dashboard

<p align="center">
  <img src="docs/screenshots/dashboard-dark.png" alt="Dashboard — live request feed and stats" width="860">
</p>

Total requests, tokens, errors, and average latency at a glance. Live request log with provider, model, status, and latency per row.

### Analytics

<p align="center">
  <img src="docs/screenshots/analytics-dark.png" alt="Analytics — daily trends and cost breakdown" width="860">
</p>

Daily and monthly request trends. Per-provider token usage. Per-model cost breakdown. Switchable between day, week, and month views.

### Provider Settings

<p align="center">
  <img src="docs/screenshots/settings-dark.png" alt="Settings — provider configuration" width="860">
</p>

Configure API keys for all 12 providers in one place. Keys are masked in the UI and encrypted in the database. Test any provider connection before saving.

---

## LLM Gateway vs LiteLLM

LiteLLM is the better choice for teams that need 100+ provider integrations, virtual key management, load balancing across multiple API keys, or enterprise routing policies.

LLM Gateway is built for teams that want a **self-contained gateway** they can deploy in minutes with no operational overhead.

| | LLM Gateway | LiteLLM |
|---|---|---|
| Runtime | Go — single binary | Python |
| Storage | SQLite (embedded) | Postgres required for most features |
| Extra services | None | Redis, separate dashboard |
| Admin UI | Built-in | Separate dashboard container |
| Provider coverage | 12 | 100+ |
| Streaming | OpenAI SSE, all providers | OpenAI SSE |
| API key encryption | AES-256-GCM at rest | Config-dependent |
| Best fit | Lightweight self-hosted gateway | Complex routing and key management |

---

## API Reference

| Method | Path | Auth | Description |
|---|---|---|---|
| `POST` | `/v1/chat/completions` | No | OpenAI-compatible proxy |
| `GET` | `/health` | No | Status and registered providers |
| `GET` | `/admin` | Yes | Dashboard |
| `GET` | `/admin/analytics` | Yes | Analytics |
| `GET` | `/admin/settings` | Yes | Provider settings |
| `GET` | `/admin/login` | No | Login |
| `GET` | `/admin/setup` | No | First-run setup |
| `GET` | `/api/dashboard` | Yes | Dashboard JSON |
| `GET` | `/api/stats` | Yes | 24 h aggregate stats |
| `GET` | `/api/logs` | Yes | Recent request logs |
| `GET` | `/api/stats/daily` | Yes | Daily stats (up to 365 days) |
| `GET` | `/api/stats/monthly` | Yes | Monthly stats (up to 36 months) |
| `GET` | `/api/stats/providers` | Yes | Per-provider daily breakdown |
| `GET` | `/api/stats/models` | Yes | Per-model token and cost breakdown |

---

## Configuration

### Environment Variables

```bash
# Server
PORT=8080
DB_PATH=gateway.db
PUBLIC_RATE_LIMIT_RPM=60          # 0 disables rate limiting

# Security
LLM_GATEWAY_ENCRYPTION_KEY=       # if unset, a key is generated and saved to ~/.llm-gateway.key
LLM_GATEWAY_TRUST_PROXY_HEADERS=false

# Cloud providers
OPENAI_API_KEY=
ANTHROPIC_API_KEY=
GOOGLE_API_KEY=
GROQ_API_KEY=
MISTRAL_API_KEY=
COHERE_API_KEY=
XAI_API_KEY=
PERPLEXITY_API_KEY=
TOGETHER_API_KEY=

# Local providers (disabled by default)
OLLAMA_ENABLED=false
OLLAMA_BASE_URL=http://localhost:11434
LMSTUDIO_ENABLED=false
LMSTUDIO_BASE_URL=http://localhost:1234
VLLM_ENABLED=false
VLLM_BASE_URL=http://localhost:8000
```

Environment variables take priority over keys saved in the admin UI. An empty value disables the provider.

### Docker Compose

```yaml
services:
  gateway:
    image: scutontech/llm-gateway
    ports:
      - "8080:8080"
    volumes:
      - gateway-data:/data
    env_file:
      - .env
    restart: unless-stopped

volumes:
  gateway-data:
```

---

## Password Reset

```bash
# Binary
./llm-gateway --reset-password

# Docker
docker exec <container> llm-gateway --reset-password
```

---

## Architecture

```
cmd/gateway/main.go
  HTTP server setup, provider registration, auth wiring

internal/proxy/router.go
  POST /v1/chat/completions — resolves provider, routes request, logs result
  Handles both streaming and non-streaming paths

internal/providers/
  interface.go        — Provider interface + OpenAI-style request/response types
  registry.go         — Model-to-provider resolution (prefix + exact match)
  streaming.go        — SSE helpers, chunk parsing, OpenAI format passthrough
  openai.go           — OpenAI + all OpenAI-compatible backends (Groq, Mistral, Perplexity,
                        xAI, Together AI, Ollama, LM Studio, vLLM)
  anthropic.go        — Anthropic adapter: message format conversion + stream translation
  gemini.go           — Gemini adapter: role mapping + stream translation
  cohere.go           — Cohere adapter: chat_history format + non-streaming response

internal/admin/
  auth.go             — Setup flow, login, logout, bcrypt password, session management
  handler.go          — Dashboard, analytics, and settings API handlers

internal/middleware/
  logging.go          — HTTP request/response logging
  ratelimit.go        — Per-IP token bucket rate limiter
  cost.go             — Per-model cost estimation from token counts

internal/storage/sqlite.go
  Schema: request_logs, provider_settings, admin_config, sessions, login_attempts
  WAL mode, indexed queries, AES-256-GCM key encryption

web/
  dashboard.html, analytics.html, settings.html, login.html, setup.html
```

---

## License

[MIT](LICENSE)
