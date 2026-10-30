package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/perryles/one-api-next/common"
	"github.com/perryles/one-api-next/common/helper"
	"github.com/perryles/one-api-next/common/logger"
	"strings"
)

func abortWithMessage(c *gin.Context, statusCode int, message string) {
	if c.Request.URL.Path == "/v1/messages" || c.Request.URL.Path == "/v1/messages/count_tokens" {
		kind := "invalid_request_error"
		if statusCode == 401 {
			kind = "authentication_error"
		} else if statusCode == 403 {
			kind = "permission_error"
		} else if statusCode == 429 {
			kind = "rate_limit_error"
		} else if statusCode >= 500 {
			kind = "api_error"
		}
		c.AbortWithStatusJSON(statusCode, gin.H{"type": "error", "error": gin.H{"type": kind, "message": message}})
		return
	}
	c.JSON(statusCode, gin.H{
		"error": gin.H{
			"message": helper.MessageWithRequestId(message, c.GetString(helper.RequestIdKey)),
			"type":    "one_api_error",
		},
	})
	c.Abort()
	logger.Error(c.Request.Context(), message)
}

func getRequestModel(c *gin.Context) (string, error) {
	var modelRequest ModelRequest
	err := common.UnmarshalBodyReusable(c, &modelRequest)
	if err != nil {
		return "", fmt.Errorf("common.UnmarshalBodyReusable failed: %w", err)
	}
	if strings.HasPrefix(c.Request.URL.Path, "/v1/moderations") {
		if modelRequest.Model == "" {
			modelRequest.Model = "text-moderation-stable"
		}
	}
	if strings.HasSuffix(c.Request.URL.Path, "embeddings") {
		if modelRequest.Model == "" {
			modelRequest.Model = c.Param("model")
		}
	}
	if strings.HasPrefix(c.Request.URL.Path, "/v1/images/generations") {
		if modelRequest.Model == "" {
			modelRequest.Model = "dall-e-2"
		}
	}
	if strings.HasPrefix(c.Request.URL.Path, "/v1/audio/transcriptions") || strings.HasPrefix(c.Request.URL.Path, "/v1/audio/translations") {
		if modelRequest.Model == "" {
			modelRequest.Model = "whisper-1"
		}
	}
	return modelRequest.Model, nil
}

func isModelInList(modelName string, models string) bool {
	modelList := strings.Split(models, ",")
	for _, model := range modelList {
		if modelName == model {
			return true
		}
	}
	return false
}
