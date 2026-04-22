package service

import "testing"

func TestCdnFailureAbortReason(t *testing.T) {
	tests := []struct {
		name   string
		total  int
		failed int
		abort  bool
	}{
		// 没有文件：放行
		{"no_files", 0, 0, false},
		// 全成功：放行
		{"all_success", 10, 0, false},
		// 少量失败，未达绝对下限：放行
		{"few_failures_below_cap", 100, 3, false},
		// 满足"绝对下限 5" 但比例不够 10%（50 失败/1000 = 5%）—— 放行
		// 注意：要求比例 >= 10% AND 绝对数 >= 5。
		{"high_total_low_ratio", 1000, 5, false}, // 0.5%，比例不达
		// 满足比例但未达绝对下限（5 失败/50 = 10%，绝对 < 5 的等价边界：4 失败/40 = 10%）
		{"ratio_ok_abs_low", 40, 4, false}, // 绝对 4 < 5，放行
		// 比例和绝对数都达标：中止
		{"ratio_and_abs_over", 50, 5, true}, // 10%，绝对 5
		// 大量失败：中止
		{"half_fail", 20, 10, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := UploadResult{Total: tt.total}
			r.Failures = make([]UploadFailure, tt.failed)
			got := cdnFailureAbortReason(r)
			if (got != "") != tt.abort {
				t.Errorf("cdnFailureAbortReason(total=%d, failed=%d) = %q, want abort=%v",
					tt.total, tt.failed, got, tt.abort)
			}
		})
	}
}

func TestUploadResult_Shape(t *testing.T) {
	r := UploadResult{
		Total:   10,
		Success: 8,
		Failures: []UploadFailure{
			{Path: "a.png", Error: "boom"},
			{Path: "b.png", Error: "timeout"},
		},
	}
	if r.GetTotal() != 10 {
		t.Errorf("GetTotal = %d", r.GetTotal())
	}
	if len(r.GetFailures()) != 2 {
		t.Errorf("GetFailures len = %d", len(r.GetFailures()))
	}
}
