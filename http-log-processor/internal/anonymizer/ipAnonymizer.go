package anonymizer

import (
	"strings"
)

func AnonymizeIP(ip string) string {
	parts := strings.Split(ip, ".")
	if len(parts) != 4 {
		return ip
	}
	parts[3] = "X"
	return strings.Join(parts, ".")
}
