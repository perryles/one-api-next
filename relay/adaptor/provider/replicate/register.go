package replicate

import (
	"github.com/perryles/one-api-next/relay/adaptor"
	"github.com/perryles/one-api-next/relay/registry"
)

func init() {
	registry.Register(registry.ChannelMeta{
		ID:             "replicate",
		Name:           "Replicate",
		DefaultBaseURL: "https://api.replicate.com/v1/models/",
		LegacyType:     46,
	}, func() adaptor.Adaptor {
		return &Adaptor{}
	})
}
