# ama-skill

[中文说明](./README.zh.md)

`ama-skill` is the standalone agent bundle for [Ask Me Anything](https://askmeanything.pro): a website that turns curated knowledge sources into high-signal answers you can search, inspect, and save.

Today, Ask Me Anything can expose source-backed knowledge from systems like newsletters, podcast transcripts, and other long-form content. An agent can search that knowledge, open the original markdown, answer with evidence, and save strong answers back to the user's private AMA dashboard.

## What Ask Me Anything is

Ask Me Anything is not just a search box.

It combines:

- source retrieval across supported knowledge libraries
- original document access, so an agent can inspect the real markdown instead of relying only on summaries
- user preferences like default source and preferred answer language
- private saved answers, so good outputs can be stored on the website for later reading

In practice, this means an agent can answer questions like:

- “What did Lenny say about MVP scope?”
- “Which guest talked about PM hiring?”
- “Find the original quote and timestamp from that podcast.”
- “Save this answer to my private dashboard.”

## What this repository contains

This repo ships the two pieces you need:

- `amacli/` — a tiny Go CLI for auth, search, document fetch, source selection, language preference, and answer saving
- `skills/ama/` — the installable AMA skill bundle with workflow rules, installation notes, query templates, and API references

Together they let a coding agent or local assistant work against the Ask Me Anything website in a reliable, repeatable way.

## What `ama-skill` can do

With `amacli` + the AMA skill installed, an agent can:

- authenticate to `https://askmeanything.pro`
- search across supported sources
- open original documents before making strong claims
- answer in the user's saved language preference
- respect the user's saved default source
- extract stronger evidence from podcasts, newsletters, and articles
- save standout answers back to the user's private AMA dashboard

The skill is designed for source-grounded answers, not generic hand-wavy summaries.

## Installation

There are two parts to install:

1. install `amacli`
2. install or reference the AMA skill bundle

### 1) Install `amacli`

#### Option A: download a release binary

Download the latest binary from GitHub Releases:

- Releases: `https://github.com/skadai/ama/releases/latest`

Pick the file for your platform, for example:

- `amacli_<version>_darwin_arm64`
- `amacli_<version>_linux_amd64`
- `amacli_<version>_windows_amd64.exe`

Install it into `~/.local/bin/amacli`:

```bash
mkdir -p ~/.local/bin
cp ./amacli_<version>_darwin_arm64 ~/.local/bin/amacli
chmod +x ~/.local/bin/amacli
export PATH="$HOME/.local/bin:$PATH"
amacli version
```

#### Option B: build from source

```bash
cd amacli
./build.sh
```

Useful variants:

```bash
./build.sh --all
./build.sh --version 0.2.1 --all
```

By default:

- the base URL is `https://askmeanything.pro`
- local config is stored at `~/.config/amacli/config.json`

### 2) Install the AMA skill

If your agent supports local `SKILL.md` bundles, install or reference `skills/ama/` from this repo.

At minimum, the key file is:

- `skills/ama/SKILL.md`

If you prefer the website-hosted public skill file, you can also download it directly:

```bash
curl -L https://askmeanything.pro/skill.md -o skill.md
```

For the hosted onboarding instructions:

```bash
curl -L https://askmeanything.pro/install.md -o install.md
```

## First-time onboarding

Once `amacli` is installed, the agent should start browser login immediately:

```bash
amacli auth login
amacli auth complete
```

Then check status and preferences:

```bash
amacli auth status
amacli me
amacli source list
amacli language show
```

Optional preference setup:

```bash
amacli source set-default lenny
amacli language set zh
amacli language set en
```

## Typical workflow

Search first:

```bash
amacli search --query 'What does Lenny say about PM hiring?' --top-k 5
```

Open the original document:

```bash
amacli document --source lenny --id 42
```

Save a strong answer:

```bash
cat answer.md | amacli save-answer \
  --question 'What does Lenny say about PM hiring?'
```

## Repository guide

- product-level skill rules: `skills/ama/SKILL.md`
- install and onboarding guide: `skills/ama/references/INSTALL.md`
- query patterns: `skills/ama/references/query-templates.md`
- API reference: `skills/ama/references/api_reference.md`
- CLI details: `amacli/README.md`

## Who this is for

This repo is useful if you want to:

- plug Ask Me Anything into a local coding agent workflow
- build source-grounded answers on top of the AMA website
- install a reusable agent skill instead of rewriting the retrieval workflow every time
- keep high-quality answers in a private dashboard after they are generated

## License

This project is licensed under the MIT License. See `LICENSE` for details.
