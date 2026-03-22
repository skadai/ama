# Query templates for Ama Lenny

These patterns use `amacli` plus the public Ama API. Do **not** use db9 commands in this workflow.

## 1. Connectivity check

```bash
amacli health
amacli me
```

## 2. Primary search

Use this first.

```bash
amacli search --query "How does Lenny think about MVP scope and product validation?" --top-k 5
```

## 3. Follow-up searches with close phrasings

Use this when the first search is weak, ambiguous, or too narrow.

```bash
amacli search --query "minimum viable product" --top-k 5
amacli search --query "product validation" --top-k 5
amacli search --query "prototype before scaling" --top-k 5
```

## 4. Narrow by content type

```bash
amacli search --query "MVP scope" --content-type podcast_episode --top-k 5
amacli search --query "MVP scope" --content-type newsletter_article --top-k 5
```

## 5. Fetch original documents

Once you have strong candidates, fetch the original markdown before answering.

```bash
amacli document --source lenny --id 42
```

Shorthand:

```bash
amacli doc lenny 42
```

## 6. Outside-command merge strategy

After running more than one search:

- merge on `id`
- preserve:
  - `title`
  - `type`
  - `date`
  - `guest`
  - `summary`
  - `document_path`
  - `source_path`
  - `score`
  - `recall_sources`
- shortlist the 3-5 strongest results before opening originals

## 7. Working answer checklist

Before you write the polished answer:

1. confirm the strongest result is actually relevant
2. open the top original document(s)
3. extract the best supporting reasoning or quote
4. note timestamps and `youtube_url` when the document is a podcast transcript
5. synthesize first, cite second
