# amacli

`amacli` 是一个极小的 Go 命令行工具，用来访问当前项目已经提供的 AMA API。

默认服务地址是 `https://askmeanything.pro`。本地开发时可通过 `AMA_BASE_URL` 环境变量覆盖为 `http://localhost:3000`。

## 前置要求

需要 Go `1.25` 或更新版本。

## 安装

```bash
cd amacli
go install .
```

如果你不想写入全局 `GOBIN`，也可以本地构建后放到 `~/.local/bin/amacli`，避免权限问题：

```bash
cd amacli
go build -o ./bin/amacli .
mkdir -p ~/.local/bin
cp ./bin/amacli ~/.local/bin/amacli
chmod +x ~/.local/bin/amacli
export PATH="$HOME/.local/bin:$PATH"
```

## 发布构建

仓库内置了 `build.sh`，可以一次性构建多个平台的裸二进制、生成 `SHA256SUMS`，并且可选创建 git tag：

```bash
cd amacli
./build.sh
./build.sh --all
./build.sh --version 0.3.1 --tag --all
./build.sh --targets darwin/arm64,linux/amd64
```

默认直接只构建当前平台；如果要全平台出包，请显式传 `--all`。

不传 `--version` 时，build 脚本会注入 `dev`；只有显式传了 `--version`，才会写入正式版本号并允许打 tag。

产物不会再压缩，直接输出为可执行二进制文件。

默认会输出到：

```text
amacli/dist/<version-or-dev>/
```

默认目标平台：

- `darwin/amd64`
- `darwin/arm64`
- `linux/amd64`
- `linux/arm64`
- `windows/amd64`
- `windows/arm64`

默认 tag 名称是 `amacli/v<version>`；如果你想自定义，可以传 `--tag-name`。

## 本地配置路径

默认配置写入：

```text
~/.config/amacli/config.json
```

你也可以通过环境变量覆盖：

```bash
export AMA_CONFIG_PATH='/custom/path/config.json'
```

## 浏览器登录

现在推荐直接用浏览器授权来完成首次配置。Agent 在确认 `amacli` 已安装并可执行后，应立即主动运行：

```bash
amacli auth login
```

CLI 会：

1. 发起 device authorization 请求
2. 打印浏览器授权地址和 user code
3. 等用户在网页中批准后，再运行：

```bash
amacli auth complete
```

成功后，`amacli` 会把 `base_url`、`api_key`、用户信息，以及当前默认 `source` 写入本地 `config.json`；`preferred_language` 可以通过 `amacli language set` 单独保存。Skill onboarding 还应该顺手问用户：是否要把优秀回答自动收藏到 AMA 网站，仅自己可见。

你也可以查看当前状态：

```bash
amacli auth status
```

或清理本地登录：

```bash
amacli auth logout
```

## 环境变量

```bash
export AMA_BASE_URL='https://askmeanything.pro'
export AMA_API_KEY='YOUR_API_KEY'
export AMA_CONFIG_PATH='~/.config/amacli/config.json'
export AMA_HTTP_TIMEOUT='1m'
```

说明：

- `AMA_BASE_URL` 可选，默认就是 `https://askmeanything.pro`
- `AMA_API_KEY` 可选；如果你已经跑过 `amacli auth complete`，CLI 会自动从本地 `config.json` 读取
- `AMA_CONFIG_PATH` 可选；默认写入 `~/.config/amacli/config.json`
- `AMA_HTTP_TIMEOUT` 可选；默认是 `1m`，本地 dev server 慢的时候建议调大，比如 `90s`

## 用法

```bash
amacli health
amacli auth login
amacli auth complete
amacli me
amacli source list
amacli source set-default lenny
amacli --timeout 90s search --query 'How does Lenny think about MVP scope?'
amacli search --balanced-content-types --query 'What does Lenny say about PM hiring?' --top-k 6
amacli document --id 42
amacli doc 42
cat answer.md | amacli save-answer --question 'What does Lenny say about PM hiring?'
```

## `source`

查看当前 API key 可访问的 source 列表，并把默认 source 写入本地配置：

```bash
amacli source list
amacli source set-default lenny
```

设置后，`search`、`document`、`doc`、`save-answer` 这些命令在你不显式传 `--source` 时，都会优先使用 `~/.config/amacli/config.json` 里的 `default_source`。

## `language`

设置默认回答语言，让 skill / agent 优先按这个语言输出：

```bash
amacli language show
amacli language set zh
amacli language set en
```

当前只支持：

- `zh`
- `en`

设置后，agent 在没有额外指示时，应优先遵循 `~/.config/amacli/config.json` 里的 `preferred_language`。

## `search` balanced by content type

When one format tends to dominate the results, use balanced search to issue one query per content type and merge the shortlist so newsletters and podcasts both get representation:

```bash
amacli search --balanced-content-types \
  --query 'What does Lenny say about PM hiring?' \
  --top-k 6
```

Notes:

- default balanced types are `newsletter_article` and `podcast_episode`
- if you also pass `--content-type`, balanced search uses those types instead
- balanced search makes multiple API requests under the hood, one per content type

## `save-answer`

如果某次 skill / CLI 回答让用户满意，可以把它存进 dashboard：

```bash
cat answer.md | amacli save-answer \
  --question 'What does Lenny say about PM hiring?'
```

如果你想最省事，直接让答案末尾带一个可解析的引用区即可，后端会自动抽取结构化 citations，例如：

```md
参考来源：
- newsletter | title: What to ask your users about Product-Market Fit | date: 2020-06-02 | source: lenny | document: /v1/documents/lenny/681
- podcast | title: Pricing your AI product | date: 2025-07-27 | guest: Madhavan Ramanujam | source: lenny | url: https://www.youtube.com/watch?v=abc&t=123 | timestamp: 02:03 | note: pricing is a demand test
```

当然你仍然可以显式传 `citations.json`。如果传了，后端会优先使用显式 citations。

`citations.json` 的结构化格式仍然长这样：

```json
[
  {
    "title": "Great PM hiring",
    "type": "podcast_episode",
    "date": "2025-02-01",
    "guest": "Some Guest",
    "source_slug": "lenny",
    "document_path": "/v1/documents/lenny/42",
    "url": "https://www.youtube.com/watch?v=abc&t=123",
    "timestamp": "02:03",
    "note": "最强证据段落。"
  }
]
```

保存成功后，用户可以在 `/dashboard` 页面看到自己的瀑布流收藏卡片。API 响应还会返回 `saved_answer.view_url`，可直接附在回答末尾，引导用户去专属详情页阅读。

除 `auth` 命令以外，其余命令默认输出格式化 JSON，便于 shell、agent、`jq`、以及 skill 工作流直接消费。
