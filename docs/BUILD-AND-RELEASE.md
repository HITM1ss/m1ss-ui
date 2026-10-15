# M1ss-UI 自动构建与部署

本仓库在 S-UI 的构建链上发布定制前端，不修改协议构建标签。`frontend/` 已作为普通源码目录纳入主仓库，因此 Actions 不再拉取上游前端子模块。前端的 GPL 许可证和原有源码说明仍保留。

## 自动化入口

| 操作 | 工作流 | 结果 |
| --- | --- | --- |
| 推送到 `main`、提交 Pull Request，或手动运行 CI | `ci.yml` | 前端测试、类型检查和生产构建；Go 格式、vet、单元测试和指定包的 race 检查。前端产物保留 7 天，不发布版本。 |
| 推送 `vX.Y.Z` 或 `vX.Y.Z-rc.N` 标签 | `release.yml` | 测试通过后，构建全部 Linux/Windows 安装包，生成校验文件并发布到当前仓库 Releases。带 `-` 的标签是预发布版。 |
| 手动运行 Build and release M1ss-UI | `release.yml` | 生成 Actions Artifacts，不创建 Release。 |
| 推送上述版本标签 | `docker.yml` | 测试并构建多架构镜像，发布到当前仓库的 GHCR。稳定版本更新 `latest`，预发布版本不更新。 |
| 手动运行 Build and publish Docker image | `docker.yml` | 发布 `manual-<分支或标签>` 和 `sha-<提交>` 镜像标签，不更新版本标签或 `latest`。 |

GitHub 上需启用 Actions。发布步骤使用该仓库自动提供的 `GITHUB_TOKEN`；无需配置 Docker Hub 账户或服务器 SSH 密钥。只有 Release 发布任务授予 `contents: write`，镜像发布任务授予 `packages: write`。首次使用 GHCR 后，可在 Package settings 将镜像设为 Public；若保留 Private，服务器拉取前需使用具有 `read:packages` 权限的账号登录 GHCR。

## 产物对照

官方核对来源（2026-10-15）：[S-UI v1.6.5](https://github.com/alireza0/s-ui/releases/tag/v1.6.5)、[官方 Release 工作流](https://github.com/alireza0/s-ui/blob/main/.github/workflows/release.yml)、[官方 Docker 工作流](https://github.com/alireza0/s-ui/blob/main/.github/workflows/docker.yml)。本地后端版本仍以 `config/version` 为准，配置构建流不会自动升级后端代码。

| 类别 | 本项目保留的构建目标 |
| --- | --- |
| Linux `.tar.gz` | `amd64`、`arm64`、`armv7`、`armv6`、`armv5`、`386`、`s390x` |
| Windows `.zip` | `amd64`、`arm64` |
| Docker | `linux/amd64`、`linux/386`、`linux/arm64/v8`、`linux/arm/v7`、`linux/arm/v6` |
| 发布附加文件 | `SHA256SUMS`、`install.sh`、`s-ui.sh` |

安装包名称保留 `s-ui-linux-<架构>.tar.gz` / `s-ui-windows-<架构>.zip`，Linux 包内目录保留 `s-ui/`，以兼容既有安装路径、服务名称与数据目录。页面名称为 M1ss-UI。CI 从同一提交构建前端，将 `dist/` 放入 `web/html/`，再由 Go 嵌入程序；服务器运行时不需要 Node.js 或 Go。

Linux 构建保留官方 CGO、musl/Cronet 工具链及 `build-tags.sh` 配置。官方不启用 Naive 的目标仍使用原有非 Naive 配置；不能用随意省略标签的 `go build` 代替。Windows ARM64 启用 CGO 并使用固定摘要校验的 LLVM-MinGW，避免关闭 CGO 后 SQLite 无法打开数据库。Docker ARMv6 的运行兼容性仍需目标服务器验证。

## 发布一个版本

先提交并推送当前代码（包括 `frontend/` 和 `.github/`），然后推送版本标签。例如当前代码基于 1.6.4，可选择尚未使用的 `v1.6.4`；如果已存在则使用新的版本号，不覆盖旧标签：

```sh
git push origin main
git tag v1.6.4
git push origin v1.6.4
```

在 Actions 页面等待 Release 与 Docker 两个工作流结束。全部安装包成功后才发布 Release；镜像的多架构清单也只在所有镜像架构成功后更新。两个工作流独立运行，失败后在对应工作流查看日志并重新运行。标签构建会把去掉 `v` 的版本写入构建工作区的 `config/version`，安装包与镜像因此显示相同版本；手动构建保留源码版本。这不会替你合并上游更新，发布标签仍应对应实际代码。

如果只想试编译，打开 Actions → Build and release M1ss-UI → Run workflow。成功后从该次运行的 Artifacts 下载对应安装包；不必先创建 Release。

## Linux 安装或更新定制版

以下命令应在自己的 Linux 服务器上执行，首次稳定 Release 成功后才可使用。已有服务器先导出面板备份并保存证书。安装器与菜单脚本的下载地址已改为 `HITM1ss/m1ss-ui`，后续升级不会自动切回上游界面。

```sh
curl -fL https://raw.githubusercontent.com/HITM1ss/m1ss-ui/main/install.sh -o install-m1ss-ui.sh
sudo bash install-m1ss-ui.sh
```

预发布版不会出现在 `releases/latest`，安装时需显式指定版本，例如 `sudo bash install-m1ss-ui.sh v1.6.4-rc.1`。安装包内仍使用 `sui` 可执行文件和 `s-ui` 服务名称，数据库与证书目录保持兼容。

## Docker 部署

首次稳定镜像发布成功后，使用仓库中的 Compose 文件：

```sh
docker compose pull
docker compose up -d
```

默认镜像为 `ghcr.io/hitm1ss/m1ss-ui:latest`。也可指定固定版本，例如 `ghcr.io/hitm1ss/m1ss-ui:1.6.4`。`db` 与 `cert` 挂载保持不变。原来使用上游 Compose 的服务器，应先保留原数据目录再更换镜像；端口、证书路径和实际使用的协议端口沿用原服务器配置。

## 本地校验与验证范围

可运行 `npm ci && npm test && npm run build`（在 `frontend/` 中）以及 `actionlint .github/workflows/*.yml` 检查工作流。`.github/actionlint.yaml` 只为校验器补充 GitHub 已提供的 `ubuntu-26.04` 标签，不改变实际 runner 选择。

完整跨架构编译需要 GitHub Runner 与对应工具链。静态校验和前端测试通过，并不代表 Linux/Windows 安装包或 Docker 镜像已经构建成功；以首次 Actions 实际运行结果为准。本配置不会自动登录或部署到服务器。
