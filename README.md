<p align="center">
  <h1 align="center">LLM Gateway</h1>
  <p align="center">
    <strong>One OpenAI-compatible endpoint for OpenAI, Claude, Gemini, Groq, Ollama, LM Studio, vLLM, and more.</strong>
  </p>
  <p align="center">
    Lightweight, self-hosted, Go-based gateway for developers who want to switch providers without rewriting app code.
  </p>
</p>

<p align="center">
  <a href="https://hub.docker.com/r/scutontech/llm-gateway"><img src="https://img.shields.io/docker/pulls/scutontech/llm-gateway?style=flat-square&logo=docker" alt="Docker Pulls"></a>
  <a href="https://github.com/scuton-technology/llm-gateway/releases"><img src="https://img.shields.io/github/v/release/scuton-technology/llm-gateway?style=flat-square" alt="Release"></a>
  <a href="https://github.com/scuton-technology/llm-gateway/blob/main/LICENSE"><img src="https://img.shields.io/github/license/scuton-technology/llm-gateway?style=flat-square" alt="License"></a>
  <img src="https://img.shields.io/badge/go-1.25-00ADD8?style=flat-square&logo=go" alt="Go 1.25">
</p>

<p align="center">
  <img src="docs/screenshots/dashboard-dark.png" alt="LLM Gateway Dashboard" width="860">
</p>

```bash
http://localhost:8080/v1/chat/completions
```

Use the same OpenAI-style request format and route it to OpenAI, Anthropic, Gemini, Groq, Mistral, Cohere, xAI, Perplexity, Together AI, Ollama, LM Studio, or vLLM.

## Why LLM Gateway?

- Keep your app on one OpenAI-compatible chat completions API.
- Switch providers by changing the model, not the client integration.
- Run it as a single Go binary or one Docker container.
- Use cloud and local/self-hosted providers behind the same endpoint.
- Get built-in streaming, provider settings, request logs, and basic cost analytics.
- Self-host it with SQLite and no extra database or cache.

## Supported Providers

- Cloud: OpenAI, Anthropic, Google Gemini, Groq, Mistral, Cohere, xAI, Perplexity, Together AI
- Local / self-hosted: Ollama, LM Studio, vLLM
- Streaming: OpenAI-format SSE for 11 providers; Cohere is non-streaming
- Routing: provider selection is based on model name or model prefix

Examples:

- `gpt-4o` -> OpenAI
- `claude-sonnet-4-6` -> Anthropic
- `gemini-2.0-flash` -> Google
- `llama3` -> Ollama
- `meta-llama/Llama-3-70b-chat-hf` -> Together AI

## OpenAI SDK Compatible

LLM Gateway is designed to work with existing OpenAI SDK clients for the chat completions API.

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

```javascript
import OpenAI from "openai";

const client = new OpenAI({
  baseURL: "http://localhost:8080/v1",
  apiKey: "unused",
});

const response = await client.chat.completions.create({
  model: "gemini-2.0-flash",
  messages: [{ role: "user", content: "Hello from the same client code" }],
});

console.log(response.choices[0].message.content);
```

Today the compatibility surface is the OpenAI-style `chat.completions` endpoint exposed at `/v1/chat/completions`.

## Quick Start

### Docker

```bash
docker run -p 8080:8080 -v gateway-data:/data scutontech/llm-gateway
```

Then:

1. Open `http://localhost:8080`
2. Create the admin password
3. Add one or more provider API keys in **Settings**
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
    "model": "claude-sonnet-4-6",
    "messages": [{"role": "user", "content": "Write a haiku about Go"}]
  }'
```

### Streaming

```bash
curl -N http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4o",
    "stream": true,
    "messages": [{"role": "user", "content": "Count to five"}]
  }'
```

## What You Get

- OpenAI-compatible request format across multiple providers
- Provider auto-routing based on model name
- Streaming translation into OpenAI SSE format
- Web UI for setup, provider settings, dashboard, and analytics
- Request logging with provider, model, token, latency, and error tracking
- Basic cost estimation in the admin UI
- Admin auth with setup flow, sessions, and lockout protection

## Screenshots

### Dashboard

<p align="center">
  <img src="docs/screenshots/dashboard-dark.png" alt="Dashboard" width="760">
</p>

### Analytics

<p align="center">
  <img src="docs/screenshots/analytics-dark.png" alt="Analytics" width="760">
</p>

### Provider Settings

<p align="center">
  <img src="docs/screenshots/settings-dark.png" alt="Settings" width="760">
</p>

## Why Teams Pick It

LLM Gateway is a good fit when you want a small self-hosted gateway that keeps application code stable while you switch models and providers.

It is especially useful if you:

- already use the OpenAI SDK or OpenAI-style request payloads
- want to compare providers behind one endpoint
- want local models and hosted models behind the same API
- prefer a lightweight Go service over a larger Python stack
- want a built-in UI instead of external admin tooling

## LLM Gateway vs LiteLLM

LiteLLM is a strong option if you need a broader routing and policy surface area. LLM Gateway is aimed at developers who want a smaller self-hosted gateway with a built-in UI and a simpler operational footprint.

| | **LLM Gateway** | **LiteLLM** |
|---|---|---|
| Core runtime | Go | Python |
| Primary focus | Lightweight self-hosted gateway | Broad provider + routing platform |
| API surface in this repo | OpenAI-style chat completions | Wider compatibility surface |
| Local model support | Built-in via OpenAI-compatible backends | Supported |
| Admin UI | Built-in | Typically paired with other tooling |
| Best fit | Small teams, self-hosting, simple deployment | More complex routing and policy setups |

## API Endpoints

| Method | Path | Auth | Description |
|---|---|---|---|
| `POST` | `/v1/chat/completions` | No | OpenAI-compatible proxy endpoint |
| `GET` | `/health` | No | Health check + registered providers |
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
| `GET` | `/api/stats/models` | Yes | Model ranking / token usage |

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
    anthropic.go                Anthropic adapter + stream conversion
    gemini.go                   Gemini adapter + stream conversion
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
web/                            Embedded admin pages
```

## License

MIT

<p align="center">
  Built by <a href="https://scuton.com">Scuton Technology</a>
</p>
