package monitoring

import (
	"fmt"
	"strings"
)

const sparklineChars = "▁▂▃▄▅▆▇█"

// Sparkline renders a slice of float64 values as an ASCII/Unicode sparkline.
// Values are clamped to [min, max] and scaled across eight vertical steps.
func Sparkline(values []float64) string {
	if len(values) == 0 {
		return ""
	}

	chars := []rune(sparklineChars)

	min, max := values[0], values[0]
	for _, v := range values[1:] {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}

	if max == min {
		return strings.Repeat(string(chars[len(chars)/2]), len(values))
	}

	var b strings.Builder
	for _, v := range values {
		normalized := (v - min) / (max - min)
		idx := int(normalized * float64(len(chars)-1))
		if idx < 0 {
			idx = 0
		}
		if idx >= len(chars) {
			idx = len(chars) - 1
		}
		b.WriteRune(chars[idx])
	}

	return b.String()
}

// SparklineWithLabel renders a sparkline prefixed with a label and current value.
func SparklineWithLabel(label string, values []float64, current float64, unit string) string {
	line := Sparkline(values)
	if line == "" {
		return fmt.Sprintf("%s: no data", label)
	}
	return fmt.Sprintf("%s: %s  %.1f%s", label, line, current, unit)
}

// FormatBytes returns a human-readable byte string.
func FormatBytes(b uint64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := uint64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	prefixes := []string{"KiB", "MiB", "GiB", "TiB", "PiB"}
	return fmt.Sprintf("%.1f %s", float64(b)/float64(div), prefixes[exp])
}
