package divination

import (
	"regexp"
)

// 过滤掉无关数字
func contentFilter(content string) string {
	gex := regexp.MustCompile("[a-zA-Z0-9]");
	resultregex := gex.ReplaceAllString(content, "");
	gex = regexp.MustCompile("\\s+");
	resultregex = gex.ReplaceAllString(resultregex, "");
	gex = regexp.MustCompile(`\p{P}|\p{S}`);
	resultregex = gex.ReplaceAllString(resultregex, "");

	return content
}