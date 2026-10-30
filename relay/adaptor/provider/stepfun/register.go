package stepfun

import (
	"github.com/perryles/one-api-next/relay/adaptor"
	"github.com/perryles/one-api-next/relay/adaptor/openai"
	"github.com/perryles/one-api-next/relay/registry"
)

func init() {
	registry.Register(registry.ChannelMeta{
		ID:             "stepfun",
		Name:           "StepFun",
		DefaultBaseURL: "https://api.stepfun.com",
		LegacyType:     32,
	}, func() adaptor.Adaptor {
		return &Adaptor{Adaptor: &openai.Adaptor{}}
	})
}
