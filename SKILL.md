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
   - Assign each piece of evidence a sequential number: `[1]`, `[2]`, `[3]`...
   - Record the UUID for each citation (from search results)
   - For podcasts: also record the exact transcript timestamp
   - Store these extractions — numbers become inline refs, `doc_id` + timestamp become the Citations section
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
Do not stop at source-level attribution. A podcast citation is not complete unless it includes the original English quote (in the answer body) and a timestamp (in the Citations line). Extract the exact transcript timestamp during evidence extraction.

**Hard rule for search breadth:**
Do not prematurely narrow to one source such as Lenny unless the user explicitly asks for that source only. Default behavior is global search across all accessible sources with `top-k 15`, then shortlist the 5 most useful results after reading summaries. Podcast sources should be weighted heavily when the question is about interviews, spoken advice, founder reasoning, or original quotes.

## Voice and tone

Answer like someone who's actually consumed the content — not a neutral summarizer. Use the source's language and framing. Be opinionated when the source is opinionated. Sound human.

**Conclusion-first:** Start with the direct answer, then support with evidence.

## Citation format

During evidence extraction (step 6), assign each piece of evidence a sequential number `[1]`, `[2]`, etc. Record the document UUID (returned by `amacli search`) and, for podcast/video sources, the start timestamp.

**Inline citations** use numbered references only:

```
The best founders obsess over retention before growth [1]. This echoes the idea that "your product is your growth strategy" [2].
```

Rules:
- Always include original English quotes when citing — attribute with the number
- One number per distinct quote or claim. Reuse the same number if citing the same passage again.
- Extract quotes from original markdown before writing answer
- For podcasts, citation quality is judged at the quote level, not just the episode level

## Citations section (MANDATORY — every answer must end with this)

The Citations section is NOT optional. It is the final section of every answer. An answer that ends without it is broken.

**Format:** one line per citation — a clickable link, with optional timestamp for podcast/video.

```
Citations
[1] https://askmeanything.pro/articles/550e8400-e29b-41d4-a716-446655440000?t=154
[2] https://askmeanything.pro/articles/7c9e6679-7425-40de-944b-e07fc1f90ae7
[3] https://askmeanything.pro/articles/f47ac10b-58cc-4372-a567-0e02b2c3d479?t=920
```

- UUID is the document identifier returned by `amacli search`
- For podcasts/videos, append `?t=<seconds>` to jump to the quoted passage
- Omit `?t=` for newsletters and articles
- Do NOT include titles, YouTube links, full quotes, or other metadata — the platform resolves these from the UUID

## Pre-save checklist

Before calling `save-answer`, verify every item. If any answer is "no", fix it first:

1. Did I search globally with `top-k 15` instead of narrowing too early?
2. Did I read summaries and shortlist the 5 strongest sources before opening originals?
3. Did I prioritize podcast sources when the user asked about interviews, spoken advice, or original quotes?
4. For each podcast claim, did I read the original transcript rather than only the search summary?
5. Did I pull at least one original English quote for each strong podcast claim?
6. Did I capture the exact transcript timestamp for each quoted passage?
7. Are inline citations numbered `[1]`, `[2]`, etc. — not the old `[type|title]` format?
8. **Does my answer end with a compact Citations section (`[n] https://askmeanything.pro/articles/<uuid>`)?**
9. Does the saved answer preserve quotes in the body and numbered references — not a citation-light summary?

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
