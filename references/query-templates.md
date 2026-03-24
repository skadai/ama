# Ama query templates

These patterns use `amacli` and the public Ask Me Anything API.

## 1. Connectivity check

```bash
amacli health
amacli auth status
amacli me
amacli language show
amacli source list
```

## 2. Primary search

Use this first.

```bash
amacli search --query "How does this source think about MVP scope and product validation?" --top-k 5
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

## 5. Balanced search by content type

Use this when one format keeps crowding out the other and you want newsletters plus podcasts both represented in the shortlist.

```bash
amacli search --balanced-content-types \
  --query "What does Lenny say about PM hiring?" \
  --top-k 6
```

## 6. Fetch original documents

Once you have strong candidates, open the original markdown before answering.

```bash
amacli document --source lenny --id 42
```

Shorthand:

```bash
amacli doc lenny 42
```

## 7. Working merge strategy

After running more than one search:
- merge on `id`
- preserve `title`, `type`, `date`, `guest`, `summary`, `document_path`, `score`, and `source_slug`
- shortlist the 3-5 strongest results before opening originals

## 8. Working answer checklist

Before you write the final answer:
1. confirm the strongest result is actually relevant
2. open the top original document(s)
3. extract the best supporting reasoning or quote
4. note timestamps and links when the source is a transcript or video
5. structure the answer as direct answer, source context, evidence, synthesis, then citations

## 9. Default answer pattern

Use this pattern unless the user wants a different format:
- direct answer first
- source and context second
- key quotes or close paraphrases third
- synthesis across sources fourth
- citations list last

## 10. Save approved answers when enabled

Use this only if the user opted into private auto-save, or explicitly asks to keep the answer.

```bash
cat answer.md | amacli save-answer \
  --question "What does this source say about PM hiring?" \
  --citations-file citations.json
```

After saving:
- read `saved_answer.view_url` from the JSON response
- append one final line pointing the user to that private page
- if `view_url` is missing, build `https://askmeanything.pro/dashboard/answers/<id>` from `saved_answer.id`
