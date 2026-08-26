package observation

import (
	"math"
	"testing"
)

// TestWrapResidualRANormalizesAcrossZeroBoundary 钉住跨 0°/360° 边界的赤经残差：
// 观测赤经≈0°、计算赤经≈360° 时，朴素相减得≈-360°，会被误判为相差近一圈，
// 归一化后应得到≈0° 的短残差。
func TestWrapResidualRANormalizesAcrossZeroBoundary(t *testing.T) {
	cases := []struct {
		name string
		in   float64
		want float64
	}{
		// 跨边界：观测≈0、计算≈360 -> 残差≈-360，应归零
		{"obs near 0 minus computed near 360", (0.001 - 359.999), 0.002},
		// 跨边界反向：观测≈360、计算≈0 -> 残差≈360，应归零
		{"obs near 360 minus computed near 0", (359.999 - 0.001), -0.002},
		// 正常小残差不受影响
		{"small positive", 5.0, 5.0},
		{"small negative", -5.0, -5.0},
		// 恰好过半圈，应取负侧短残差
		{"exactly 180 maps to -180", 180.0, -180.0},
		// 多圈偏移
		{"multiple wraps", 360.0 + 10.0, 10.0},
		{"multiple wraps negative", -360.0 - 10.0, -10.0},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := wrapResidualRA(c.in)
			if math.Abs(got-c.want) > 1e-9 {
				t.Fatalf("wrapResidualRA(%v) = %v, want %v", c.in, got, c.want)
			}
			// 契约：结果必须落在 [-180,180)
			if got < -180.0 || got >= 180.0 {
				t.Fatalf("wrapResidualRA(%v) = %v out of [-180,180)", c.in, got)
			}
		})
	}
}
