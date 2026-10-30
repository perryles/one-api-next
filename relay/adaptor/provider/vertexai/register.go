package vertexai

import (
	"github.com/perryles/one-api-next/relay/adaptor"
	"github.com/perryles/one-api-next/relay/registry"
)

func init() {
	registry.Register(registry.ChannelMeta{
		ID:             "vertexai",
		Name:           "Vertex AI",
		DefaultBaseURL: "",
		LegacyType:     42,
	}, func() adaptor.Adaptor {
		return &Adaptor{}
	})
}
