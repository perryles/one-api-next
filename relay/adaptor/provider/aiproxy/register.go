package aiproxy

import (
	"github.com/perryles/one-api-next/relay/adaptor"
	"github.com/perryles/one-api-next/relay/registry"
)

func init() {
	registry.Register(registry.ChannelMeta{
		ID:             "aiproxylibrary",
		Name:           "AIProxy Library",
		DefaultBaseURL: "https://api.aiproxy.io",
		LegacyType:     21,
	}, func() adaptor.Adaptor {
		return &Adaptor{}
	})
}
