# Changelog（one-api-next）

本项目基于 [songquanpeng/one-api](https://github.com/songquanpeng/one-api)（基线 commit `8df4a26`，2025-02-21）向后迭代。血缘与对账见 [docs/UPSTREAM.md](./docs/UPSTREAM.md)。

## v0.6.11-next.0（2026-10-31 起，进行中）

### 定位与工程
- 明确项目定位：基于 songquanpeng/one-api 的下游延续项目；原 README 归档为 `README.upstream.md`，新增 `docs/UPSTREAM.md` 血缘对账。
- `VERSION` 由空文件改为 `v0.6.11-next.0`（延续上游 v0.6 系列）。
- 新增 [CONTRIBUTING.md](./CONTRIBUTING.md)：贡献流程、本地开发环境、**价格表维护 SOP**（ratio 换算公式、数据源要求、只增不删原则）、新渠道适配器五步清单、提交规范。
- 新增 [SECURITY.md](./SECURITY.md)：私密漏洞报告通道（GitHub Security Advisories 优先）、72h/7d/14d 响应时间线、网关类项目专属漏洞范围（鉴权/计费绕过/SSRF/SQLi/XSS/密钥泄露）、负责任披露与致谢。
- Issue 模板 `config.yml` 增加「安全漏洞报告」引导链接，防止漏洞经公开 Issue 泄露。

### 计费 / 价格表（核心更新）
- `ModelRatio` / `CompletionRatio` 全量刷新至 2026-10 官方牌价（数据源：厂商定价页 + models.dev），新增覆盖（摘选）：
  - **OpenAI**：gpt-4.1 / 4.1-mini / 4.1-nano、o3 / o3-pro / o4-mini、gpt-5 / 5-mini / 5-nano / 5-pro / 5.1 / 5.2 / 5.2-pro / 5.3-codex / 5.4 系 / 5.5 系 / 5.6 系、gpt-6 系（astra / luna / sol / terra）
  - **Anthropic**：claude-3.7-sonnet、claude-3-5-haiku 新价、claude-4 / 4.5（sonnet / opus / haiku）、claude-5 / 5.5 系
  - **Google Gemini**：2.0-flash / flash-lite、2.5 pro / flash / flash-lite、3.x（pro / flash / flash-lite）、gemini-4-argon
  - **DeepSeek**：V3.1 / V3.2（chat / reasoner 新价）、V4 系 deepseek-flash / deepseek-v4-pro（含峰谷双价说明）
  - **xAI**：grok-3 / grok-4、grok-4.3 – 4.7、grok-4.20 系
  - **阿里通义**：qwen3 全系、qwen3-coder、qwen3-vl、qwen3.5 – 3.8 系、qwen-max/turbo/flash 新价、qwq-plus
  - **智谱**：glm-4.5 / 4.5-air / 4.5v、glm-4.6 / 4.6v、glm-4.7 系、glm-5 – 5.3 系（含 flash / flashx）
  - **月之暗面**：kimi-k2 系（0905 / thinking / instruct / K2.5 / K2.6 / K2.7-code）、kimi-k3 / k3.1-pro
  - **MiniMax**：M2 / M2.1 / M2.5 / M2.7、M3、M3.1-Flash
  - **字节豆包**：doubao-seed-1.6 / 1.8 / 2.0 / 2.1 系
  - **阶跃**：step-3.5-flash / step-3.7-flash / step-5-preview
  - **SiliconFlow / Mistral / Cohere / Perplexity / Groq / Together / Novita / Cloudflare**：主力模型价格刷新
- `GetCompletionRatio` 前缀默认规则同步更新（gpt-4.1 / gpt-5.x / gpt-6 / o3 / o4-mini / claude-4+ / gemini-2.5+ / deepseek-v4 / qwen3+ / glm-4.5+ / kimi-k2+ / doubao-seed 等）。

### 缺陷修复
- 修复：`CompletionRatio` 默认表对存量数据库不做增量合并的问题（上游仅合并 `ModelRatio`），存量实例升级后新模型的输出倍率可正常生效。

### 依赖
- Go / gin / gorm / x-crypto 等依赖升级（详见 go.mod 变更与 docs/UPSTREAM.md 对账表）。

### 计划中（本版本目标）
- 前端三主题（default / berry / air）依赖刷新与构建验证。
