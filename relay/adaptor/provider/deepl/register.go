package deepl

import (
	"github.com/perryles/one-api-next/relay/adaptor"
	"github.com/perryles/one-api-next/relay/registry"
)

func init() {
	registry.Register(registry.ChannelMeta{
		ID:             "deepl",
		Name:           "DeepL",
		DefaultBaseURL: "https://api-free.deepl.com",
		LegacyType:     38,
	}, func() adaptor.Adaptor {
		return &Adaptor{}
	})
}
