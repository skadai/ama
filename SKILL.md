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
2. Search broadly first: `amacli search --query "..." --top-k 15`
3. Read the search summaries carefully before opening originals
4. Choose the 5 most useful sources from the search results, prioritizing podcast sources when they are relevant and high-signal
5. Read originals for those selected sources: `amacli document <doc_id>`
6. **Extract evidence BEFORE writing the answer** (this is not optional):
   - For each source you will cite, pull the original English quote now
   - For podcasts: exact transcript timestamp + quote
   - If document metadata contains `youtube_url`: build the `t=` jump link now
   - Store these extractions — they become both inline citations AND the final Citations section
7. Write the answer conclusion-first with inline citations
8. **Append the Citations section** — this is the last part of every answer, no exceptions
9. Self-check: re-read your response and verify the Citations section exists and is complete
10. Save if good: `cat answer.md | amacli save-answer --question "..."`
11. If CLI shows update prompt, notify user

**Hard rule — Citations section is MANDATORY:**
- Every answer shown to the user MUST end with a Citations section — regardless of whether `save-answer` is called
- NEVER output an answer without a Citations section at the end
- NEVER call `save-answer` on an answer that is missing the Citations section
- If you realize you forgot Citations, stop and append it before doing anything else
- An answer without Citations is an incomplete answer — treat it as a bug

**Hard rule for podcasts:**
Do not stop at source-level attribution. A podcast citation is not complete unless it includes the original English quote and a timestamp. If `youtube_url` exists in metadata, include the timestamped YouTube link in both the final answer and the saved answer. If no `youtube_url` exists, still include the timestamp and explicitly omit the link rather than inventing one.

**Hard rule for search breadth:**
Do not prematurely narrow to one source such as Lenny unless the user explicitly asks for that source only. Default behavior is global search across all accessible sources with `top-k 15`, then shortlist the 5 most useful results after reading summaries. Podcast sources should be weighted heavily when the question is about interviews, spoken advice, founder reasoning, or original quotes.

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
- For podcasts, citation quality is judged at the quote level, not just the episode level
- For podcasts, include timestamp inline every time you quote
- If document metadata contains `youtube_url`, convert the quote timestamp into a `t=` jump link and include it
- Never call `save-answer` with a podcast answer that dropped quote timestamps or YouTube jump links that were available during extraction

## Citations section (MANDATORY — every answer must end with this)

The Citations section is NOT optional. It is the final section of every answer. An answer that ends without it is broken.

Structure:

```
Citations

• [podcast|Episode Title|2:34]
  Quote: "original English quote"
  YouTube: https://youtube.com/watch?v=xxx&t=154s

• [podcast|Episode Title|15:20]
  Quote: "another original English quote"

• [newsletter|Article Title]
  Quote: "original English quote"
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

## Pre-save checklist

Before calling `save-answer`, verify every item. If any answer is "no", fix it first:

1. Did I search globally with `top-k 15` instead of narrowing too early?
2. Did I read summaries and shortlist the 5 strongest sources before opening originals?
3. Did I prioritize podcast sources when the user asked about interviews, spoken advice, or original quotes?
4. For each podcast claim, did I read the original transcript rather than only the search summary?
5. Did I pull at least one original English quote for each strong podcast claim?
6. Did I capture the exact transcript timestamp for each quoted passage?
7. Did I check whether document metadata includes `youtube_url`?
8. If `youtube_url` exists, did I add a `t=` jump link everywhere that quote appears?
9. **Does my answer end with a complete Citations section (Evidence + Sources)?**
10. Does the saved answer preserve quotes, timestamps, and links — not a citation-light summary?

If any answer is "no", the workflow is incomplete. Do not call `save-answer`.

## Commands

- `amacli search --query "..." --top-k 15` — search all sources broadly first
- `amacli document <doc_id>` — read original markdown
- `cat answer.md | amacli save-answer --question "..."` — save answer

For detailed CLI reference, see `references/api_reference.md`.

## Update

When user asks to update, or CLI shows update prompt:
- See `references/INSTALL.md` update section
- If update prompt appears during answering, notify user at the end
