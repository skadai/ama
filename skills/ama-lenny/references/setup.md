# Ama Lenny setup

This skill now depends on `amacli`, which talks to the Ama API.

Default local server:

```bash
http://localhost:3000
```

## 1. Install Go 1.25

Use Go `1.25` or newer for `amacli`.

## 2. Build or install `amacli`

From the repo root:

```bash
cd amacli
go install .
```

If you prefer a local binary instead of a global install:

```bash
cd amacli
go build -o ./bin/amacli .
```

## 3. Sign in with the browser flow

Recommended first-run setup:

```bash
amacli auth login
```

The CLI will print:

- a browser authorization URL
- a short user code
- the next step to run after authorization

After the user approves in the browser, finish the bootstrap locally:

```bash
amacli auth complete
```

This writes the resulting API key into local `config.json` automatically.

## 4. Where config is stored

By default `amacli` writes to:

```bash
~/.config/amacli/config.json
```

You can override this if needed:

```bash
export AMA_CONFIG_PATH='/custom/path/config.json'
```

## 5. Optional environment variables

```bash
export AMA_BASE_URL='http://localhost:3000'
```

`AMA_API_KEY` is optional now because `amacli` can read the stored key from `config.json` after browser authorization.

## 6. Verify the environment

```bash
amacli health
amacli auth status
amacli me
```

Expected result:

- `auth status` shows `authenticated: true`
- `me` returns the current user, key metadata, and allowed sources

## 7. Smoke-test retrieval

```bash
amacli search --query 'product mvp' --top-k 3
amacli doc lenny 42
```

Replace `42` with a real result id from the search response.

## 8. Smoke-test saving a good answer

```bash
cat answer.md | amacli save-answer \
  --question 'What does Lenny say about PM hiring?' \
  --citations-file citations.json \
  --source lenny
```

Use a `citations.json` array shaped like:

```json
[
  {
    "title": "Great PM hiring",
    "type": "podcast_episode",
    "date": "2025-02-01",
    "source_slug": "lenny",
    "document_path": "/v1/documents/lenny/42",
    "note": "Strongest supporting evidence."
  }
]
```

## 9. Register the skill for another agent runtime

Use one of these patterns depending on the runtime:

### Codex-style local skills

Copy or symlink this folder into the runtime skill directory:

```bash
skills/ama-lenny
```

Make sure the runtime can also access `amacli` on `PATH`.

### Repo-local agent skills

If your runtime reads skills from a project-local `.agents/skills` directory, copy the folder there and keep the same name:

```bash
ama-lenny
```

## 10. Operational rule

This skill should never expose or require db9 credentials or db9 shell commands.

All retrieval and saving must go through:

- `amacli search`
- `amacli document`
- `amacli save-answer`
- the corresponding Ama API endpoints
