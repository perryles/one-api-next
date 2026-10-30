# 协议入口

以下入口共用当前项目的用户令牌、模型权限、路由和计费。需要先在渠道中配置可用模型，并在模型价格表中配置相应价格。

## OpenAI Responses

`POST /v1/responses`，使用 `Authorization: Bearer <token>`。

支持 OpenAI、自定义和 OpenAI Compatible 渠道；上游必须原生提供 Responses API。模型字段按渠道映射重写，其余请求字段保留。普通 JSON 与 SSE 响应均透传，按 input_tokens、output_tokens、input_tokens_details.cached_tokens 结算，并复用分组折扣和套餐逻辑。

当前仅提供同步创建入口，不提供 GET/DELETE、后台任务、取消任务或 `previous_response_id` 会话续接。后台任务和会话续接参数会被明确拒绝；请在 input 中传入完整会话历史。没有把 Responses 请求降级转成 Chat Completions，因此仅有 Chat 接口的上游不可使用本入口。

```bash
curl http://localhost:3002/v1/responses \
  -H "Authorization: Bearer $ONE_API_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"model":"gpt-4o","input":"Hello","stream":false}'
```

## Anthropic Messages

`POST /v1/messages`，支持 `x-api-key: <token>` 或 `Authorization: Bearer <token>`。沿用项目令牌，不是上游供应商的 API Key。请求转换到现有 Chat 链路，响应转换回 Messages，包括工具调用和流式事件。

支持 system、文本、图片、tool_use、tool_result、tools、tool_choice、stop_sequences、max_tokens 和流式返回。max_tokens 必须为正数。支持范围以现有渠道 Chat 适配器能力为准；不支持的内容块返回 400，不会静默丢弃。扩展思考、服务端工具等高级协议特性不属于本批兼容范围。

```bash
curl http://localhost:3002/v1/messages \
  -H "x-api-key: $ONE_API_TOKEN" \
  -H 'anthropic-version: 2023-06-01' \
  -H 'Content-Type: application/json' \
  -d '{"model":"claude-sonnet-4-5","max_tokens":256,"messages":[{"role":"user","content":"Hello"}]}'
```

## 本地 token 估算

`POST /v1/messages/count_tokens`：同 Messages 鉴权和模型权限检查，要求 model 和 messages，不选渠道、不发起上游调用、不扣调用费用。返回 `{ "input_tokens": N }`。

这是基于请求大小的本地近似值，不是 Anthropic 官方 tokenizer 或官方计数接口结果；不能用于精确账单或保证不会超出模型上下文。真实费用仍按转发响应的 usage 结算。

## 验证范围

使用本地模拟上游验证模型映射、请求字段、响应透传和缓存 token 扣费；转换测试覆盖 Messages 文本、图片、工具往返、SSE 和错误格式。没有使用真实供应商密钥做在线兼容认证。
