# Ama install and onboarding

Use this once per machine.

## 1. Download `amacli`

Download the latest binary from GitHub Releases:

- Releases page: `https://github.com/skadai/ama/releases/latest`
- Pick the file matching your platform, for example:
  - `amacli_<version>_darwin_arm64`
  - `amacli_<version>_linux_amd64`
  - `amacli_<version>_windows_amd64.exe`

Install it to `~/.local/bin/amacli` so you do not need `sudo`.

For macOS / Linux:

```bash
mkdir -p ~/.local/bin
cp ./amacli_<version>_darwin_arm64 ~/.local/bin/amacli
chmod +x ~/.local/bin/amacli
export PATH="$HOME/.local/bin:$PATH"
amacli version
```

If `~/.local/bin` is not already on `PATH`, add this to your shell profile:

```bash
export PATH="$HOME/.local/bin:$PATH"
```

If you build from source instead:

```bash
cd ama-skill/amacli
./build.sh
```

Default behavior:
- `./build.sh` builds only the current platform
- `./build.sh --all` builds every supported platform
- without `--version`, the binary version is `dev`
- default API base URL is `https://askmeanything.pro`

## 2. Download the skill file

Download the current public skill markdown:

```bash
curl -L https://askmeanything.pro/skill.md -o skill.md
```

If you also want the onboarding guide locally:

```bash
curl -L https://askmeanything.pro/install.md -o install.md
```

## 3. Start browser login immediately

As soon as `amacli` is available on `PATH`, the agent should run `amacli auth login` automatically.

Do **not** stop and wait for the user to manually type `amacli auth login`.

```bash
amacli auth login
```

After the user finishes browser approval, complete the flow:

```bash
amacli auth complete
```

Useful checks:

```bash
amacli auth status
amacli me
```

## 4. Set answer language

Pick one:

```bash
amacli language set zh
amacli language set en
```

Check it:

```bash
amacli language show
```

This preference decides the default answer language for the skill.

## 5. Set your default source

See available sources:

```bash
amacli source list
```

If the public source you want is `lenny`, save it as default:

```bash
amacli source set-default lenny
```

## 6. Ask about private auto-save

During onboarding, ask the user:
- do you want me to automatically save strong answers to your AMA website account, visible only to you?

Rules:
- if yes, after each strong final answer call `amacli save-answer` automatically
- if no, do not save unless they ask
- after a successful save, read `saved_answer.view_url` from the response and append that private link at the end of the answer

## 7. First question checklist

After onboarding, this minimal flow should work:

```bash
amacli health
amacli auth status
amacli language show
amacli source list
amacli search --query "What does Lenny say about MVP scope?" --top-k 5
```

## 8. What the normal runtime should remember

The local config stores:
- API key and account info from browser login
- default source
- preferred language
- base URL, which defaults to `https://askmeanything.pro`

The skill should:
- trigger `amacli auth login` automatically once installation is complete
- answer in the saved language unless the user overrides it
- use the saved default source unless the user overrides it
- ask whether private auto-save should be enabled during onboarding
- if auto-save is enabled, save strong final answers and append the private `view_url` at the end
- open original documents before making strong claims
