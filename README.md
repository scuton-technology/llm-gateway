# LLM Gateway

Route requests to OpenAI, Claude, Gemini, Groq, Ollama, LM Studio, vLLM, and more — through one OpenAI-compatible endpoint.

Deploy as a single Go binary or Docker container. No Redis. No Postgres. No extra services.

```bash
docker run -p 8080:8080 -v gateway-data:/data scutontech/llm-gateway
```

[![Docker Pulls](https://img.shields.io/docker/pulls/scutontech/llm-gateway?style=flat-square&logo=docker)](https://hub.docker.com/r/scutontech/llm-gateway)
[![GitHub Release](https://img.shields.io/github/v/release/scuton-technology/llm-gateway?style=flat-square)](https://github.com/scuton-technology/llm-gateway/releases)
[![License](https://img.shields.io/github/license/scuton-technology/llm-gateway?style=flat-square)](LICENSE)
![Go 1.25](https://img.shields.io/badge/go-1.25-00ADD8?style=flat-square&logo=go)

<p align="center">
  <img src="docs/screenshots/dashboard-dark.png" alt="LLM Gateway dashboard" width="860">
</p>

---

## Why LLM Gateway?

Most teams end up with provider-specific code scattered across their codebase. LLM Gateway eliminates that by presenting a single OpenAI-compatible surface your application talks to — and routing requests to whichever backend you configure.

- Switch providers by changing the model name, not your application code
- Use the same OpenAI SDK for every provider — hosted and local
- Run behind one endpoint with no additional infrastructure
- Track requests, costs, latency, and errors in a built-in admin UI
- Deploy anywhere a single binary runs

---

## OpenAI SDK Compatibility

If your application already uses the OpenAI SDK, point it at LLM Gateway and change the model name:

```python
from openai import OpenAI

client = OpenAI(
    base_url="http://localhost:8080/v1",
    api_key="unused",
)

response = client.chat.completions.create(
    model="claude-sonnet-4-6",       # or gpt-4o, gemini-2.0-flash, llama3 ...
    messages=[{"role": "user", "content": "Hello"}],
)
print(response.choices[0].message.content)
```

No other changes required.

---

## Supported Providers

| Provider | Type | Streaming |
|---|---|:---:|
| OpenAI | Cloud | ✓ |
| Anthropic | Cloud | ✓ |
| Google Gemini | Cloud | ✓ |
| Groq | Cloud | ✓ |
| Mistral | Cloud | ✓ |
| Cohere | Cloud | — |
| xAI | Cloud | ✓ |
| Perplexity | Cloud | ✓ |
| Together AI | Cloud | ✓ |
| Ollama | Local | ✓ |
| LM Studio | Local | ✓ |
| vLLM | Local | ✓ |

**Model routing** is resolved automatically from the model name:

| Model name pattern | Routes to |
|---|---|
| `gpt-*`, `o1`, `o3-mini` | OpenAI |
| `claude-*` | Anthropic |
| `gemini-*` | Google Gemini |
| `llama*`, `mixtral-*` | Groq |
| `mistral-*`, `codestral` | Mistral |
| `command-*` | Cohere |
| `grok-*` | xAI |
| `sonar-*` | Perplexity |
| `org/model` (slash in name) | Together AI |
| Any other name | Ollama → LM Studio → vLLM (first enabled) |

---

## Quick Start

### Docker

```bash
docker run -p 8080:8080 -v gateway-data:/data scutontech/llm-gateway
```

1. Open `http://localhost:8080`
2. Create the admin password
3. Add provider API keys in **Settings**
4. Send requests to `http://localhost:8080/v1/chat/completions`

> If you access the gateway remotely before setup completes, use the one-time token printed in the startup logs:
> ```
> Remote setup URL: /admin/setup?token=...
> ```

### Build From Source

```bash
git clone https://github.com/scuton-technology/llm-gateway.git
cd llm-gateway
go build -o llm-gateway ./cmd/gateway
./llm-gateway
```

### First Request

```bash
curl http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4o",
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
    "messages": [{"role": "user", "content": "Count to five"}]
  }'
```

---

## Features

- OpenAI-compatible `/v1/chat/completions` endpoint
- OpenAI-format SSE streaming across 11 providers
- Automatic model-to-provider routing by name
- Built-in admin UI: setup, settings, dashboard, analytics
- Request logs with provider, model, tokens, latency, and error tracking
- Basic cost estimation per request and provider
- Provider secrets encrypted at rest in SQLite
- Rate limiting on public endpoints
- Single binary or Docker deployment — SQLite only, no external dependencies

---

## Admin UI

### Analytics

<p align="center">
  <img src="docs/screenshots/analytics-dark.png" alt="LLM Gateway analytics" width="760">
</p>

### Provider Settings

<p align="center">
  <img src="docs/screenshots/settings-dark.png" alt="LLM Gateway settings" width="760">
</p>

---

## LLM Gateway vs LiteLLM

LiteLLM is the right choice if you need broad provider coverage, load balancing across multiple keys, virtual key management, or a managed proxy with enterprise policy controls.

LLM Gateway is built for teams that want a smaller, self-contained gateway they can deploy in minutes without operating a Python service or additional infrastructure.

| | LLM Gateway | LiteLLM |
|---|---|---|
| Runtime | Go — single binary | Python |
| Storage | SQLite (built-in) | Postgres required for most features |
| Admin UI | Built-in | Separate dashboard service |
| Deployment | One container, no dependencies | Multiple services |
| Streaming | OpenAI-format SSE | OpenAI-format SSE |
| Provider coverage | 12 providers | 100+ providers |
| Best fit | Simple self-hosted gateway | Complex routing, policy, and key management |

---

## API Reference

| Method | Path | Auth | Description |
|---|---|---|---|
| `POST` | `/v1/chat/completions` | No | OpenAI-compatible proxy |
| `GET` | `/health` | No | Health check and registered providers |
| `GET` | `/admin` | Yes | Dashboard UI |
| `GET` | `/admin/analytics` | Yes | Analytics UI |
| `GET` | `/admin/settings` | Yes | Provider settings UI |
| `GET` | `/admin/login` | No | Login page |
| `GET` | `/admin/setup` | No | First-run setup |
| `GET` | `/api/dashboard` | Yes | Dashboard data |
| `GET` | `/api/stats` | Yes | 24h aggregate stats |
| `GET` | `/api/logs` | Yes | Recent request logs |
| `GET` | `/api/stats/daily` | Yes | Daily statistics |
| `GET` | `/api/stats/monthly` | Yes | Monthly statistics |
| `GET` | `/api/stats/providers` | Yes | Provider breakdown |
| `GET` | `/api/stats/models` | Yes | Model ranking and token usage |

Proxy responses include `X-LLM-Provider` and `X-LLM-Latency-Ms` headers.

---

## Configuration

### Environment Variables

```bash
# Server
PORT=8080
DB_PATH=gateway.db
PUBLIC_RATE_LIMIT_RPM=60

# Security
LLM_GATEWAY_ENCRYPTION_KEY=changeme
LLM_GATEWAY_TRUST_PROXY_HEADERS=false

# Cloud providers
ANTHROPIC_API_KEY=
OPENAI_API_KEY=
GOOGLE_API_KEY=
GROQ_API_KEY=
MISTRAL_API_KEY=
COHERE_API_KEY=
XAI_API_KEY=
PERPLEXITY_API_KEY=
TOGETHER_API_KEY=

# Local providers
OLLAMA_ENABLED=false
OLLAMA_BASE_URL=http://localhost:11434
LMSTUDIO_ENABLED=false
LMSTUDIO_BASE_URL=http://localhost:1234
VLLM_ENABLED=false
VLLM_BASE_URL=http://localhost:8000
```

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

### Settings Priority

Environment variables override database values. Keys saved in the admin UI are used when no environment variable is set. An empty value disables the provider.

---

## Security

- First-run setup requires a password and, for remote access, a one-time startup token
- Admin sessions use secure cookies when served over HTTPS
- Provider API keys are encrypted at rest in SQLite
- Proxy header trust (`X-Forwarded-For`) is disabled by default
- Public `/v1` requests are rate-limited via `PUBLIC_RATE_LIMIT_RPM`

### Password Reset

```bash
# Binary
./llm-gateway --reset-password

# Docker
docker exec <container> llm-gateway --reset-password
```

---

## Architecture

```
cmd/gateway/main.go              HTTP server, provider registration, auth setup
internal/
  providers/
    interface.go                 Provider interface and OpenAI-style types
    streaming.go                 SSE helpers and passthrough streaming
    openai.go                    OpenAI-compatible backends (OpenAI, Groq, Ollama, LM Studio, vLLM, xAI, Perplexity, Together AI)
    anthropic.go                 Anthropic adapter with stream conversion
    gemini.go                    Gemini adapter with stream conversion
    mistral.go                   Mistral adapter
    cohere.go                    Cohere adapter
    registry.go                  Model-to-provider resolution
  proxy/router.go                Request routing and streaming dispatch
  admin/
    auth.go                      Setup, login, logout, session management
    handler.go                   Dashboard, analytics, and settings APIs
  middleware/                    Logging, rate limiting, cost tracking
  storage/sqlite.go              SQLite storage for logs, auth, and settings
web/                             Admin UI pages
```

---

## License

MIT
