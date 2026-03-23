# ama-skill

[English README](./README.md)

`ama-skill` 是 [Ask Me Anything](https://askmeanything.pro) 的独立 agent 工具包。Ask Me Anything 是一个把精选知识源变成高质量可追溯回答的网站：你可以搜索内容、打开原始文档、生成带证据的答案，并把好的答案保存回自己的私有 dashboard。

目前 Ask Me Anything 可以承载像 newsletter、podcast transcript、长文文章等知识源。Agent 可以先搜索，再读取原始 markdown，然后基于证据回答，最后把高质量答案收藏到用户自己的 AMA 页面。

## Ask Me Anything 是什么

Ask Me Anything 不只是一个搜索框。

它把这几件事组合在了一起：

- 面向知识库的 source retrieval
- 原始文档访问，让 agent 不只依赖 summary，而是真的去读原始 markdown
- 用户偏好设置，比如默认 source、默认回答语言
- 私有答案收藏，把高质量回答存回网站里，方便以后复查

所以它适合处理这类问题：

- “Lenny 怎么看 MVP scope？”
- “哪位嘉宾聊过 PM hiring？”
- “帮我找到那段 podcast 原文引用和时间戳。”
- “把这条答案存到我的私有 dashboard。”

## 这个仓库里有什么

这个仓库主要包含两部分：

- `amacli/` —— 一个很小的 Go CLI，负责登录、搜索、读取文档、设置 source、设置语言、保存答案
- `skills/ama/` —— 可安装的 AMA skill，包括工作流规则、安装说明、查询模板和 API 参考

这两个部分组合起来，就能让本地 agent 或 coding assistant 更稳定地接入 Ask Me Anything 网站。

## `ama-skill` 能做什么

安装 `amacli` + AMA skill 之后，agent 可以：

- 登录 `https://askmeanything.pro`
- 搜索当前支持的知识源
- 在下结论前先打开原始文档
- 按用户保存的语言偏好输出答案
- 按用户保存的默认 source 工作
- 从 podcast、newsletter、article 里抽取更可靠的证据
- 把值得保留的答案保存到用户私有 AMA dashboard

它的目标不是生成泛泛而谈的总结，而是生成有出处、有原文依据的答案。

## 安装方式

安装分两部分：

1. 安装 `amacli`
2. 安装或引用 AMA skill

### 1）安装 `amacli`

#### 方式 A：直接下载 release 二进制

从 GitHub Releases 下载最新二进制：

- Releases：`https://github.com/skadai/ama/releases/latest`

根据你的平台选择对应文件，例如：

- `amacli_<version>_darwin_arm64`
- `amacli_<version>_linux_amd64`
- `amacli_<version>_windows_amd64.exe`

安装到 `~/.local/bin/amacli`：

```bash
mkdir -p ~/.local/bin
cp ./amacli_<version>_darwin_arm64 ~/.local/bin/amacli
chmod +x ~/.local/bin/amacli
export PATH="$HOME/.local/bin:$PATH"
amacli version
```

#### 方式 B：从源码构建

```bash
cd amacli
./build.sh
```

常见变体：

```bash
./build.sh --all
./build.sh --version 0.2.1 --all
```

默认情况下：

- API base URL 是 `https://askmeanything.pro`
- 本地配置保存在 `~/.config/amacli/config.json`

### 2）安装 AMA skill

先 clone AMA 仓库，这样 agent 才能拿到完整的 `skills/ama/` bundle 以及本地 `references/` 文档：

```bash
git clone https://github.com/skadai/ama.git
```

然后从 clone 下来的仓库里安装或引用 `skills/ama/`。

最关键的入口文件是：

- `skills/ama/SKILL.md`

网站托管的 `skill.md` 只是公开的单文件版本，**不包含**完整的 `references/` 目录，所以不足以支持完整的本地 onboarding。

如果你只想单独下载网站托管的 onboarding 说明，仍然可以：

```bash
curl -L https://askmeanything.pro/install.md -o install.md
```

## 首次 onboarding

一旦 `amacli` 可用，agent 应该立即开始浏览器登录：

```bash
amacli auth login
amacli auth complete
```

然后检查状态和偏好：

```bash
amacli auth status
amacli me
amacli source list
amacli language show
```

可选的偏好设置：

```bash
amacli source set-default lenny
amacli language set zh
amacli language set en
```

## 典型工作流

先搜索：

```bash
amacli search --query 'What does Lenny say about PM hiring?' --top-k 5
```

再打开原始文档：

```bash
amacli document --source lenny --id 42
```

把高质量答案保存下来：

```bash
cat answer.md | amacli save-answer \
  --question 'What does Lenny say about PM hiring?'
```

## 仓库导览

- skill 主规则：`skills/ama/SKILL.md`
- 安装与 onboarding：`skills/ama/references/INSTALL.md`
- 查询模板：`skills/ama/references/query-templates.md`
- API 参考：`skills/ama/references/api_reference.md`
- CLI 详细说明：`amacli/README.md`

## 适合谁用

如果你想做这些事，这个仓库就很适合：

- 把 Ask Me Anything 接到本地 agent 工作流里
- 在 AMA 网站之上做 source-grounded answers
- 安装一个可复用的 agent skill，而不是每次都重写 retrieval 工作流
- 把高质量答案沉淀到自己的私有 dashboard

## License

本项目使用 MIT License，详见 `LICENSE`。
