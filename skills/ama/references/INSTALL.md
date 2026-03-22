# Ama install and onboarding

Use this once per machine.

## 1. Download `amacli`

Download the latest binary from GitHub Releases:

- Releases page: `https://github.com/skadai/ama/releases/latest`
- Pick the file matching your platform, for example:
  - `amacli_<version>_darwin_arm64`
  - `amacli_<version>_linux_amd64`
  - `amacli_<version>_windows_amd64.exe`

Install it somewhere on `PATH`, for example on macOS / Linux:

```bash
chmod +x ./amacli_<version>_darwin_arm64
mv ./amacli_<version>_darwin_arm64 /usr/local/bin/amacli
amacli version
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

## 2. Download the skill file

Download the current public skill markdown:

```bash
curl -L https://askmeanything.pro/skill.md -o skill.md
```

If you also want the onboarding guide locally:

```bash
curl -L https://askmeanything.pro/install.md -o install.md
```

## 3. Log in with the browser flow

```bash
amacli auth login
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

## 6. First question checklist

After onboarding, this minimal flow should work:

```bash
amacli health
amacli auth status
amacli language show
amacli source list
amacli search --query "What does Lenny say about MVP scope?" --top-k 5
```

## 7. What the normal runtime should remember

The local config stores:
- API key and account info from browser login
- default source
- preferred language

The skill should:
- answer in the saved language unless the user overrides it
- use the saved default source unless the user overrides it
- open original documents before making strong claims
