# LLM Gateway

One OpenAI-compatible endpoint for OpenAI, Claude, Gemini, Groq, Ollama, LM Studio, vLLM, and more.

Switch providers by changing the model name, not your application code, and run the gateway as a single Go service with SQLite, built-in admin UI, request logs, analytics, and OpenAI-format streaming.

```bash
docker run -p 8080:8080 -v gateway-data:/data scutontech/llm-gateway
```

[![Docker Pulls](https://img.shields.io/docker/pulls/scutontech/llm-gateway?style=flat-square&logo=docker)](https://hub.docker.com/r/scutontech/llm-gateway)
[![GitHub Release](https://img.shields.io/github/v/release/scuton-technology/llm-gateway?style=flat-square)](https://github.com/scuton-technology/llm-gateway/releases)
[![License](https://img.shields.io/github/license/scuton-technology/llm-gateway?style=flat-square)](https://github.com/scuton-technology/llm-gateway/blob/main/LICENSE)
![Go 1.25](https://img.shields.io/badge/go-1.25-00ADD8?style=flat-square&logo=go)

<p align="center">
  <img src="docs/screenshots/dashboard-dark.png" alt="LLM Gateway dashboard" width="860">
</p>

```bash
curl http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{"model":"claude-sonnet-4-6","messages":[{"role":"user","content":"Hello"}]}'
```

## Why LLM Gateway?

- Keep one stable OpenAI-style `chat.completions` integration in front of multiple providers.
- Switch between hosted and local models without rewriting client code.
- Deploy as one Go binary or one Docker container.
- Run with SQLite and a built-in admin UI instead of extra operational pieces.
- Track requests, latency, token usage, and errors from one place.

## OpenAI SDK Example

If your application already uses the OpenAI SDK, point it at LLM Gateway and change the model:

```python
from openai import OpenAI

client = OpenAI(
    base_url="http://localhost:8080/v1",
    api_key="unused",
)

response = client.chat.completions.create(
    model="claude-sonnet-4-6",
    messages=[{"role": "user", "content": "Hello from the same client code"}],
)

print(response.choices[0].message.content)
```

The compatibility surface in this repo is the OpenAI-style `/v1/chat/completions` endpoint.

## Provider Support

LLM Gateway currently supports:

- Cloud providers: OpenAI, Anthropic, Google Gemini, Groq, Mistral, Cohere, xAI, Perplexity, Together AI
- Local or self-hosted backends: Ollama, LM Studio, vLLM
- Streaming: OpenAI-format SSE for 11 providers; Cohere is non-streaming

Routing examples:

- `gpt-4o` -> OpenAI
- `claude-sonnet-4-6` -> Anthropic
- `gemini-2.0-flash` -> Google Gemini
- `llama3` -> Ollama
- `meta-llama/Llama-3-70b-chat-hf` -> Together AI

## Quick Start

### Docker

```bash
docker run -p 8080:8080 -v gateway-data:/data scutontech/llm-gateway
```

Then:

1. Open `http://localhost:8080`
2. Create the admin password
3. Add provider credentials in **Settings**
4. Send requests to `http://localhost:8080/v1/chat/completions`

If you expose the gateway before setup is complete, remote setup requires the one-time token printed in the startup logs:

```text
Remote setup URL: /admin/setup?token=...
```

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

## What It Includes

- OpenAI-compatible request format for chat completions
- Provider routing based on model name or model prefix
- OpenAI-format SSE streaming
- Web UI for setup, provider settings, dashboard, and analytics
- Request logs with provider, model, token, latency, and error tracking
- Basic cost estimation in the admin UI
- Admin auth with setup flow, sessions, and lockout protection

## Admin UI

### Analytics

<p align="center">
  <img src="docs/screenshots/analytics-dark.png" alt="LLM Gateway analytics" width="760">
</p>

### Provider Settings

<p align="center">
  <img src="docs/screenshots/settings-dark.png" alt="LLM Gateway settings" width="760">
</p>

## LLM Gateway vs LiteLLM

LiteLLM is a strong choice if you need a broader routing, policy, or compatibility surface. LLM Gateway is aimed at teams that want a smaller self-hosted gateway with a built-in UI and a simpler runtime footprint.

| | **LLM Gateway** | **LiteLLM** |
|---|---|---|
| Core runtime | Go | Python |
| Focus | Lightweight self-hosted gateway | Broader provider and routing platform |
| API surface in this repo | OpenAI-style chat completions | Wider compatibility surface |
| Local model support | Built-in via OpenAI-compatible backends | Supported |
| Admin UI | Built-in | Usually paired with other tooling |
| Best fit | Small teams, self-hosting, simple deployment | More complex routing and policy setups |

## API Endpoints

| Method | Path | Auth | Description |
|---|---|---|---|
| `POST` | `/v1/chat/completions` | No | OpenAI-compatible proxy endpoint |
| `GET` | `/health` | No | Health check and registered providers |
| `GET` | `/admin` | Yes | Dashboard UI |
| `GET` | `/admin/analytics` | Yes | Analytics UI |
| `GET` | `/admin/settings` | Yes | Provider settings UI |
| `GET` | `/admin/login` | No | Login page |
| `GET` | `/admin/setup` | No | First-run password setup |
| `GET` | `/api/dashboard` | Yes | Dashboard data |
| `GET` | `/api/stats` | Yes | 24h aggregate stats |
| `GET` | `/api/logs` | Yes | Recent request logs |
| `GET` | `/api/stats/daily` | Yes | Daily statistics |
| `GET` | `/api/stats/monthly` | Yes | Monthly statistics |
| `GET` | `/api/stats/providers` | Yes | Provider breakdown |
| `GET` | `/api/stats/models` | Yes | Model ranking and token usage |

Proxy responses include:

- `X-LLM-Provider`
- `X-LLM-Latency-Ms`

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
    build: .
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

### Provider Settings Priority

- Environment variables override database values
- Values saved in the UI are used when no env var is present
- Empty means the provider is unavailable

## Security Notes

- Admin setup is protected by a first-run password flow
- Remote first-run setup requires the startup setup token unless you connect from localhost
- Admin sessions use secure cookies when the gateway is served over HTTPS
- Provider secrets entered in the UI are encrypted at rest
- Proxy header trust is off by default; enable it only behind a trusted reverse proxy
- Public chat requests are rate-limited by default through `PUBLIC_RATE_LIMIT_RPM`

## Password Reset

```bash
./llm-gateway --reset-password
```

Docker:

```bash
docker exec <container> llm-gateway --reset-password
```

## Architecture

```text
cmd/gateway/main.go              HTTP server, provider registration, auth setup
internal/
  providers/
    interface.go                Provider interfaces and OpenAI-style types
    streaming.go                SSE helpers and passthrough streaming
    openai.go                   OpenAI-compatible providers
    anthropic.go                Anthropic adapter and stream conversion
    gemini.go                   Gemini adapter and stream conversion
    mistral.go                  Mistral adapter
    cohere.go                   Cohere adapter
    xai.go                      xAI adapter
    perplexity.go               Perplexity adapter
    registry.go                 Model to provider resolution
  proxy/router.go               Request routing and streaming dispatch
  admin/
    auth.go                     Setup, login, logout, sessions
    handler.go                  Dashboard, analytics, settings APIs
  middleware/                   Logging, rate limiting, request helpers
  storage/sqlite.go             SQLite storage for logs, auth, settings
web/                            Admin pages
```

## License

MIT
