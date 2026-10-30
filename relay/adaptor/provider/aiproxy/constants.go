package aiproxy

import "github.com/perryles/one-api-next/relay/adaptor/openai"

var ModelList = []string{""}

func init() {
	ModelList = openai.ModelList
}
