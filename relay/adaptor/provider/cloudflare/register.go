package cloudflare

import (
	"github.com/perryles/one-api-next/relay/adaptor"
	"github.com/perryles/one-api-next/relay/registry"
)

func init() {
	registry.Register(registry.ChannelMeta{
		ID:             "cloudflare",
		Name:           "Cloudflare",
		DefaultBaseURL: "https://api.cloudflare.com",
		LegacyType:     37,
	}, func() adaptor.Adaptor {
		return &Adaptor{}
	})
}
