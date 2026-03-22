#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$SCRIPT_DIR"
ALL_TARGETS=(
  'darwin/amd64'
  'darwin/arm64'
  'linux/amd64'
  'linux/arm64'
  'windows/amd64'
  'windows/arm64'
)

usage() {
  cat <<'USAGE'
Usage:
  ./build.sh [options]

Options:
  -a, --all                Build for all supported platforms
  --version <version>      Release version to embed, e.g. 0.3.1 or v0.3.1
                           default: dev
  --targets <list>         Comma-separated GOOS/GOARCH pairs
                           default: current platform only
  --output-dir <dir>       Output directory for release binaries
  --tag                    Create an annotated git tag after a successful build
                           requires an explicit --version
  --push-tag               Create and push the annotated git tag
                           requires an explicit --version
  --tag-name <name>        Override the git tag name (default: amacli/v<VERSION>)
  --allow-dirty            Allow tagging with uncommitted changes
  -h, --help               Show this help

Examples:
  ./build.sh
  ./build.sh --all
  ./build.sh --version 0.3.1 --tag --all
  ./build.sh --targets darwin/arm64,linux/amd64 --output-dir ./dist
USAGE
}

require_command() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "Missing required command: $1" >&2
    exit 1
  fi
}

normalize_version() {
  local value="$1"
  value="${value#v}"

  if [[ -z "$value" ]]; then
    echo 'Version cannot be empty' >&2
    exit 1
  fi

  echo "$value"
}

current_target() {
  echo "$(go env GOOS)/$(go env GOARCH)"
}

is_valid_target() {
  local target="$1"
  local candidate

  for candidate in "${ALL_TARGETS[@]}"; do
    if [[ "$candidate" == "$target" ]]; then
      return 0
    fi
  done

  return 1
}

write_checksums() {
  local release_dir="$1"

  (
    cd "$release_dir"
    : > SHA256SUMS
    for artifact in *; do
      if [[ ! -f "$artifact" || "$artifact" == 'SHA256SUMS' ]]; then
        continue
      fi

      if command -v shasum >/dev/null 2>&1; then
        shasum -a 256 "$artifact" >> SHA256SUMS
      else
        sha256sum "$artifact" >> SHA256SUMS
      fi
    done
  )
}

ensure_clean_git_tree() {
  local repo_root="$1"

  if ! git -C "$repo_root" diff --quiet --ignore-submodules HEAD --; then
    echo 'Working tree has uncommitted changes. Commit or pass --allow-dirty before tagging.' >&2
    exit 1
  fi
}

