package middleware

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/songquanpeng/one-api/channelrouter"
	"github.com/songquanpeng/one-api/common/config"
	"github.com/songquanpeng/one-api/common/ctxkey"
	"github.com/songquanpeng/one-api/common/logger"
	"github.com/songquanpeng/one-api/model"
	"github.com/songquanpeng/one-api/relay/channeltype"
)

type ModelRequest struct {
	Model string `json:"model" form:"model"`
}

func Distribute() func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		userId := c.GetInt(ctxkey.Id)
		userGroup, _ := model.CacheGetUserGroup(userId)
		c.Set(ctxkey.Group, userGroup)
		var requestModel string
		var channel *model.Channel
		channelId, ok := c.Get(ctxkey.SpecificChannelId)
		if ok {
			id, err := strconv.Atoi(channelId.(string))
			if err != nil {
				abortWithMessage(c, http.StatusBadRequest, "无效的渠道 Id")
				return
			}
			channel, err = model.GetChannelById(id, true)
			if err != nil {
				abortWithMessage(c, http.StatusBadRequest, "无效的渠道 Id")
				return
			}
			if channel.Status != model.ChannelStatusEnabled {
				abortWithMessage(c, http.StatusForbidden, "该渠道已被禁用")
				return
			}
			// one-api-pro 移植：指定渠道也要校验分组与模型权限
			if !channel.ContainsGroup(userGroup) {
				abortWithMessage(c, http.StatusForbidden, "当前分组无权使用该渠道")
				return
			}
			requestModel = c.GetString(ctxkey.RequestModel)
			if requestModel != "" && !channel.ContainsModel(requestModel) {
				abortWithMessage(c, http.StatusForbidden, "该渠道不支持请求的模型")
				return
			}
		} else {
			requestModel = c.GetString(ctxkey.RequestModel)
			var err error
			channel, err = selectChannelViaRouter(c, userGroup, requestModel, userId, false)
			if err != nil {
				message := fmt.Sprintf("当前分组 %s 下对于模型 %s 无可用渠道", userGroup, requestModel)
				if channel != nil {
					logger.SysError(fmt.Sprintf("渠道不存在：%d", channel.Id))
					message = "数据库一致性已被破坏，请联系管理员"
				}
				abortWithMessage(c, http.StatusServiceUnavailable, message)
				return
			}
		}
		logger.Debugf(ctx, "user id %d, user group: %s, request model: %s, using channel #%d", userId, userGroup, requestModel, channel.Id)
		SetupContextForSelectedChannel(c, channel, requestModel)
		c.Next()
	}
}

// selectChannelViaRouter 经 channelrouter 选路（移植自 one-api-pro）：
// 粘性会话（可选）→ 冷却/并发/RPM 过滤 → 权重+优先级选择；
// router 未初始化或无候选时回退上游原生的随机选择。
func selectChannelViaRouter(c *gin.Context, group, requestModel string, userId int, ignoreFirstPriority bool) (*model.Channel, error) {
	if channelrouter.DefaultRouter == nil {
		return model.CacheGetRandomSatisfiedChannel(group, requestModel, ignoreFirstPriority)
	}
	candidates := model.GetChannelCandidates(group, requestModel)
	if len(candidates) == 0 {
		return model.CacheGetRandomSatisfiedChannel(group, requestModel, ignoreFirstPriority)
	}

	sessionKey := ""
	if config.ChannelStickySessionEnabled && userId > 0 && requestModel != "" {
		sessionKey = channelrouter.MakeSessionKey(userId, requestModel)
		c.Set(ctxkey.SessionKey, sessionKey)
	}

	req := &channelrouter.RouteRequest{
		Group:               group,
		Model:               requestModel,
		UserId:              userId,
		IgnoreFirstPriority: ignoreFirstPriority,
		SessionKey:          sessionKey,
	}
	return channelrouter.DefaultRouter.Route(c.Request.Context(), req, candidates)
}

func SetupContextForSelectedChannel(c *gin.Context, channel *model.Channel, modelName string) {
	c.Set(ctxkey.Channel, channel.Type)
	c.Set(ctxkey.ChannelId, channel.Id)
	c.Set(ctxkey.ChannelName, channel.Name)
	if channel.SystemPrompt != nil && *channel.SystemPrompt != "" {
		c.Set(ctxkey.SystemPrompt, *channel.SystemPrompt)
	}
	c.Set(ctxkey.ModelMapping, channel.GetModelMapping())
	c.Set(ctxkey.OriginalModel, modelName) // for retry
	c.Request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", channel.Key))
	c.Set(ctxkey.BaseURL, channel.GetBaseURL())
	cfg, _ := channel.LoadConfig()
	// this is for backward compatibility
	if channel.Other != nil {
		switch channel.Type {
		case channeltype.Azure:
			if cfg.APIVersion == "" {
				cfg.APIVersion = *channel.Other
			}
		case channeltype.Xunfei:
			if cfg.APIVersion == "" {
				cfg.APIVersion = *channel.Other
			}
		case channeltype.Gemini:
			if cfg.APIVersion == "" {
				cfg.APIVersion = *channel.Other
			}
		case channeltype.AIProxyLibrary:
			if cfg.LibraryID == "" {
				cfg.LibraryID = *channel.Other
			}
		case channeltype.Ali:
			if cfg.Plugin == "" {
				cfg.Plugin = *channel.Other
			}
		}
	}
	c.Set(ctxkey.Config, cfg)
}
