package coze

import (
	"github.com/perryles/one-api-next/relay/adaptor"
	"github.com/perryles/one-api-next/relay/registry"
)

func init() {
	registry.Register(registry.ChannelMeta{
		ID:             "coze",
		Name:           "Coze",
		DefaultBaseURL: "https://api.coze.com",
		LegacyType:     34,
	}, func() adaptor.Adaptor {
		return &Adaptor{}
	})
}
