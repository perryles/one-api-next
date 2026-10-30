package ollama

import (
	"github.com/perryles/one-api-next/relay/adaptor"
	"github.com/perryles/one-api-next/relay/registry"
)

func init() {
	registry.Register(registry.ChannelMeta{
		ID:             "ollama",
		Name:           "Ollama",
		DefaultBaseURL: "http://localhost:11434",
		LegacyType:     30,
	}, func() adaptor.Adaptor {
		return &Adaptor{}
	})
}
