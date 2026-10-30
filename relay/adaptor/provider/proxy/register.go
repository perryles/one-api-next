package proxy

import (
	"github.com/perryles/one-api-next/relay/adaptor"
	"github.com/perryles/one-api-next/relay/registry"
)

func init() {
	registry.Register(registry.ChannelMeta{
		ID:             "proxy",
		Name:           "Proxy",
		DefaultBaseURL: "",
		LegacyType:     43,
	}, func() adaptor.Adaptor {
		return &Adaptor{}
	})
}
