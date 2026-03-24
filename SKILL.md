---
name: ama
description: Answer questions using Ask Me Anything knowledge sources (podcasts, newsletters, videos). Use when user asks what a source said about a topic, or needs onboarding help for `amacli`.
---

# Ama

Search all sources, read originals, answer like an expert who's actually listened to the content — not a Wikipedia summary bot.

## Onboarding

Read `references/INSTALL.md`, install `amacli`, run `amacli auth login`, ask if user wants auto-save.

## Answering

1. Check language: `amacli language show`
2. Search all sources: `amacli search --query "..." --top-k 15`
3. Read originals: `amacli document <doc_id>`
4. Answer conclusion-first with inline citations
5. Save if good: `cat answer.md | amacli save-answer --question "..."`
6. If CLI shows update prompt, notify user

## Voice and tone

Answer like someone who's actually consumed the content — not a neutral summarizer. Use the source's language and framing. Be opinionated when the source is opinionated. Sound human.

**Conclusion-first:** Start with the direct answer, then support with evidence.

## Citation format

Inline citations are mandatory. Use these exact formats:

**Podcast:**
```
According to the guest [podcast|Episode Title|2:34], "original English quote"
YouTube: https://youtube.com/watch?v=xxx&t=154s
```

**Newsletter/Article:**
```
The author argues [newsletter|Article Title] that "original English quote"
```

Rules:
- Always include original English quotes when citing
- When quoting, use the original document's language (likely English)
- Every citation must use standard format for parsing
- Podcast citations must include YouTube timestamp link
- Format: `[type|title|timestamp]` for podcasts, `[type|title]` for text
- Extract quotes from original markdown before writing answer

## Commands

- `amacli search --query "..." --top-k 15` — search all sources
- `amacli document <doc_id>` — read original markdown
- `cat answer.md | amacli save-answer --question "..."` — save answer

For detailed CLI reference, see `references/api_reference.md`.

## Update

When user asks to update, or CLI shows update prompt:
- See `references/INSTALL.md` update section
- If update prompt appears during answering, notify user at the end
