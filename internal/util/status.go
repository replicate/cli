package util

import "github.com/replicate/replicate-go"

func StatusSymbol(status replicate.Status) string {
	switch status {
	case replicate.Starting:
		return "⚪️"
	case replicate.Processing:
		return "🟡"
	case replicate.Failed:
		return "🔴"
	case replicate.Succeeded:
		return "🟢"
	case replicate.Canceled:
		return "🔵"
	default:
		return string(status)
	}
}
