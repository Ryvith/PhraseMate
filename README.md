# PhraseMate

感觉市面上当前已经有的单词本很难用，很多短语都无法记录，于是制作一个简单的单词本小程序，可以接入 API，在边看 B 站网课的时候边听边记录。

本文以 **macOS 本地使用与开发**为主。技术栈：**Go + Cocoa / WKWebView 原生窗口 + HTML / CSS / JavaScript + SQLite + OpenAI 兼容 API**。

网页资源直接嵌入 Go 程序，不需要单独安装 Node.js 或构建前端。桌面模式会启动仅监听本机的 HTTP 服务，再由 WKWebView 显示界面，不需要部署服务器或登录账号。

## 功能

- **原生桌面窗口**：生词本、写作本和独立范文阅读窗口。
- **置顶速记窗**：280 × 66 像素的柔和灰白速记栏，快速收录单词和短语；按住空白区域拖动，右键（或 Control-click）选择「隐藏速记窗」，再从顶部菜单栏选择「显示速记窗」恢复，本次运行期间保留拖动位置。
- **macOS 编辑快捷键**：支持 Command+C / V / X / A 复制、粘贴、剪切和全选，以及 Command+Z / Shift+Command+Z 撤销和重做。
- **图形化 API 配置**：在「设置」里填写 Key、接口地址和模型。
- **AI 双语释义**：中文与英文释义、音标、词性、例句、派生词和常用搭配。
- **词典优先**：单词先查免费词典，必要时使用 AI 补全。
- **生词本与自测**：去重更新、筛选、删除和选择题自测。
- **本地保存**：SQLite 保存生词和应用配置，可选同步生词到 Git 分支。

## macOS：从源码快速运行

### 1. 准备开发环境

