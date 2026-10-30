package anthropic

import (
	"github.com/perryles/one-api-next/relay/adaptor"
	"github.com/perryles/one-api-next/relay/registry"
)

func init() {
	registry.Register(registry.ChannelMeta{
		ID:             "anthropic",
		Name:           "Anthropic",
		DefaultBaseURL: "https://api.anthropic.com",
		LegacyType:     14,
	}, func() adaptor.Adaptor {
		return &Adaptor{}
	})
}