build_target() {
  local version_label="$1"
  local embedded_version="$2"
  local release_dir="$3"
  local target="$4"
  local os_name arch_name ext binary_name output_path

  if [[ "$target" != */* ]]; then
    echo "Invalid target: $target (expected GOOS/GOARCH)" >&2
    exit 1
  fi

  if ! is_valid_target "$target"; then
    echo "Unsupported target: $target" >&2
    echo "Supported targets: ${ALL_TARGETS[*]}" >&2
    exit 1
  fi

  os_name="${target%%/*}"
  arch_name="${target##*/}"
  ext=''
  if [[ "$os_name" == 'windows' ]]; then
    ext='.exe'
  fi

  binary_name="amacli_${version_label}_${os_name}_${arch_name}${ext}"
  output_path="$release_dir/$binary_name"

  echo "-> $target"
  (
    cd "$PROJECT_DIR"
    GOOS="$os_name" \
    GOARCH="$arch_name" \
    CGO_ENABLED=0 \
    go build \
      -trimpath \
      -ldflags="-s -w -X main.version=$embedded_version" \
      -o "$output_path" \
      .
  )

  if [[ "$os_name" != 'windows' ]]; then
    chmod +x "$output_path"
  fi
}

main() {
  require_command go
  if ! command -v shasum >/dev/null 2>&1 && ! command -v sha256sum >/dev/null 2>&1; then
    echo 'Missing checksum tool: shasum or sha256sum' >&2
    exit 1
  fi

  local version='dev'
  local version_label='dev'
  local embedded_version='dev'
  local output_dir tag_name selected_target default_target
  local build_all=0
  local create_tag=0
  local push_tag=0
  local allow_dirty=0
  local has_explicit_version=0
  local -a targets=()

  output_dir="$PROJECT_DIR/dist"
  tag_name=''

  while [[ $# -gt 0 ]]; do
    case "$1" in
      -a|--all)
        build_all=1
        shift
        ;;
      --version)
        [[ $# -ge 2 ]] || { echo '--version requires a value' >&2; exit 1; }
        version="$(normalize_version "$2")"
        version_label="$version"
        embedded_version="v$version"
        has_explicit_version=1
        shift 2
        ;;
      --targets)
        [[ $# -ge 2 ]] || { echo '--targets requires a value' >&2; exit 1; }
        IFS=',' read -r -a targets <<< "$2"
        shift 2
        ;;
      --output-dir)
        [[ $# -ge 2 ]] || { echo '--output-dir requires a value' >&2; exit 1; }
        output_dir="$2"
        shift 2
        ;;
      --tag)
        create_tag=1
        shift
        ;;
      --push-tag)
        create_tag=1
        push_tag=1
        shift
        ;;
      --tag-name)
        [[ $# -ge 2 ]] || { echo '--tag-name requires a value' >&2; exit 1; }
        tag_name="$2"
        shift 2
        ;;
      --allow-dirty)
        allow_dirty=1
        shift
        ;;
      -h|--help)
        usage
        exit 0
        ;;
      *)
        echo "Unknown option: $1" >&2
        usage >&2
        exit 1
        ;;
    esac
  done

  if [[ "$create_tag" -eq 1 && "$has_explicit_version" -ne 1 ]]; then
    echo '--tag and --push-tag require an explicit --version' >&2
    exit 1
  fi

  if [[ -z "$tag_name" && "$has_explicit_version" -eq 1 ]]; then
    tag_name="amacli/v$version"
  fi

  default_target="$(current_target)"
  if [[ ${#targets[@]} -eq 0 ]]; then
    if [[ "$build_all" -eq 1 ]]; then
      targets=("${ALL_TARGETS[@]}")
    else
      targets=("$default_target")
    fi
  fi

  local repo_root='' release_dir
  release_dir="$output_dir/$version_label"

  rm -rf "$release_dir"
  mkdir -p "$release_dir"

  echo "Building amacli $embedded_version"
  if [[ "$build_all" -eq 1 && ${#targets[@]} -eq ${#ALL_TARGETS[@]} ]]; then
    echo 'Mode:    all supported targets'
  elif [[ ${#targets[@]} -eq 1 && "${targets[0]}" == "$default_target" ]]; then
    echo "Mode:    current platform only (${targets[0]})"
  else
    echo "Mode:    custom targets (${targets[*]})"
  fi
  echo "Output:  $release_dir"

  for selected_target in "${targets[@]}"; do
    build_target "$version_label" "$embedded_version" "$release_dir" "$selected_target"
  done

  write_checksums "$release_dir"

  if [[ "$create_tag" -eq 1 ]]; then
    require_command git
    repo_root="$(git -C "$PROJECT_DIR" rev-parse --show-toplevel)"

    if [[ "$allow_dirty" -ne 1 ]]; then
      ensure_clean_git_tree "$repo_root"
    fi

    if git -C "$repo_root" rev-parse "$tag_name" >/dev/null 2>&1; then
      echo "Git tag already exists: $tag_name" >&2
      exit 1
    fi

    git -C "$repo_root" tag -a "$tag_name" -m "Release $tag_name"
    echo "Created git tag: $tag_name"

    if [[ "$push_tag" -eq 1 ]]; then
      git -C "$repo_root" push origin "$tag_name"
      echo "Pushed git tag: $tag_name"
    fi
  fi

  echo 'Done.'
  echo "Binaries are in $release_dir"
}

main "$@"
