package ratio

import "testing"

// 未配置倍率的模型必须报告 known=false，让调用方拒绝计费；
// 不能再像历史实现那样静默按兜底值 30 收费。
// （机制与测试移植自 omnirelays，见 docs/UPSTREAM.md）
func TestGetModelRatioWithKnownReportsUnknownModels(t *testing.T) {
	const unknown = "one-api-next-model-without-ratio-2099"

	if _, known := GetModelRatioWithKnown(unknown, 1); known {
		t.Fatalf("未配置倍率的模型 %s 不应被报告为 known", unknown)
	}
	// 兜底值仍然保留，避免老调用方拿到 0 导致"免费"，
	// 但计费路径必须通过 known 判断拒绝服务。
	if got := GetModelRatio(unknown, 1); got != 30 {
		t.Fatalf("兜底倍率应为 30，实际 %v", got)
	}
}

// 表内已配置的模型必须报告 known=true 并返回正倍率。
func TestGetModelRatioWithKnownReportsConfiguredModels(t *testing.T) {
	for _, name := range []string{"gpt-4o", "gpt-4.1", "gpt-5", "claude-sonnet-4-5", "gemini-2.5-pro", "deepseek-chat", "glm-4.6", "kimi-k3", "qwen3-235b-a22b"} {
		ratio, known := GetModelRatioWithKnown(name, 1)
		if !known {
			t.Fatalf("模型 %s 应被报告为 known", name)
		}
		if ratio <= 0 {
			t.Fatalf("模型 %s 倍率应为正数，实际 %v", name, ratio)
		}
	}
}

// 渠道限定键 name(channelType) 优先于裸模型名。
func TestGetModelRatioWithKnownChannelQualifiedKeyWins(t *testing.T) {
	const qualified = "one-api-next-qualified-2099(33)"
	const bare = "one-api-next-bare-2099"
	modelRatioLock.Lock()
	ModelRatio[qualified] = 7
	ModelRatio[bare] = 3
	modelRatioLock.Unlock()
	defer func() {
		modelRatioLock.Lock()
		delete(ModelRatio, qualified)
		delete(ModelRatio, bare)
		modelRatioLock.Unlock()
	}()

	if ratio, known := GetModelRatioWithKnown("one-api-next-qualified-2099", 33); !known || ratio != 7 {
		t.Fatalf("渠道限定键未优先生效: known=%v ratio=%v", known, ratio)
	}
	if ratio, known := GetModelRatioWithKnown(bare, 34); !known || ratio != 3 {
		t.Fatalf("裸模型名应兜底生效: known=%v ratio=%v", known, ratio)
	}
}
