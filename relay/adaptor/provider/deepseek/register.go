package deepseek

import (
	"github.com/perryles/one-api-next/relay/adaptor"
	"github.com/perryles/one-api-next/relay/registry"
)

func init() {
	registry.Register(registry.ChannelMeta{
		ID:             "deepseek",
		Name:           "DeepSeek",
		DefaultBaseURL: "https://api.deepseek.com",
		LegacyType:     36,
	}, func() adaptor.Adaptor {
		return &Adaptor{}
	})
}
