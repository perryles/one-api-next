# 安全政策（Security Policy / 安全问题报告）

本项目基于 [songquanpeng/one-api](https://github.com/songquanpeng/one-api) 向后迭代，当前采用 one-api-pro 技术底座；来源见 [血缘记录](docs/UPSTREAM.md)。

## 支持范围

优先处理 one-api-next 当前主干可复现的安全问题。版本以 `VERSION` 和提交 SHA 为准；不承诺为所有历史版本提供回补。若问题也影响原始 one-api 或 one-api-pro，鼓励通过对应上游的私密渠道协调报告。

## 如何报告安全漏洞

**请勿通过公开 Issue、Discussion、群聊或 PR 披露漏洞细节、PoC、密钥或用户数据。**

首选 [GitHub 私密漏洞报告](https://github.com/perryles/one-api-next/security/advisories/new)。该入口只有仓库启用私密报告后才可用；若入口不可见，请通过普通 Issue 仅请求维护者提供私密联系方式，不包含复现或漏洞细节。目前没有公布可用的安全邮箱，不应使用旧文档中的 `.example` 占位邮箱。

报告时请尽量包含：

1. 影响版本号（`VERSION` 文件或 `/api/status` 返回）与部署形态（Docker / 二进制 / 反代）；
2. 复现步骤（请求报文、必要 payload、截图）；涉及鉴权的请用**你自己搭建的实例**演示；
3. 影响评估（越权范围、数据泄露面、是否需要登录态）；
4. 如有，附修复建议或 PoC。

## 处理流程与时间线

以下为维护目标，不构成响应或修复时限保证。

- **确认（目标 72 小时内）**：维护者确认收悉并给出初步判断；
- **评估（目标 7 天内）**：定级（Critical / High / Medium / Low）与修复方案；
- **修复发布**：Critical/High 原则上 14 天内发布修复版本；中低危随常规版本发布；
- **披露**：修复发布后统一在 [CHANGELOG](./CHANGELOG/README.md) 与 GitHub Security Advisories 公告；报告者可获得致谢（可匿名）。在修复发布前，请勿公开细节（协调披露）。

## 漏洞范围（In Scope）

本项目是 LLM API 网关，重点接收以下类别：

- **鉴权与越权**：user token / access token / session 伪造、绕过；普通用户提权到管理员；
- **计费与配额绕过**：quota 扣减绕过、负数/整数溢出、令牌额度耗尽后仍可用；
- **SQL 注入 / 命令注入**：任何进入 `model/` 层查询或 shell 的注入；
- **SSRF**：渠道 `base_url`、代理设置、出站回调等触发内网探测（本网关管理大量上游地址，SSRF 影响放大）；
- **前端安全**：控制台 XSS（渠道名、模型名、日志渲染等注入点）、CSRF；
- **敏感信息泄露**：日志/报错接口回显上游 API key、管理员数据；
- **供应链**：依赖中的已知 CVE（请附受影响版本与利用条件）。

## 不在范围（Out of Scope）

- 需要管理员权限或服务器 shell 才能触发的问题（已有最高权限）；
- 上游 API key 因管理员自己泄露/配置错误导致的滥用；
- 纯粹的速率限制缺失、暴力破解弱口令（属加固建议，走普通 Issue）；
- 社工、钓鱼、物理访问，以及针对第三方上游客网的攻击；
- 已经公开且上游基线同样存在的历史问题（仍欢迎报告，按低优先级处理）。

## 加固建议（Hardening，非漏洞）

生产部署请至少：容器只读 + 非 root 运行、`SESSION_SECRET` 显式设置、数据库与 Redis 不暴露公网、管理后台加 IP 白名单/反代认证、定期轮换全部上游渠道 key。当前部署说明见 [README.md](./README.md) 与 [docs](./docs/)；[README.upstream.md](./README.upstream.md) 仅作原始上游历史参考。

## 致谢

感谢每一位负责任披露的安全研究者；名单随修复公告滚动更新。
