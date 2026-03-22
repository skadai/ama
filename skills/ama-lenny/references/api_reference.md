# Ama Lenny API reference

This skill uses `amacli`, which wraps the local Ama API.

Default base URL:

```text
http://localhost:3000
```

## Authentication

Protected endpoints accept either:

```http
Authorization: Bearer YOUR_API_KEY
```

or:

```http
x-api-key: YOUR_API_KEY
```

`amacli` currently sends:

```http
Authorization: Bearer YOUR_API_KEY
x-ama-client-type: cli
```

## CLI commands

### `amacli health`

Checks server health.

```bash
amacli health
```

### `amacli auth login`

Starts the browser approval flow and stores a pending device session locally.

```bash
amacli auth login
```

### `amacli auth complete`

Claims the approved device session, creates the API key on the server, and writes it to local `config.json`.

```bash
amacli auth complete
```

### `amacli auth status`

Shows whether local CLI auth is already configured and whether a pending browser authorization exists.

```bash
amacli auth status
```

### `amacli me`

Inspects the current API key, user, and source access.

```bash
amacli me
```

### `amacli search`

Searches indexed Lenny content.

```bash
amacli search --query "How does Lenny think about MVP scope?" --top-k 5
```

Supported flags:
- `--query`, `-q`
- `--top-k`
- `--source` (repeatable)
- `--content-type` (repeatable)

Current useful content types:
- `podcast_episode`
- `newsletter_article`
- `video`

### `amacli document`

Fetches original markdown content for a result.

```bash
amacli document --source lenny --id 42
```

Shorthand:

```bash
amacli doc lenny 42
```

### `amacli save-answer`

Saves one final answer into the user's dashboard knowledge wall.

```bash
cat answer.md | amacli save-answer \
  --question "What does Lenny say about PM hiring?" \
  --citations-file citations.json \
  --source lenny
```

## Raw API equivalents

### `GET /v1/health`

```bash
curl 'http://localhost:3000/v1/health'
```

### `GET /v1/me`

```bash
curl 'http://localhost:3000/v1/me' \
  -H 'Authorization: Bearer YOUR_API_KEY'
```

### `POST /v1/search`

```bash
curl -X POST 'http://localhost:3000/v1/search' \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer YOUR_API_KEY' \
  -H 'x-ama-client-type: cli' \
  -d '{
    "query": "How does Lenny think about MVP scope?",
    "sources": ["lenny"],
    "top_k": 5
  }'
```

### `GET /v1/documents/:sourceSlug/:articleId`

```bash
curl 'http://localhost:3000/v1/documents/lenny/42' \
  -H 'Authorization: Bearer YOUR_API_KEY'
```

## Working fields to keep from responses

When reasoning in the skill, preserve these fields from search responses when available:

- `id`
- `source_slug`
- `document_path`
- `title`
- `type`
- `guest`
- `date`
- `summary`
- `source_path`
- `score`
- `recall_sources`

When opening a document response, preserve:

- `title`
- `type`
- `guest`
- `date`
- `summary`
- `source_path`
- `content`
- `content_format`
- `tags`

Hide these working details in the final polished answer unless the user explicitly asks for retrieval mechanics.

### `POST /v1/saved-answers`

```bash
curl -X POST 'http://localhost:3000/v1/saved-answers' \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer YOUR_API_KEY' \
  -d '{
    "question": "What does Lenny say about PM hiring?",
    "answer": "My take: structured interviews matter...",
    "source_slugs": ["lenny"],
    "citations": []
  }'
```
