<p align="right">
   <strong>中文</strong> · 上游原文请见 <a href="./README.upstream.md">README.upstream.md</a>
</p>

<div align="center">

# One API Next

_✨ 基于 [songquanpeng/one-api](https://github.com/songquanpeng/one-api) 向后迭代的项目 ✨_

**通过标准的 OpenAI API 格式访问所有的大模型，开箱即用**

</div>

---

## 项目定位（必读）

**One API Next（one-api-next）是基于 [songquanpeng/one-api](https://github.com/songquanpeng/one-api) 的下游延续项目（downstream continuation），不是重写、不是替代，而是在原项目基础上的向后迭代。**

- **上游仓库**：<https://github.com/songquanpeng/one-api>
- **基线快照**：上游 `main` 分支 commit [`8df4a26`](https://github.com/songquanpeng/one-api/commit/8df4a26)（2025-02-21 19:30:16 +0800，`docs: update ByteDance Doubao model link in README`）
- **分叉原因**：上游自 **2025 年 2 月** 起基本停止功能性维护，模型价格表、新模型支持与依赖安全更新长期滞后；本项目在此基线上继续向前迭代。
- **许可继承**：本项目继承上游 [MIT License](./LICENSE)（Copyright © 2023 JustSong / songquanpeng），原版权声明保持不变，所有修改以同一 MIT 协议开放。
- **成果归属**：核心架构（Go 后端、渠道适配器体系、React 前端三主题、计费与令牌模型等）全部来自原作者 [songquanpeng](https://github.com/songquanpeng) 的工作，本项目在其之上做增量演进，详见 [docs/UPSTREAM.md](./docs/UPSTREAM.md)。

> 简单说：**one-api 的代码底座 = songquanpeng 的贡献；one-api-next = 这份底座的 2025-2026 持续维护与增强。**

## 与上游的差异（迭代方向）

| # | 方向 | 状态 | 说明 |
|---|------|------|------|
| 1 | **模型价格表大更新** | 进行中 | 上游价格停留在 2025-02（最高只到 `o3-mini` / `claude-3-5`）。补齐 2025–2026 全系：OpenAI GPT-4.1 / GPT-5.x / GPT-6 系列、o3 / o3-pro / o4-mini；Anthropic Claude 3.7 / 4 / 4.5 / 5.x；Google Gemini 2.5 / 3.x / 4；DeepSeek V3.x / V4（flash / pro，含峰谷价说明）；Qwen3–3.8、GLM-4.5–5.3、Kimi K2–K3、MiniMax M2–M3.1、豆包 Seed、StepFun、SiliconFlow、Mistral / Cohere / Perplexity 等 |
| 2 | **CompletionRatio 合并缺陷修复** | 进行中 | 上游仅对 `ModelRatio` 做存量库增量合并（`AddNewMissingRatio`），`CompletionRatio` 更新永远无法下发到已有数据库 —— 本项目补齐 |
| 3 | **依赖与安全升级** | 进行中 | Go 工具链、gin / gorm / `golang.org/x/crypto` 等依赖升级，消化上游滞后的 CVE |
| 4 | **前端构建链刷新** | 计划中 | 三主题（default / berry / air）依赖升级与可构建性验证 |
| 5 | **版本管理规范化** | 已完成 | 上游 `VERSION` 为空文件；本项目引入 `v0.6.x-next.N` 版本号（延续上游 v0.6 系列） |

完整变更记录见 [CHANGELOG.md](./CHANGELOG.md)；与上游的精确对账表见 [docs/UPSTREAM.md](./docs/UPSTREAM.md)。

## 快速开始（部署方式与上游一致）

```bash
# Docker
docker run -d --name one-api-next -p 3000:3000 \
  -v /path/to/data:/data \
  <本项目镜像>   # 与上游 justsong/one-api 镜像用法兼容

# 源码构建
go build -o one-api-next
# 前端（三主题任选其一）
cd web/default && npm install && npm run build
```

环境变量、数据库（SQLite/MySQL/PostgreSQL）、渠道配置、令牌管理等使用方法与上游完全兼容，参见 [README.upstream.md](./README.upstream.md) 的「部署」「使用方法」「常见问题」章节。

## 参与贡献与安全报告

- **参与贡献**：开发环境、价格表维护 SOP、新渠道适配器流程见 [CONTRIBUTING.md](./CONTRIBUTING.md)。
- **安全漏洞**：请走私密通道（GitHub Security Advisories / 邮箱），**不要开公开 Issue** —— 范围、时间线与加固建议见 [SECURITY.md](./SECURITY.md)。

## 版本规则

- 延续上游 `v0.6.x` 系列：本项目首个版本为 **`v0.6.11-next.0`**（基线 = 上游 `8df4a26` + 本项目增量）。
- `-next.N` 后缀表示基于同一上游基线的第 N 次迭代发布。

## 致谢

- **[songquanpeng / JustSong](https://github.com/songquanpeng)** —— one-api 原作者，本项目全部核心能力的奠基者。
- one-api 社区的所有贡献者（见 [上游贡献者列表](https://github.com/songquanpeng/one-api/graphs/contributors)）。
- 价格数据参考：各厂商官方定价页与 [models.dev](https://models.dev) 开放数据集。

## License

[MIT](./LICENSE) · Copyright © 2023 JustSong（上游），one-api-next 贡献者（增量修改）
