---
name: ama
description: Search and answer questions over Ask Me Anything knowledge sources through `amacli`. Use when the user asks what a supported source said about a topic, wants evidence-backed summaries from podcast/newsletter/video content, needs original markdown inspection for top search hits, or needs installation and onboarding help for `amacli`, `skill.md`, browser login, source selection, and language preference. The current public source may be `lenny`, but this skill stays source-agnostic.
---

# Ama

Use `amacli` to search the Ask Me Anything knowledge layer, inspect original markdown, and answer in a way that reflects what the source actually said — with evidence.

Current public source availability may be narrow, but the skill name stays generic on purpose. Treat source choice as configuration, not as the skill identity.

## Quick start

1. For first-time install or onboarding, read `references/INSTALL.md`.
2. During onboarding, once `amacli` is installed and on `PATH`, trigger `amacli auth login` automatically. Do not wait for the user to type it manually.
3. During onboarding, ask whether the user wants strong answers to be auto-saved to their private AMA website account (visible only to them).
4. Before answering in a fresh session, check:
   - `amacli health`
   - `amacli auth status`
   - `amacli me`
   - `amacli language show`
   - `amacli source list`
5. Prefer the saved `preferred_language` for answer language.
6. Prefer the saved `default_source` unless the user explicitly asks for another source.
7. Start with one focused `amacli search`, then open the strongest originals with `amacli document`.

## How the knowledge layer works

The workflow has two levels:

- `amacli search` gives you the shortlist: titles, summaries, snippets, dates, guests, and document pointers.
- `amacli document` gives you the original markdown, which is the authoritative layer for nuanced answers.

Do not stop at search summaries when the user is asking a substantive question. Use search to find the right sources, then read the original markdown before making strong claims.

## Answering workflow

### Step 1: Honor language and source defaults

Run:

```bash
amacli language show
amacli source list
```

Rules:
- if `preferred_language` is `zh`, answer in Chinese unless the user clearly asks otherwise
- if `preferred_language` is `en`, answer in English unless the user clearly asks otherwise
- if no language preference is saved, mirror the user's current language
- use the saved default source when the user does not specify one
- if the current public source is only `lenny`, that is fine; do not rename the skill because of one source
- during first-time onboarding, once `amacli` is available, immediately run `amacli auth login`
- during onboarding, explicitly ask whether the user wants auto-save for good answers into their private AMA dashboard

### Step 2: Search for relevant content

Start with the user's exact wording first.

```bash
amacli search --query "How does this source think about MVP scope?" --top-k 5
```

If the first pass is weak, expand into 3-6 nearby phrasings. Keep the expansions tight.

Examples:
- `mvp scope`
- `minimum viable product`
- `product validation`
- `prototype before scaling`

When the user asks about a specific guest or person:
- search their name directly
- inspect titles, guests, and summaries in the results

When the user asks about a topic:
- search broadly across close phrasings
- prefer 3-5 strong candidates over a wide noisy list
- when podcasts dominate and newsletters disappear, or vice versa, use balanced per-type search so both formats stay represented

When the question clearly points to one format, narrow by content type:

```bash
amacli search --query "MVP scope" --content-type podcast_episode --top-k 5
amacli search --query "MVP scope" --content-type newsletter_article --top-k 5
amacli search --balanced-content-types --query "PM hiring" --top-k 6
```

### Step 3: Read the original sources carefully

Once you identify the strongest 1-3 candidates, fetch the original markdown.

```bash
amacli document --source lenny --id 42
```

Shorthand:

```bash
amacli doc lenny 42
```

This is critical. The search result is just the pointer. The real evidence is in the original document.

For podcast transcripts, pay attention to:
- the guest's actual words and argument structure
- concrete examples, stories, and data points
- the back-and-forth between host and guest
- timestamps, links, or transcript markers when present

For newsletters and articles, pay attention to:
- the author's own frameworks and opinions
- examples, data, and case studies
- actionable advice and decision rules

Preferred reading order:
1. read the top result in full if it is short enough
2. read the most relevant sections first if it is long
3. open 1-2 additional sources when the topic clearly spans multiple documents
4. extract the strongest supporting passages before writing the final answer

### Step 4: Compose the answer

Your answer should feel like it comes from someone who actually read the source material, not from someone who only skimmed metadata.

Default answer structure:

