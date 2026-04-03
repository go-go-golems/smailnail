package enrich

import (
	"encoding/json"
	"regexp"
	"strings"
)

var unsubscribeLinkPattern = regexp.MustCompile(`<([^>]+)>`)

func GetHeader(headersJSON, name string) string {
	if strings.TrimSpace(headersJSON) == "" {
		return ""
	}

	headers := map[string]string{}
	if err := json.Unmarshal([]byte(headersJSON), &headers); err != nil {
		return ""
	}

	for key, value := range headers {
		if strings.EqualFold(key, name) {
			return value
		}
	}

	return ""
}

func ParseListUnsubscribe(raw string, postHeader ...string) (string, string, bool) {
	var mailto string
	var httpURL string

	for _, match := range unsubscribeLinkPattern.FindAllStringSubmatch(raw, -1) {
		if len(match) < 2 {
			continue
		}
		uri := strings.TrimSpace(match[1])
		switch {
		case strings.HasPrefix(strings.ToLower(uri), "mailto:"):
			mailto = uri
		case strings.HasPrefix(strings.ToLower(uri), "https://"), strings.HasPrefix(strings.ToLower(uri), "http://"):
			httpURL = uri
		}
	}

	oneClick := strings.Contains(strings.ToLower(raw), "list-unsubscribe=one-click")
	for _, rawPost := range postHeader {
		if strings.Contains(strings.ToLower(rawPost), "list-unsubscribe=one-click") {
			oneClick = true
			break
		}
	}

	return mailto, httpURL, oneClick
}
