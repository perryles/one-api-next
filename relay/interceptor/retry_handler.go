package interceptor

import (
	"github.com/perryles/one-api-next/common/ctxkey"
	"github.com/perryles/one-api-next/relay/errcode"
)

type RetryHandler struct{}

func (h *RetryHandler) Name() string {
	return "retry"
}

func (h *RetryHandler) Handle(ctx *ErrorContext) bool {
	if _, ok := ctx.GinContext.Get(ctxkey.SpecificChannelId); ok {
		ctx.ShouldRetry = false
		return true
	}

	effectiveCfg := getEffectiveConfig(ctx.StatusCode)
	ctx.ChannelAction = effectiveCfg.Action

	if effectiveCfg.CooldownSeconds > 0 {
		ctx.CooldownSeconds = effectiveCfg.CooldownSeconds
	}

	ctx.ShouldRetry = errcode.ShouldRetry(effectiveCfg.Action)
	return true
}