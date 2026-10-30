package xunfei

import (
	"github.com/perryles/one-api-next/relay/adaptor"
	"github.com/perryles/one-api-next/relay/registry"
)

func init() {
	registry.Register(registry.ChannelMeta{
		ID:             "xunfei",
		Name:           "Xunfei",
		DefaultBaseURL: "",
		LegacyType:     18,
	}, func() adaptor.Adaptor {
		return &Adaptor{}
	})
}
