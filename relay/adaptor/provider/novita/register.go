package novita

import (
	"github.com/perryles/one-api-next/relay/adaptor"
	"github.com/perryles/one-api-next/relay/adaptor/openai"
	"github.com/perryles/one-api-next/relay/registry"
)

func init() {
	registry.Register(registry.ChannelMeta{
		ID:             "novita",
		Name:           "Novita",
		DefaultBaseURL: "https://api.novita.ai/v3/openai",
		LegacyType:     41,
	}, func() adaptor.Adaptor {
		return &Adaptor{Adaptor: &openai.Adaptor{}}
	})
}
