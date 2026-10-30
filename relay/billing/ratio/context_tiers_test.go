package ratio

import (
	"testing"

	"github.com/stretchr/testify/require"
)

const twoContextTiers = `{"context-model":[{"name":"standard","max_input_tokens":272000,"prompt_ratio":1,"completion_ratio":5},{"name":"long_context","max_input_tokens":0,"prompt_ratio":2,"completion_ratio":3.75}]}`

func TestContextTiersBoundaryScopeAndSnapshot(t *testing.T) {
	previous := ContextPriceTiersJSON()
	t.Cleanup(func() { require.NoError(t, UpdateContextPriceTiers(previous)) })
	require.NoError(t, UpdateContextPriceTiers(twoContextTiers))
	snapshot := GetContextPriceTiers("context-model", 1)
	for _, tc := range []struct {
		input int
		name  string
	}{
		{0, "standard"}, {271999, "standard"}, {272000, "standard"}, {272001, "long_context"}, {1000000, "long_context"},
	} {
		tier, ok := SelectContextPriceTier(snapshot, tc.input)
		require.True(t, ok)
		require.Equal(t, tc.name, tier.Name)
	}
	changed := GetContextPriceTiers("context-model", 1)
	changed[0].PromptRatio = 99
	require.Equal(t, 1.0, GetContextPriceTiers("context-model", 1)[0].PromptRatio)
	require.NoError(t, UpdateContextPriceTiers(twoContextTiers[:len(twoContextTiers)-1]+`,"context-model(1)":[]}`))
	require.Empty(t, GetContextPriceTiers("context-model", 1))
	require.Len(t, GetContextPriceTiers("context-model", 2), 2, "type-specific disabling must preserve the generic schedule for other channel types")
	require.Len(t, snapshot, 2, "in-flight price snapshot survives updates")
	require.Equal(t, 2.0, snapshot[1].PromptRatio)
}

func TestContextTiersRejectInvalidSchedulesWithoutReplacingPrices(t *testing.T) {
	previous := ContextPriceTiersJSON()
	t.Cleanup(func() { require.NoError(t, UpdateContextPriceTiers(previous)) })
	require.NoError(t, UpdateContextPriceTiers(twoContextTiers))
	for _, raw := range []string{
		`null`, `[]`, `{"x":null}`, `{"x":[{"name":"a","max_input_tokens":0,"prompt_ratio":1,"completion_ratio":1}]}`,
		`{"x":[{"name":"a","max_input_tokens":100,"prompt_ratio":1},{"name":"b","max_input_tokens":0,"prompt_ratio":2,"completion_ratio":1}]}`,
		`{"x":[{"name":"a","max_input_tokens":100,"prompt_ratio":1,"completion_ratio":1},{"name":"b","prompt_ratio":2,"completion_ratio":1}]}`,
		`{"x":[{"name":"a","max_input_tokens":0,"prompt_ratio":1,"completion_ratio":1},{"name":"b","max_input_tokens":100,"prompt_ratio":2,"completion_ratio":1}]}`,
		`{"x":[{"name":"a","max_input_tokens":100,"prompt_ratio":-1,"completion_ratio":1},{"name":"b","max_input_tokens":0,"prompt_ratio":2,"completion_ratio":1}]}`,
		`{"x":[{"name":"a","max_input_tokens":100,"prompt_ratio":1,"completion_ratio":1},{"name":"a","max_input_tokens":0,"prompt_ratio":2,"completion_ratio":1}]}`,
		`{"x":[{"name":"a","max_input_tokens":100,"prompt_ratio":1,"completion_ratio":1},{"name":"b","max_input_tokens":100,"prompt_ratio":2,"completion_ratio":1},{"name":"c","max_input_tokens":0,"prompt_ratio":3,"completion_ratio":1}]}`,
	} {
		require.Error(t, UpdateContextPriceTiers(raw), raw)
		require.Len(t, GetContextPriceTiers("context-model", 1), 2)
	}
}
