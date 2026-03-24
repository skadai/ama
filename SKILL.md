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
2. Search all sources: `amacli search --query "..." --top-k 5`
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
```

**Newsletter/Article:**
```
The author argues [newsletter|Article Title] that "original English quote"
```

Rules:
- Always include original English quotes when citing
- When quoting, use the original document's language (likely English)
- Every citation must use standard format for parsing
- Format: `[type|title|timestamp]` for podcasts, `[type|title]` for text
- Extract quotes from original markdown before writing answer

**Citations section at end:**

After the answer, include a "Citations" section listing all sources:

```
Citations

• [source/doc_id] Title (podcast, YYYY-MM-DD) — Guest Name
  YouTube: https://youtube.com/watch?v=xxx&t=154s
• [source/doc_id] Title (newsletter, YYYY-MM-DD) — Author Name
```

**YouTube link rules:**
- Only include YouTube link if the document metadata contains a valid `youtube_url`
- Check the original document's frontmatter for `youtube_url` field
- If no `youtube_url` exists, omit the YouTube line entirely
- Never fabricate or guess YouTube URLs

**Examples:**

Podcast with YouTube:
```
• [lenny/abc123] Episode Title (podcast, 2024-01-15) — Guest Name
  YouTube: https://youtube.com/watch?v=xxx&t=154s
```

Podcast without YouTube:
```
• [lenny/abc123] Episode Title (podcast, 2024-01-15) — Guest Name
```

Newsletter:
```
• [newsletter/xyz789] Article Title (newsletter, 2024-01-15) — Author Name
```

## Commands

- `amacli search --query "..." --top-k 5` — search all sources
- `amacli document <doc_id>` — read original markdown
- `cat answer.md | amacli save-answer --question "..."` — save answer

For detailed CLI reference, see `references/api_reference.md`.

## Update

When user asks to update, or CLI shows update prompt:
- See `references/INSTALL.md` update section
- If update prompt appears during answering, notify user at the end
