package doubao

import (
	"github.com/perryles/one-api-next/relay/adaptor"
	"github.com/perryles/one-api-next/relay/adaptor/openai"
	"github.com/perryles/one-api-next/relay/registry"
)

func init() {
	registry.Register(registry.ChannelMeta{
		ID:             "doubao",
		Name:           "Doubao",
		DefaultBaseURL: "https://ark.cn-beijing.volces.com",
		LegacyType:     40,
	}, func() adaptor.Adaptor {
		return &Adaptor{Adaptor: &openai.Adaptor{}}
	})
}
