package relay

import (
	"github.com/perryles/one-api-next/relay/adaptor"
	"github.com/perryles/one-api-next/relay/registry"

	// Register all adaptors via init()
	_ "github.com/perryles/one-api-next/relay/adaptor/provider/ai360"
	_ "github.com/perryles/one-api-next/relay/adaptor/provider/aiproxy"
	_ "github.com/perryles/one-api-next/relay/adaptor/provider/ali"
	_ "github.com/perryles/one-api-next/relay/adaptor/provider/alibailian"
	_ "github.com/perryles/one-api-next/relay/adaptor/anthropic"
	_ "github.com/perryles/one-api-next/relay/adaptor/provider/aws"
	_ "github.com/perryles/one-api-next/relay/adaptor/provider/azure"
	_ "github.com/perryles/one-api-next/relay/adaptor/provider/baichuan"
	_ "github.com/perryles/one-api-next/relay/adaptor/provider/baidu"
	_ "github.com/perryles/one-api-next/relay/adaptor/provider/baiduv2"
	_ "github.com/perryles/one-api-next/relay/adaptor/provider/cloudflare"
	_ "github.com/perryles/one-api-next/relay/adaptor/provider/cohere"
	_ "github.com/perryles/one-api-next/relay/adaptor/provider/coze"
	_ "github.com/perryles/one-api-next/relay/adaptor/provider/deepl"
	_ "github.com/perryles/one-api-next/relay/adaptor/provider/deepseek"
	_ "github.com/perryles/one-api-next/relay/adaptor/provider/doubao"
	_ "github.com/perryles/one-api-next/relay/adaptor/provider/gemini"
	_ "github.com/perryles/one-api-next/relay/adaptor/provider/geminiv2"
	_ "github.com/perryles/one-api-next/relay/adaptor/provider/groq"
	_ "github.com/perryles/one-api-next/relay/adaptor/provider/lingyiwanwu"
	_ "github.com/perryles/one-api-next/relay/adaptor/provider/minimax"
	_ "github.com/perryles/one-api-next/relay/adaptor/provider/mistral"
	_ "github.com/perryles/one-api-next/relay/adaptor/provider/moonshot"
	_ "github.com/perryles/one-api-next/relay/adaptor/provider/novita"
	_ "github.com/perryles/one-api-next/relay/adaptor/provider/ollama"
	_ "github.com/perryles/one-api-next/relay/adaptor/openai"
	_ "github.com/perryles/one-api-next/relay/adaptor/provider/openrouter"
	_ "github.com/perryles/one-api-next/relay/adaptor/provider/palm"
	_ "github.com/perryles/one-api-next/relay/adaptor/provider/proxy"
	_ "github.com/perryles/one-api-next/relay/adaptor/provider/replicate"
	_ "github.com/perryles/one-api-next/relay/adaptor/provider/siliconflow"
	_ "github.com/perryles/one-api-next/relay/adaptor/provider/stepfun"
	_ "github.com/perryles/one-api-next/relay/adaptor/provider/tencent"
	_ "github.com/perryles/one-api-next/relay/adaptor/provider/togetherai"
	_ "github.com/perryles/one-api-next/relay/adaptor/provider/vertexai"
	_ "github.com/perryles/one-api-next/relay/adaptor/provider/xai"
	_ "github.com/perryles/one-api-next/relay/adaptor/provider/xunfei"
	_ "github.com/perryles/one-api-next/relay/adaptor/provider/xunfeiv2"
	_ "github.com/perryles/one-api-next/relay/adaptor/provider/zhipu"
)

func GetAdaptorByChannel(channelType int) adaptor.Adaptor {
	return registry.GetAdaptorByLegacyType(channelType)
}

func GetAdaptorByChannelID(channelID string) adaptor.Adaptor {
	return registry.GetAdaptor(channelID)
}

func GetAdaptor(apiType int) adaptor.Adaptor {
	return registry.GetAdaptorByLegacyType(apiType)
}
