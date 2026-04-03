# The Dumbest April Fools Joke Ever&#8482;

coahGPT was an april fools joke. the whole thing. i built an entire AI chat app with accounts and email verification and a CLI and an openai-compatible API and a whole persona system that talks like me. it ran on a my server with a 3060 for like 2 days. now its open source so you can run it yourself if you want a chat app that says "ngl" and "ts is fire" in every response.

anyway heres how to actually run it yourself.

## self-hosting

### requirements

- [Go 1.22+](https://go.dev/dl/)
- [Ollama](https://ollama.com) running locally (or anywhere reachable)
- [Node.js 20+](https://nodejs.org) (for building the frontend)

### setup

```bash
# clone it
git clone https://github.com/coah80/coahgpt.git
cd coahgpt

# pull the model (default is qwen3:8b, change in internal/ollama/client.go if u want)
ollama pull qwen3:8b

# build the frontend
cd web
npm install
npm run build
cd ..

# build and run the server
go build -o coahgpt-server ./cmd/coahgpt-server
./coahgpt-server
```

server starts on `http://localhost:8095`. open it in a browser and start chatting.

### environment variables

| variable | default | description |
|----------|---------|-------------|
| `PORT` | `8095` | server port |
| `OLLAMA_URL` | `http://localhost:11434` | ollama API url (if running on another machine) |

### changing the model

edit `internal/ollama/client.go` and change the `Model` constant:

```go
const Model = "qwen3:8b" // change this to whatever u want
```

then rebuild. any model ollama supports works.

### how it works

- **Go backend** serves the API and the static frontend
- **SvelteKit frontend** (static build) talks to the backend via SSE streaming
- **Ollama** handles the actual inference
- conversations are stored in your browser's localStorage. nothing is saved server-side
- theres a system prompt in `internal/persona/prompt.go` that makes it talk like a zoomer. modify or delete it idc
- the `/v1/chat/completions` endpoint is openai-compatible so you can point other tools at it

### project structure

```
cmd/coahgpt-server/    server entrypoint
internal/
  api/                 http handlers, middleware, rate limiting
  chat/                in-memory session store
  ollama/              ollama client
  persona/             system prompt
  search/              web search (duckduckgo scraping)
web/                   sveltekit frontend (catppuccin mocha theme)
```

### features

- streaming chat with SSE
- web search mode (prefix your message with `[Web Search]`)
- deep research mode (prefix with `[Deep Research]`, fetches more results)
- prompt injection detection (tries to catch jailbreaks)
- rate limiting (60 requests/min per IP)
- openai-compatible API at `/v1/chat/completions`
- cat that meows when you click it
