package xai

import (
	"github.com/perryles/one-api-next/relay/adaptor"
	"github.com/perryles/one-api-next/relay/adaptor/openai"
	"github.com/perryles/one-api-next/relay/registry"
)

func init() {
	registry.Register(registry.ChannelMeta{
		ID:             "xai",
		Name:           "xAI",
		DefaultBaseURL: "https://api.x.ai",
		LegacyType:     45,
	}, func() adaptor.Adaptor {
		return &Adaptor{Adaptor: &openai.Adaptor{}}
	})
}
