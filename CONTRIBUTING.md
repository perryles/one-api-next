# 参与贡献（Contributing to one-api-next）

one-api-next 是基于 [songquanpeng/one-api](https://github.com/songquanpeng/one-api) 向后迭代的项目，保留原作者的版权与贡献记录。当前技术底座采用 [modelbus/one-api-pro](https://github.com/modelbus/one-api-pro)，后续在此基础上持续开发。完整血缘见 [docs/UPSTREAM.md](docs/UPSTREAM.md)。欢迎将兼容的通用修复同时回馈相关上游。

## 报告问题与贡献代码

- Bug 和功能建议请使用 [Issue 模板](.github/ISSUE_TEMPLATE/)。
- 安全问题请先阅读 [SECURITY.md](SECURITY.md)，不要公开漏洞细节。
- Fork 后创建分支并提交 PR，使用 [PR 模板](pull_request_template.md)；一个 PR 聚焦一个功能或修复。
- 分支可使用 `feat/`、`fix/`、`docs/`；提交信息使用 Conventional Commits，清楚说明行为和验证结果。
- 修改第三方代码时保留版权和来源，在 [THIRD_PARTY_NOTICES.md](THIRD_PARTY_NOTICES.md) 与血缘记录中注明来源。

## 本地开发与验证

Go 版本以 `go.mod` 为准（当前至少 1.25）；前端使用 Node.js 22.12+ 或满足当前 Vite 要求的更新版本。

```bash
# Vue 控制台：先构建供 Go 嵌入的静态资源
cd web/default-pro
npm install
npm run build
cd ../..

# 后端与测试
go run .
go build ./...
go test ./model/ ./controller/ ./middleware/
go test ./...
node --test web/default-pro/src/utils/*.test.mjs

# 前端开发（另一个终端）
cd web/default-pro
npm run dev
```

前端开发服务默认在 :3001，代理后端 :3000。仅维护 `web/default-pro/`，不直接编辑 `web/build/` 产物。后端入口为 `main.go`；适配器位于 `relay/adaptor/`，通过 `relay/registry/` 自注册。

全量测试的已知上游问题见 [docs/BASELINE.md](docs/BASELINE.md)，请在 PR 中区分已有失败与本次引入的问题。前端改动提供截图和构建结果。

## 模型价格维护

当前使用数据库模型价格体系，定义见 `model/model_price.go`，计费实现见 `relay/billing/`。更新价格以厂商官方资料为准，可用 models.dev 交叉核对，记录币种、计价单位、数据来源和日期。

- 分别核对输入、输出、缓存 token 和按请求计费，提供额度计算测试。
- 不应覆盖管理员自定义价格，或随意删除存量模型；数据库变更使用兼容的增量迁移。
- 旧底座的静态倍率维护 SOP 保留在 [历史贡献指南](docs/history/CONTRIBUTING.md)，不能直接用于当前数据库定价体系。

## 新渠道与代码要求

适配器实现及注册流程见 [新增供应商指南](docs/contribute/add-provider-zh.md)。已有渠道编号不可复用或重排，须验证存量配置兼容性。Go 代码使用 gofmt；鉴权、额度扣减、支付回调验签及幂等性等关键路径必须覆盖测试。

Issue 和 PR 可使用中文或英文；尊重原作者及其他贡献者。不得提交真实密钥、账号、支付证书、内部数据或可识别用户信息。
