package cohere

import (
	"github.com/perryles/one-api-next/relay/adaptor"
	"github.com/perryles/one-api-next/relay/registry"
)

func init() {
	registry.Register(registry.ChannelMeta{
		ID:             "cohere",
		Name:           "Cohere",
		DefaultBaseURL: "https://api.cohere.ai",
		LegacyType:     35,
	}, func() adaptor.Adaptor {
		return &Adaptor{}
	})
}
