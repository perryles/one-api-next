package aws

import (
	"github.com/perryles/one-api-next/relay/adaptor"
	"github.com/perryles/one-api-next/relay/registry"
)

func init() {
	registry.Register(registry.ChannelMeta{
		ID:             "aws_claude",
		Name:           "AWS Claude",
		DefaultBaseURL: "",
		LegacyType:     33,
	}, func() adaptor.Adaptor {
		return &Adaptor{}
	})
}
