package main

import "crypto/hmac"
import "crypto/sha1"
import "strconv"
import "strings"

var (
	extraReplace = map[string]string{
		"/": "\\/",
		"<": "\\u003C",
		"%": "\\u0025",
		"@": "\\u0040",
	}
)

func getASCII(r rune) string {
	s := strconv.QuoteRuneToASCII(r)
	return s[1 : len(s)-1]
}

func facebookStr(str string) (ret string) {
	for _, r := range str {
		ret += getASCII(r)
	}
	for k, v := range extraReplace {
		ret = strings.Replace(ret, k, v, -1)
	}
	return ret
}

func checkSHA(sha string, body []byte) bool {
	expectedSum := []byte(facebookStr(sha[5:]))

	mac := hmac.New(sha1.New, []byte(cfg.appSecret))
	mac.Write(body)
	actualSum := mac.Sum(nil)

	return hmac.Equal(expectedSum, actualSum)
}
