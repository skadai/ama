# Ama API reference

This skill uses `amacli`, which wraps the Ask Me Anything API.

Default base URL:

```text
http://localhost:3000
```

## Authentication

Protected endpoints accept:

```http
Authorization: Bearer YOUR_API_KEY
```

`amacli` sends:

```http
Authorization: Bearer YOUR_API_KEY
x-ama-client-type: cli
```

## CLI commands

### `amacli health`

```bash
amacli health
```

### `amacli auth login`

```bash
amacli auth login
```

### `amacli auth complete`

```bash
amacli auth complete
```

### `amacli auth status`

```bash
amacli auth status
```

### `amacli me`

```bash
amacli me
```

### `amacli language show`

```bash
amacli language show
```

### `amacli language set`

```bash
amacli language set zh
amacli language set en
```

### `amacli source list`

```bash
amacli source list
```

### `amacli source set-default`

```bash
amacli source set-default lenny
```

### `amacli search`

```bash
amacli search --query "How does Lenny think about MVP scope?" --top-k 5
```

Supported flags:
- `--query`, `-q`
- `--top-k`
- `--source` (repeatable)
- `--content-type` (repeatable)

### `amacli document`

```bash
amacli document --source lenny --id 42
```

Shorthand:

```bash
amacli doc lenny 42
```

### `amacli save-answer`

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

### `POST /v1/saved-answers`

```bash
curl -X POST 'http://localhost:3000/v1/saved-answers' \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer YOUR_API_KEY' \
  -d '{
    "question": "What does Lenny say about PM hiring?",
    "answer": "Structured interviews matter...",
    "source_slugs": ["lenny"],
    "citations": []
  }'
```
