package cron

import (
	"fmt"
	"time"
)

func durationToCron(dur time.Duration) string {
	minutes := int(dur.Minutes())

	switch {
	case minutes < 60:
		return fmt.Sprintf("*/%d * * * *", minutes)

	case minutes%60 == 0 && minutes < 60*24:
		hours := minutes / 60
		return fmt.Sprintf("0 */%d * * *", hours)

	case minutes%(60*24) == 0:
		days := minutes / (60 * 24)
		return fmt.Sprintf("0 0 */%d * *", days)

	default:
		return ""
	}
}
