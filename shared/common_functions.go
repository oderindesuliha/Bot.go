// package shared

// import (
// 	"fmt"
// 	"time"
// )

// func FormatDuration(d time.Duration) string {
// 	if d.Hours() >= 24 {
// 		days := int(d.Hours() / 24)
// 		return formatWithUnit(days, "day")
// 	} else if d.Hours() >= 1 {
// 		hours := int(d.Hours())
// 		return formatWithUnit(hours, "hour")
// 	} else if d.Minutes() >= 1 {
// 		minutes := int(d.Minutes())
// 		return formatWithUnit(minutes, "minute")
// 	} else {
// 		seconds := int(d.Seconds())
// 		return formatWithUnit(seconds, "second")
// 	}
// }

// func formatWithUnit(value int, unit string) string {
// 	if value == 1 {
// 		return fmt.Sprintf("%d %s", value, unit)
// 	}
// 	return fmt.Sprintf("%d %ss", value, unit)
// }

// func ContainsString(slice []string, s string) bool {
// 	for _, item := range slice {
// 		if item == s {
// 			return true
// 		}
// 	}
// 	return false
// }

// func RemoveString(slice []string, s string) []string {
// 	result := make([]string, 0)
// 	for _, item := range slice {
// 		if item != s {
// 			result = append(result, item)
// 		}
// 	}
// 	return result
// }

// func FormatTimeLeft(endTime time.Time) string {
// 	duration := time.Until(endTime)
// 	if duration <= 0 {
// 		return "Closed"
// 	}

// 	days := int(duration.Hours()) / 24
// 	hours := int(duration.Hours()) % 24
// 	minutes := int(duration.Minutes()) % 60

// 	if days > 0 {
// 		return formatDays(days, hours)
// 	} else if hours > 0 {
// 		return formatHours(hours, minutes)
// 	} else {
// 		return formatMinutes(minutes)
// 	}
// }

// func formatDays(days, hours int) string {
// 	if days == 1 {
// 		if hours == 1 {
// 			return "1 day, 1 hour"
// 		} else if hours == 0 {
// 			return "1 day"
// 		} else {
// 			return "1 day, " + pluralize(hours, "hour")
// 		}
// 	} else {
// 		if hours == 1 {
// 			return pluralize(days, "day") + ", 1 hour"
// 		} else if hours == 0 {
// 			return pluralize(days, "day")
// 		} else {
// 			return pluralize(days, "day") + ", " + pluralize(hours, "hour")
// 		}
// 	}
// }

// func formatHours(hours, minutes int) string {
// 	if hours == 1 {
// 		if minutes == 1 {
// 			return "1 hour, 1 minute"
// 		} else if minutes == 0 {
// 			return "1 hour"
// 		} else {
// 			return "1 hour, " + pluralize(minutes, "minute")
// 		}
// 	} else {
// 		if minutes == 1 {
// 			return pluralize(hours, "hour") + ", 1 minute"
// 		} else if minutes == 0 {
// 			return pluralize(hours, "hour")
// 		} else {
// 			return pluralize(hours, "hour") + ", " + pluralize(minutes, "minute")
// 		}
// 	}
// }

// func formatMinutes(minutes int) string {
// 	if minutes == 1 {
// 		return "1 minute"
// 	} else {
// 		return pluralize(minutes, "minute")
// 	}
// }

// func pluralize(count int, word string) string {
// 	if count == 1 {
// 		return "1 " + word
// 	} else {
// 		return fmt.Sprintf("%d %ss", count, word)
// 	}
// }
