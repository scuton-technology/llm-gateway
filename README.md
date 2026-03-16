# LLM Gateway

Universal LLM proxy. Single Go binary that translates OpenAI-compatible `POST /v1/chat/completions` requests to **12 providers**.

```
curl → LLM Gateway → Anthropic / OpenAI / Gemini / Groq / Mistral / Cohere / xAI / Perplexity / Together / Ollama / LM Studio / vLLM
```

## Features

- **One API, all providers** — Send OpenAI-format requests, gateway translates to each provider's native format
- **Auto-routing** — Model name determines provider (`claude-*` → Anthropic, `gpt-*` → OpenAI, `gemini-*` → Google, etc.)
- **SQLite request logging** — Every request logged with tokens, latency, cost
- **Admin dashboard** — Real-time Chart.js dashboard at `/admin`
- **Cost estimation** — Per-model pricing for all supported models
- **Rate limiting** — IP-based token bucket
- **Zero dependencies** — Single binary, no external services needed (except SQLite)

## Supported Providers

| Provider | Models | Adapter |
|----------|--------|---------|
| **Anthropic** | claude-opus-4, claude-sonnet-4, claude-haiku-4 | Native Messages API |
| **OpenAI** | gpt-4o, gpt-4o-mini, o1, o3-mini | Native |
| **Google** | gemini-2.0-flash, gemini-1.5-pro | Native Gemini API |
| **Groq** | llama-3.3-70b, mixtral-8x7b | OpenAI-compatible |
| **Mistral** | mistral-large, mistral-small, codestral | OpenAI-compatible |
| **Cohere** | command-r-plus, command-r | Native Chat API |
| **xAI** | grok-2, grok-2-mini | OpenAI-compatible |
| **Perplexity** | sonar-large, sonar-small | OpenAI-compatible |
| **Together AI** | meta-llama/Llama-3-70b, etc. | OpenAI-compatible |
| **Ollama** | llama3, mistral, codellama, phi3 | OpenAI-compatible (local) |
| **LM Studio** | Any loaded model | OpenAI-compatible (local) |
| **vLLM** | Any served model | OpenAI-compatible (local) |

## Quick Start

```bash
# Clone
git clone https://github.com/scuton-technology/llm-gateway.git
cd llm-gateway

# Configure
cp .env.example .env
# Edit .env — add at least one API key

# Run
go run ./cmd/gateway

# Or build & run
go build -o llm-gateway ./cmd/gateway
./llm-gateway
```

## Docker

```bash
cp .env.example .env
# Edit .env with your API keys

docker compose up -d
```

## Usage

Send standard OpenAI chat completion requests:

```bash
# Anthropic
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-sonnet-4",
    "messages": [{"role": "user", "content": "Hello!"}]
  }'

# OpenAI
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4o",
    "messages": [{"role": "user", "content": "Hello!"}]
  }'

# Google Gemini
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gemini-2.0-flash",
    "messages": [{"role": "user", "content": "Hello!"}]
  }'
```

The gateway automatically routes to the correct provider based on the model name.

## Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/v1/chat/completions` | Proxy endpoint (OpenAI-compatible) |
| `GET` | `/health` | Health check + registered providers |
| `GET` | `/admin` | Dashboard UI |
| `GET` | `/api/stats` | Usage statistics (JSON) |
| `GET` | `/api/logs` | Recent request logs (JSON) |
| `GET` | `/api/dashboard` | Combined dashboard data (JSON) |

## Response Headers

Every proxied response includes:

- `X-LLM-Provider` — Which provider handled the request
- `X-LLM-Latency-Ms` — End-to-end latency in milliseconds

## Environment Variables

```bash
# Server
PORT=8080              # Server port (default: 8080)
DB_PATH=gateway.db     # SQLite database path
DASHBOARD_PATH=...     # Custom dashboard.html path

# Cloud Providers — set API key to enable
ANTHROPIC_API_KEY=
OPENAI_API_KEY=
GOOGLE_API_KEY=
GROQ_API_KEY=
MISTRAL_API_KEY=
COHERE_API_KEY=
XAI_API_KEY=
PERPLEXITY_API_KEY=
TOGETHER_API_KEY=

# Local Providers — set ENABLED=true to activate
OLLAMA_ENABLED=false
OLLAMA_BASE_URL=http://localhost:11434
LMSTUDIO_ENABLED=false
LMSTUDIO_BASE_URL=http://localhost:1234
VLLM_ENABLED=false
VLLM_BASE_URL=http://localhost:8000
```

## Model Routing

The gateway resolves providers by model name prefix:

| Prefix | Provider |
|--------|----------|
| `claude-*` | Anthropic |
| `gpt-*`, `o1`, `o3-*` | OpenAI |
| `gemini-*` | Google |
| `llama*`, `mixtral-*` | Groq |
| `mistral-*` | Mistral |
| `command-*` | Cohere |
| `grok-*` | xAI |
| `sonar-*` | Perplexity |
| `*/` (contains slash) | Together AI |

## Architecture

```
cmd/gateway/main.go          HTTP server, provider registration
internal/
  providers/
    interface.go              Provider interface + OpenAI types
    openai.go                 OpenAI + Groq + Ollama + LM Studio + vLLM
    anthropic.go              Anthropic Messages API translation
    gemini.go                 Google Gemini API translation
    mistral.go                Mistral (OpenAI-compat, own endpoint)
    cohere.go                 Cohere Chat API translation
    xai.go                    xAI Grok (OpenAI-compat, own endpoint)
    perplexity.go             Perplexity Sonar (OpenAI-compat)
    together.go               Together AI (uses OpenAI adapter)
    registry.go               Model → provider resolution
  proxy/router.go             Request routing + response formatting
  middleware/
    logging.go                HTTP request logging
    ratelimit.go              IP-based rate limiter
    cost.go                   Per-model cost estimation
  storage/sqlite.go           SQLite logging + statistics
  admin/handler.go            Dashboard data API
web/dashboard.html            Admin dashboard (Chart.js)
```

## License

MIT