需要安装 [Go](https://go.dev/dl/) 和 Xcode Command Line Tools。Go 版本要求见 [go.mod](go.mod)。这些工具仅开发和打包时需要，使用构建好的 App 不需要安装。

检查环境：

```bash
go version
xcode-select -p
uname -m
```

- `go version`：确认 Go 已安装。
- `xcode-select -p`：确认 Command Line Tools 或 Xcode 开发工具路径已配置。
- `uname -m`：M 系列芯片通常显示 `arm64`，Intel 芯片显示 `x86_64`。

缺少 Command Line Tools 时执行，并按系统提示完成安装：

```bash
xcode-select --install
```

原生窗口依赖 Cocoa / WebKit，需要在 Mac 上使用支持 CGO 的工具链构建。

### 2. 进入项目目录

进入包含 `go.mod` 的项目根目录。例如，仓库位于 `~/Workspaces/PhraseMate` 时：

```bash
cd ~/Workspaces/PhraseMate
```

如果克隆到了其他位置，请替换为自己的实际路径。以下命令都在项目根目录执行。

### 3. 直接运行

```bash
PHRASEMATE_GIT_SYNC=0 go run .
```

- `go run .` 会编译当前项目，然后运行临时程序；首次运行会自动下载缺少的 Go 依赖。
- `PHRASEMATE_GIT_SYNC=0` 仅对本次运行关闭 Git 生词同步，避免试用时自动创建、提交和推送 `data` 分支。
- 成功后会打开图形界面，终端中会显示本地服务地址和运行日志。
- `go run` 使用临时程序，不会自动创建桌面别名。

关闭主窗口只是隐藏到菜单栏。完全退出请点击顶部菜单栏的 PhraseMate 图标，选择「退出 PhraseMate」。修改代码后，退出程序，再执行上述命令查看效果。

### 4. 在界面中配置 AI

打开右上角「设置」，填写：

| 配置 | 含义 |
|------|------|
| API Key | 你在模型服务商平台创建的密钥 |
| Base URL | 同一家服务商提供的、兼容 Chat Completions 的 API 基础地址 |
| 模型 | 该服务商支持的准确模型 ID |

当前客户端会向 `Base URL + /chat/completions` 发送请求，并请求 JSON 对象格式的响应，请选择支持这些能力的服务和模型。设置面板提供服务商预填选项，使用前仍需确认地址和模型有效。

保存后立即生效，重启后保留；Key 输入框留空表示保持已保存的 Key，需要删除时使用「清除 API Key」。不需要把 Key 写入源码，也不需要创建 `.env`。

**没有 Key 仍可收录生词，但不保证获得完整释义。** 词典未命中时只会保存基础占位提示；中文释义、短语解释、派生词和翻译等功能可能需要 AI。AI 请求由本机直接发送到所配置的服务商，用量计入对应账户。给其他人分发 App 时，每人填写自己的 Key，不要内置或分享你的密钥。

当前 Key 保存在本地 SQLite 中，尚未使用 macOS Keychain。不要公开数据库、含密钥的 `.env` 或把它们放进安装包。

## macOS：构建和打包 App

开发调试使用 `go run`；需要保留可执行程序或生成 `.app` 时，按以下顺序操作。

### 1. 构建可执行程序

```bash
CGO_ENABLED=1 go build -o dist/PhraseMate .
```

- `CGO_ENABLED=1`：启用 Go 与原生 Cocoa / WebKit 代码的编译连接。
- `-o dist/PhraseMate`：将生成的程序保存到指定位置。
- `.`：编译当前目录的项目。

默认构建当前机器对应的架构版本：M 系列 Mac 为 `arm64`，Intel Mac 为 `amd64`。这一步不会生成通用二进制。

可以先直接运行构建结果：

```bash
PHRASEMATE_GIT_SYNC=0 ./dist/PhraseMate
```

### 2. 生成应用包

```bash
./dist/PhraseMate --pack-app dist/PhraseMate.app
```

`--pack-app` 是本项目实现的打包命令，不是 Go 自带参数。它会复制当前可执行程序、生成应用信息和图标：

```text
dist/PhraseMate.app/
└── Contents/
    ├── Info.plist
    ├── MacOS/
    │   └── PhraseMate
    └── Resources/
```

打包命令只生成应用包，不会启动单词本服务。

### 3. 本机临时签名

```bash
codesign --force --sign - dist/PhraseMate.app
```

`--sign -` 表示 ad hoc 签名，适合本地试用，不包含 Developer ID 身份，也不等于 Apple 公证。面向其他同学正式分发时，需要另外完善签名、公证和兼容性验证。

### 4. 检查构建结果

```bash
file dist/PhraseMate.app/Contents/MacOS/PhraseMate
vtool -show-build dist/PhraseMate.app/Contents/MacOS/PhraseMate
codesign --verify --verbose=2 dist/PhraseMate.app
```

- `file`：检查架构，M 系列版本应包含 `arm64`。
- `vtool`：查看二进制的最低 macOS 版本（`minos`）和 SDK 信息。
- `codesign`：验证应用签名是否完整。

应用的 `Info.plist` 当前声明最低 macOS 11，但构建脚本尚未统一指定编译时的最低系统版本，不能仅凭该声明判断兼容性。实际要求还需检查二进制并在目标系统测试。例如，本次开发机生成的 arm64 程序显示 `minos 15.0`，不代表其他环境构建的产物具有相同要求。

### 5. 启动应用

首次试用推荐从终端运行应用包里的程序，关闭 Git 同步并保留日志：

```bash
PHRASEMATE_GIT_SYNC=0 ./dist/PhraseMate.app/Contents/MacOS/PhraseMate
```

确认正常后，可以在 Finder 中打开 `dist`，将 `PhraseMate.app` 拖到「应用程序」再打开。**上面命令中的环境变量不会随应用包保存，Finder 启动不会自动继承这次设置。** Git 同步的触发条件见下一节。

重新修改代码后，需要依次重新构建、打包和签名，原有 `.app` 才会包含新代码。正常重新打包不会清空独立保存的本地数据库。

### 常用命令速查

开发调试：

```bash
PHRASEMATE_GIT_SYNC=0 go run .
```

构建、打包和签名（按顺序执行，每一步成功后再继续）：

```bash
CGO_ENABLED=1 go build -o dist/PhraseMate .
./dist/PhraseMate --pack-app dist/PhraseMate.app
codesign --force --sign - dist/PhraseMate.app
```

## 本地数据与 Git 同步

macOS 默认将生词和 API 配置保存在：

```text
~/Library/Application Support/PhraseMate/phrasemate.db
```

`PHRASEMATE_DB` 可覆盖这个路径。备份时先完全退出 App，再复制数据库；数据库含 API Key，请妥善保管。

**添加和保存生词不需要 GitHub。** GitHub 用于托管项目源码、分发版本，以及可选的生词同步。

当前代码默认尝试发现工作目录或程序所在位置的 Git 仓库。发现仓库且有可用的 `origin` 时，会将生词快照写入 `data` 分支的 `words.json`，并尝试自动提交和推送；API 配置不写入该分支。未发现可用仓库时，仅使用本地数据库。

只想本地使用时，运行命令中加上 `PHRASEMATE_GIT_SYNC=0`。需要在本机源码调试时持续关闭，也可以在项目根目录创建 `.env`，只写入：

```dotenv
PHRASEMATE_GIT_SYNC=0
```

`.env` 已被 Git 忽略，但不会自动包含在打包后的 App 中。程序优先查找可执行文件旁的 `.env`，其次查找工作目录中的 `.env`。

其他高级配置见 [.env.example](.env.example)。如果复制整个示例文件，注意它包含 `PHRASEMATE_DB=data/phrasemate.db`，会把数据库位置改为相对工作目录的路径。日常使用推荐在界面配置 AI，并保留默认数据库位置。

## 浏览器模式

只调试网页界面，不启动原生窗口：

```bash
PHRASEMATE_GIT_SYNC=0 PHRASEMATE_WEB=1 PHRASEMATE_ADDR=127.0.0.1:8080 go run .
```

浏览器打开终端打印的地址，默认是 `http://127.0.0.1:8080`。浏览器模式没有菜单栏图标和系统置顶速记窗，停止服务可在终端按 `Control+C`。

## Docker（网页实例）

Docker 只提供网页服务，不提供原生窗口、菜单栏图标或系统置顶速记窗。安装 Docker 后，在项目根目录执行：

```bash
docker compose up -d --build
```

浏览器打开 [http://localhost:8080](http://localhost:8080)，在「设置」中配置 AI。Compose 默认关闭 Git 同步，并将数据库保存在 `phrasemate-data` 数据卷中。

也可以使用原作者的镜像（不是本 fork 修改后自动生成的版本）：

```bash
docker pull ghcr.io/lucas2029460802/phrasemate:latest
docker run --name phrasemate -d -p 127.0.0.1:8080:8080 -v phrasemate-data:/data ghcr.io/lucas2029460802/phrasemate:latest
```

管理容器：

```bash
docker logs -f phrasemate
docker stop phrasemate
docker start phrasemate
```

删除容器通常不会删除命名数据卷；删除 `phrasemate-data` 数据卷会丢失其中的生词和设置。Compose 创建的实际卷名可能带有项目名前缀。

如果已有导出的镜像文件，可以离线导入后运行：

```bash
docker load -i phrasemate-docker.tar.gz
docker run --name phrasemate -d -p 127.0.0.1:8080:8080 -v phrasemate-data:/data phrasemate:latest
```

容器内可使用主页面收录生词，或单独打开 `/float.html` 作为简易速记页。

## 发布与分发

本 fork 的发布入口：[Releases](https://github.com/Ryvith/PhraseMate/releases)。若尚无可用的 macOS 安装包，请按本文从源码构建。

现有 [GitHub Actions 构建配置](.github/workflows/build.yml) 包含 macOS 通用二进制流程：分别构建 `arm64` 和 `amd64`，使用 `lipo` 合并，再生成 `.app` 和 ZIP。macOS 产物目前仅做 ad hoc 签名，尚未公证，也不代表已在所有目标系统上验证可用。

可以在 **Actions → Build → Run workflow** 手动触发构建；推送 `v*` 标签会在相关构建任务全部成功后发布 Release。Docker 镜像会按当前仓库名称生成 GHCR 地址，首次对外分享前需确认包的可见性。

分发前应确认芯片架构、最低系统要求和关键功能；安装包不要包含个人数据库、`.env` 或 API Key。用户拿到 `.app` 后无需安装开发工具，在应用内填写自己的 AI 配置即可。

## 桌面别名

构建后的程序首次启动会尝试创建桌面别名，也可以通过菜单栏「创建桌面快捷方式」手动创建，或执行：

```bash
./dist/PhraseMate.app/Contents/MacOS/PhraseMate --install-shortcut
```

`go run` 的临时程序不会创建桌面别名。直接运行未打包的可执行程序时，创建别名的逻辑还可能在 `~/Applications` 中生成应用包。
