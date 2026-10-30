# 上游对账与血缘记录（Upstream Lineage Ledger）

**项目定位：基于 songquanpeng/one-api 向后迭代；当前技术底座采用 modelbus/one-api-pro。**

以下第 1—5 节保留切换前历史记录，描述历史基线与功能状态，不代表当前运行代码；当前状态见第 6 节和 [BASELINE.md](BASELINE.md)。

本文件是 one-api-next 与上游 songquanpeng/one-api 之间的精确血缘记录。
任何人拿到本项目代码，都应能凭本文件追溯每一行修改的出处。

## 1. 上游基线（Baseline）

| 项 | 值 |
|---|---|
| 上游仓库 | https://github.com/songquanpeng/one-api |
| 基线分支 | `main` |
| 基线 commit | `8df4a26`（`docs: update ByteDance Doubao model link in README`） |
| 基线时间 | 2025-02-21 19:30:16 +0800 |
| 基线版本 | 上游未写入 VERSION（空文件；上游 release 最新为 v0.6.x 系列） |
| 许可 | MIT，Copyright (c) 2023 JustSong —— **保持不变** |

> 基线快照与本项目工作区起点逐字节一致（`diff -rq` 校验，仅排除 `.DS_Store` / `.idea` 等本机文件）。

## 2. 同步策略（Sync Policy）

- 本项目为主干继续开发（forward iteration），不定期回看上游；若上游恢复更新，择机以 `merge upstream/main` 方式合入，合入记录登记在本文件第 4 节。
- 上游原始文档保留于：
  - `README.upstream.md`（原 `README.md`，中文）
  - `README.en.md`、`README.ja.md`（上游英/日文翻译，未改动）

## 3. 变更清单（相对基线）

| 日期 | 类型 | 范围 | 说明 |
|---|---|---|---|
| 2026-10-31 | docs | `README.md`、`README.upstream.md`、`VERSION`、`docs/UPSTREAM.md`、`CHANGELOG.md` | 建立项目定位：声明基于 songquanpeng/one-api 向后迭代；原 README 归档为 README.upstream.md；引入版本号 v0.6.11-next.0 |
| 2026-10-31 | feat/billing | `relay/billing/ratio/model.go` | 模型价格表（ModelRatio / CompletionRatio）全量更新至 2026-10 官方牌价，覆盖 2025-2026 新模型（详见 CHANGELOG）；openrouter 旧快照整体替换 |
| 2026-10-31 | fix | `model/option.go`、`relay/billing/ratio/model.go` | CompletionRatio 支持存量库增量合并（对齐上游仅有的 ModelRatio 合并逻辑） |
| 2026-10-31 | feat/billing | `relay/billing/ratio/context_tiers.go`、`relay/controller/{text,helper}.go`、`model/option.go` | 上下文分档计价（机制移植自 omnirelays，默认关闭，运营设置 `ContextPriceTiers` JSON） |
| 2026-10-31 | feat/billing | `relay/billing/ratio/model.go`、`relay/controller/{text,audio,image,helper}.go`、`common/config/config.go` | 未知倍率模型拒绝计费（`GetModelRatioWithKnown`，`REJECT_UNKNOWN_MODEL_RATIO` 默认 true，机制移植自 omnirelays） |
| 2026-10-31 | chore | `go.mod` / `go.sum`、`.github/workflows/*` | Go 依赖全量升级（gorm 1.31、gin 1.10.x、x/crypto 等），go 指令升至 1.26，CI 同步；修复 go1.24+ vet 的非常量格式串问题 |
| 2026-10-31 | feat/router | `channelrouter/`、`relay/interceptor/`、`relay/errcode/`、`controller/relay.go`、`middleware/distributor.go`、`model/{channel,cache}.go`、`common/{config,ctxkey}` | 整包移植 one-api-pro 路由与韧性子系统：粘性会话/失败冷却/并发上限/RPM/错误拦截链/兜底渠道/指定渠道权限校验（剔除 cluster 多节点，见 THIRD_PARTY_NOTICES.md） |
| 2026-10-31 | docs | `CONTRIBUTING.md`、`SECURITY.md`、`.github/ISSUE_TEMPLATE/config.yml`、`README.md`、`THIRD_PARTY_NOTICES.md` | 贡献指南（含价格表维护 SOP）、安全报告政策、第三方来源声明（含 one-api-pro 无 LICENSE 文件的许可留痕） |

（后续每批修改持续追加于上表与 CHANGELOG.md）

## 4. 上游回灌记录（Upstream Merges）

暂无。上游自 2025-02-21 后无新 commit。

## 5. 归属与引用

- 核心架构、渠道适配器、前端主题、计费模型：**songquanpeng**（MIT）。
- 本项目增量修改：one-api-next 贡献者，同样以 MIT 发布。
- 价格数据来源：各厂商官方定价页 + [models.dev](https://models.dev) 开放数据集（抓取时间 2026-10-31）。

## 6. 技术底座切换（2026-10-31）

- 原始来源：songquanpeng/one-api，MIT，Copyright (c) 2023 JustSong；原始 README 及许可继续保留。
- 当前直接技术底座：modelbus/one-api-pro，精确提交 `b4369989ae59c2c2f901df3d6aed212f7415d821`，版本 0.0.25。
- 本仓库切换提交：`a0f8603`，保留原 Git 历史；Go module 改为 `github.com/perryles/one-api-next`。
- 原底座的贡献、安全、血缘与来源文档恢复保留，开发步骤按 Vue 和当前计费体系更新。
- 历史静态价格更新、上下文分档等实现没有直接覆盖新的数据库计费体系；其旧记录仅用于追溯。