1. **Direct answer**
   - lead with what the source actually said about the topic
   - use the source's ideas, not generic filler advice

2. **Source and context**
   - tell the user exactly where this came from
   - for podcasts: mention the guest and date when available
   - for newsletters/articles: mention the title and date when available

3. **Key quotes or close paraphrases**
   - pull specific supporting language, reasoning, or examples
   - paraphrase carefully when that reads better, but stay faithful to the source
   - keep timestamps or links when they materially improve traceability
   - when quoting directly from English-language sources, always include the original English text
   - you may add a Chinese translation alongside, but the original English must be present

4. **Synthesis across sources**
   - if multiple sources are relevant, weave them together
   - note whether they agree, disagree, or build on each other
   - separate the source's view from your synthesis

5. **Citations**
   - always end the answer with a citation list
   - list every source you relied on, even if it was only quoted once
   - include document id, title, type, date, and guest when available
   - preferred format: `[source_slug/id] Title (type, date) — Guest`
   - example: `[lenny/700] How to know if you've got product-market fit (newsletter_article, 2020-01-28)`

6. **Private save link when enabled**
   - if the user opted into auto-save during onboarding, call `amacli save-answer` after you finish the final answer
   - only auto-save final answers that are strong enough to keep; never save drafts, partial work, or low-confidence replies
   - prefer passing structured citations with `--citations-file`; otherwise make sure the answer ends with a parseable citations section
   - read `saved_answer.view_url` from the command response
   - if `view_url` is missing, fall back to `saved_answer.id` and build `https://askmeanything.pro/dashboard/answers/<id>`
   - after saving, add one final line at the end of the answer telling the user they can read it more comfortably on that private page

### Step 5: Use deep reading when needed

If the best evidence lives in long transcripts or long-form articles:
- shortlist documents first
- inspect the most relevant sections before broadening out
- avoid dumping long raw source text into the answer
- return a compact evidence package instead of a transcript recap

## Response language

Match the user's language unless a saved preference overrides it.

Rules:
- Chinese user + no explicit override -> respond in Chinese
- English user + no explicit override -> respond in English
- when responding in Chinese about English-language sources, keep short key terms or short original phrases in English only when they improve accuracy
- when quoting directly from an English-language source, always include the original English quote verbatim
- if helpful, add a Chinese translation after it, but do not omit the original English
- keep the main explanation in the chosen answer language

## What makes a great answer

- specificity over generality
- clear source attribution
- evidence from original documents, not only search summaries
- strong synthesis when multiple sources matter
- honesty about gaps when the source does not actually cover the topic

Examples of good behavior:
- “In Lenny's interview with Brian Halligan on 2025-02-01...”
- “In the newsletter '[Title]' ([date]), the main argument is...”
- “Across these two sources, the shared idea is..., but they differ on...”

## Common question patterns

- "How does X view Y?"
  - If the source has first-person essays or newsletters, check these first
  - Then use interviews or podcasts to deepen or contrast the view

- "Who has discussed Y?"
  - Search broadly using closely related phrasings
  - Gather multiple relevant guests or articles, then synthesize

- "What did guest X say?"
  - Search directly for the guest name
  - Open the transcript before drafting the answer

- "On which day did they talk about Y?"
  - Search for the topic, then use the result metadata and the original document to confirm the date

- "What is the best advice about Y?"
  - Search across both interviews and written pieces when available
  - Synthesize recurring themes and highlight the strongest, source-backed recommendations


## Save standout answers

During onboarding, ask once: do you want me to automatically save especially good answers to your private AMA dashboard, visible only to you?

Rules:
- if the user says yes, auto-save good final answers after you finish them
- if the user says no, do not auto-save unless they explicitly ask
- if the preference is still unknown, ask before the first save-worthy answer

Command pattern:

```bash
cat answer.md | amacli save-answer \
  --question "What does this source say about PM hiring?"
```

After saving:
- parse `saved_answer.view_url` from the response
- append a final sentence such as:
  - Chinese: `你也可以在这个仅自己可见的专属页面里更舒服地查看这条收藏：https://askmeanything.pro/dashboard/answers/<id>`
  - English: `You can also read this saved answer on your private AMA page: https://askmeanything.pro/dashboard/answers/<id>`

## References

Load only what you need:
- install and onboarding: `references/INSTALL.md`
- search patterns: `references/query-templates.md`
- CLI and API details: `references/api_reference.md`
