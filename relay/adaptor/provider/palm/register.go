package palm

import (
	"github.com/perryles/one-api-next/relay/adaptor"
	"github.com/perryles/one-api-next/relay/registry"
)

func init() {
	registry.Register(registry.ChannelMeta{
		ID:             "palm",
		Name:           "PaLM",
		DefaultBaseURL: "https://generativelanguage.googleapis.com",
		LegacyType:     11,
	}, func() adaptor.Adaptor {
		return &Adaptor{}
	})
}
