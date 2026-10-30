# 参与贡献（Contributing to One API Next）

感谢你愿意为 One API Next 出力！本项目基于 [songquanpeng/one-api](https://github.com/songquanpeng/one-api) 向后迭代（血缘见 [docs/UPSTREAM.md](./docs/UPSTREAM.md)），**所有与上游兼容的通用改进，我们鼓励同时回赠上游**；本项目优先接纳上游已无力跟进的部分（价格表、新模型、依赖与安全）。

## 0. 快速导航

- 报 Bug / 提需求：走 [Issue 模板](./.github/ISSUE_TEMPLATE/)（中文）
- 安全漏洞：**不要开公开 Issue**，走 [SECURITY.md](./SECURITY.md) 的私密通道
- 提交代码：Fork → 分支 → PR，模板见 [pull_request_template.md](./pull_request_template.md)
- 价格表更新：见下文「A. 价格表维护 SOP」（本项目最高频贡献类型）

## 1. 本地开发环境

```bash
# 后端（Go >= 1.22，与 CI 一致）
go run main.go            # 默认 SQLite，监听 :3000
go test ./...             # 全量测试

# 前端（三主题：default / berry / air，均为 CRA 工程）
cd web/default && npm install && npm run start   # 开发
npm run build                                    # 产物输出到 ../build/default
```

- 后端入口 `main.go`；渠道适配器在 `relay/adaptor/`；计费在 `relay/billing/`。
- 前端与后端联调：`npm run start` 默认代理到 :3000。
- **不要**直接改动 `web/build/` 下的构建产物，它们由 `web/build.sh` 生成。

## 2. 分支与提交规范

- 分支命名：`feat/<topic>`、`fix/<topic>`、`chore/<topic>`、`docs/<topic>`。
- 提交信息沿用上游风格（conventional commits）：
  `feat: support xxx` / `fix: correct xxx ratio` / `docs: ...` / `chore: ...`
- 一个 PR 只做一件事；价格表批量更新单独成 PR，便于对账。

## 3. A. 价格表维护 SOP（高频）

价格定义在 `relay/billing/ratio/model.go`：

```go
// ModelRatio: 1 === $0.002 / 1K tokens，即 ratio = (每 1M tokens 美元价) / 2
// 例：gpt-4.1 输入 $2 / 1M  -> ratio 1.0；输出 $8 / 1M
// CompletionRatio: 输出价 / 输入价 的倍率，例：8 / 2 = 4
```

1. **数据源**：以厂商官方定价页为准；聚合数据可用 [models.dev](https://models.dev) 交叉核对。人民币牌价按 `RMB` 常量（默认 1 USD = 7 CNY）换算。
2. **峰谷价**（如 DeepSeek）：默认取**非高峰（off-peak）牌价**计入 `ModelRatio`，并在行尾注释标注 `peak` 价；管理员可在运营设置里自行覆盖。
3. **新增模型**时同步做三件事：
   - `ModelRatio` 加输入价；
   - `CompletionRatio` 加倍率（或确认 `GetCompletionRatio` 的前缀默认分支已覆盖该系列）；
   - 行尾注释标明数据来源与抓取日期，例：`// $2 / 1M in, 2026-10-31 openai.com/pricing`。
4. **兼容性**：只增不删；厂商停产的旧模型条目保留（存量渠道可能仍在路由）。
5. **存量库生效**：`ModelRatio` / `CompletionRatio` 均会在启动时对老数据库做增量合并（后者为本项目修复的上游缺陷），新默认值无需手工刷库；但管理员**手动改过**的键不会被覆盖。

## 4. B. 代码与测试要求

- Go：`gofmt` 格式化；新适配器需带 `*_test.go`（参照 `relay/adaptor_test.go`）。
- 计费相关改动必须附测试：给定 input/output tokens 断言 quota 计算。
- 前端改动需在对应主题内完成并通过 `npm run build`。
- PR 描述中给出**自测步骤**与截图（UI 改动）。

## 5. C. 新渠道适配器（中频）

1. 在 `relay/adaptor/<provider>/` 新建 adaptor，实现 `relay/adaptor/interface.go`。
2. `relay/channeltype/define.go` 追加常量（**只允许追加在末尾**，已有编号不可复用/重排，存量数据库按编号存储渠道类型）。
3. `relay/adaptor/`（`GetAdaptor`）、`relay/channeltype/helper.go`、`controller/channel-test.go`、前端 `web/*/src/constants/channel.constants.js` 同步登记。
4. 提供可公开验证的官方 API 文档链接；密钥用占位符，**不要提交真实 key**。

## 6. 行为准则

- Issue / PR 使用简体中文或英文；保持对上游原作者与其他贡献者的尊重与致谢。
- 涉及上游基线代码的修复，注明 `upstream-relevant`，维护者会协助回赠上游。
- 禁止在仓库内提交任何真实账号、密钥、内部数据。
