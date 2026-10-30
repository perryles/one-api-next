package gemini

import (
	"github.com/perryles/one-api-next/relay/adaptor"
	"github.com/perryles/one-api-next/relay/registry"
)

func init() {
	registry.Register(registry.ChannelMeta{
		ID:             "gemini",
		Name:           "Gemini",
		DefaultBaseURL: "https://generativelanguage.googleapis.com",
		LegacyType:     24,
	}, func() adaptor.Adaptor {
		return &Adaptor{}
	})
}
