package main

import "crypto/hmac"
import "crypto/sha1"
import "encoding/hex"
import "strconv"

func asciiRune(r rune) string {
	if r == '\\' {
		return "\\"
	}

	s := strconv.QuoteRuneToASCII(r)
	return s[1 : len(s)-1]
}

func asciiString(b []byte) (ret string) {
	for _, r := range string(b) {
		ret += asciiRune(r)
	}

	return ret
}

func checkSHA(sha string, body []byte) bool {
	expectedSum, err := hex.DecodeString(sha[5:])
	if err != nil {
		return false
	}

	mac := hmac.New(sha1.New, []byte(cfg.appSecret))
	mac.Write([]byte(asciiString(body)))
	actualSum := mac.Sum(nil)

	return hmac.Equal(expectedSum, actualSum)
}
