# one-api-next 开发基线

本项目从 2026-10-31 起以 https://github.com/modelbus/one-api-pro 为完整代码底座。

- 上游分支：master
- 精确提交：b4369989ae59c2c2f901df3d6aed212f7415d821
- 上游版本：0.0.25
- 保留当前仓库 Git 历史及 origin，不重写提交历史。
- 当前维护前端：web/default-pro（Vue 3）。上游原生套餐、订单、支付、集群、渠道和国际化模块一并保留。
- 原 songquanpeng/one-api 底座及全部未提交改动已在项目外完整备份；旧 React 界面适配由上游 Vue 实现替代。
- 旧底座新增的静态价格表、上下文分档机制没有直接覆盖上游数据库定价体系；原实现保存在迁移前快照和旧提交中，后续须按新计费架构适配。
- 此次为源代码基线切换，没有连接或迁移生产数据库。已有数据库部署应先备份并验证升级路径。

上游 README 声明 MIT，但快照不含 LICENSE；保留原 one-api MIT 许可文本和上游版权说明，不将上游新增代码的权利归属于本项目。

## 验证结果

- `npm run build`（web/default-pro）：通过。
- `node --test web/default-pro/src/utils/*.test.mjs`：62 项通过。
- `go build ./...`：通过。
- `go test ./model/ ./controller/ ./middleware/`：通过。
- `go test ./...`：未全通过。上游 common/image 的测试依赖远程 Wikimedia 图片，收到非图片响应后解码失败；relay 的 TestGetAdaptor 对已废弃/未注册编号仍断言非空。此次保留上游测试，没有跳过或弱化断言。
- 上游源码完整导入；仅清理 docs/API.md 三处尾随空格，并新增本文件说明。

## 下游模块标识

Go module 与内部引用统一为 `github.com/perryles/one-api-next`，同步更新发布脚本、Docker 和 CI 的版本注入路径；上游来源与历史记录继续保留。
