# 第三方来源声明（Third-Party Notices）

本文件按 [docs/UPSTREAM.md](./docs/UPSTREAM.md) 的血缘体系，逐一声明并入本仓库的第三方来源与其许可状态。

## 1. songquanpeng/one-api（直接上游基线）

- 来源：<https://github.com/songquanpeng/one-api>，commit `8df4a26`（2025-02-21）
- 许可：MIT，Copyright (c) 2023 JustSong —— 本仓库根目录 [LICENSE](./LICENSE) 即其原文，保持不变
- 状态：本项目代码底座；全部修改以同一 MIT 协议开放

## 2. omnirelays（本地项目，机制移植）

- 来源：`omnirelays/omnirelays-gateway`（维护者本人私有仓库）
- 移植内容：
  - `relay/billing/ratio/model.go`：`GetModelRatioWithKnown`（未知倍率拒绝计费）与 `omni-moderation-*` 占位倍率
  - `relay/billing/ratio/context_tiers.go`（含测试）：上下文分档计价机制
- 许可：维护者本人作品，本人书面确认（本仓库存在即为准据）可按 MIT 并入
- 备注：omnirelays 的 Reservation（预扣额度事务）、渠道级 model_prices 落表价格、按 token 快照计费等列为后续批次，尚未并入

## 3. modelbus/one-api-pro（路由与韧性子系统，整包移植）

- 来源：<https://github.com/modelbus/one-api-pro>（clone 快照，VERSION 0.0.25，commit `b436998` 附近，2026-10-31 抓取）
- 移植内容：
  - `channelrouter/`：渠道路由引擎（过滤/冷却/并发/RPM/粘性会话/选择器，含测试）
  - `relay/interceptor/`：错误拦截链（响应映射 / 重试决策 / 渠道动作）
  - `relay/errcode/`：按状态码的错误处置策略（含测试）
  - `controller/relay.go`：重试主流程重写（含兜底渠道 fallback 与请求模型改写）
  - `middleware/distributor.go`：指定渠道的分组/模型权限校验、经 router 选路
  - `model/channel.go`、`model/cache.go`：渠道新列（max_concurrency / rpm / cooldown_seconds / is_fallback / fallback_priority / last_error*）、候选集与兜底渠道查询
  - `common/config/config.go`：`CHANNEL_*` 开关族；`common/ctxkey/key.go`：`SessionKey`
- **适配性修改**（与上游 one-api-pro 的差异）：
  - 剔除 `cluster`（多节点）依赖：并发/RPM 仅保留单机内存态实现；多节点列为后续批次
  - `relay/schema` → 本仓库 `relay/model` 路径映射
  - 兜底渠道查询的随机挑选逻辑按本仓库风格重写
- ⚠️ **许可注意**：one-api-pro 的 README 展示 MIT 徽章并在文首声明"Deep refactor and extension of one-api"，
  但其仓库内**未包含 LICENSE 文件**。本项目依据其徽章声明与 one-api（MIT）的衍生关系按 MIT 对待并入，
  并在此留痕；建议上游尽快补交 LICENSE 文件，本声明将随其更新而修订。
- 未并入（列为后续批次）：DB 级模型价格表（model_prices）、套餐/订阅/订单/支付、多节点集群、Vue3 控制台

## 4. 价格数据来源

- 各厂商官方定价页（OpenAI / Anthropic / Google / DeepSeek 等，抓取于 2026-10-31）
- [models.dev](https://models.dev) 开放数据集（API JSON，抓取于 2026-10-31）
- 用途：`relay/billing/ratio/model.go` 中 2025-2026 增补条目；数据本身为事实型价格信息，不构成代码版权
