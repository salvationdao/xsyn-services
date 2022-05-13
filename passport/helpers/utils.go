package helpers

import (
	"encoding/json"
	"github.com/microcosm-cc/bluemonday"
	"github.com/ninja-software/terror/v2"
	"github.com/volatiletech/null/v8"
	"math/rand"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"unicode"
	"xsyn-services/types"
)

// EncodeJSON will encode json to response writer and return status ok.
func EncodeJSON(w http.ResponseWriter, result interface{}) (int, error) {
	err := json.NewEncoder(w).Encode(result)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "")
	}
	return http.StatusOK, nil
}

// SanitiseNullString creates a null.String with sanitisation
func SanitiseNullString(s string, sp *bluemonday.Policy) null.String {
	return null.StringFrom(sp.Sanitize(s))
}

// ParseQueryText parse the search text for full text search
func ParseQueryText(queryText string) string {
	// sanity check
	if queryText == "" {
		return ""
	}

	// trim leading and trailing spaces
	re2 := regexp.MustCompile(`\\s+`)
	keywords := strings.TrimSpace(queryText)
	// to lowercase
	keywords = strings.ToLower(keywords)
	// remove excess spaces
	keywords = re2.ReplaceAllString(keywords, " ")
	// no non-alphanumeric
	re := regexp.MustCompile(`[^a-z0-9\\-\\. ]`)
	keywords = re.ReplaceAllString(keywords, "")

	// keywords array
	xkeywords := strings.Split(keywords, " ")
	// for sql construction
	var keywords2 []string
	// build sql keywords
	for _, keyword := range xkeywords {
		// skip blank, to prevent error on construct sql search
		if len(keyword) == 0 {
			continue
		}

		// add prefix for partial word search
		keyword = keyword + ":*"
		// add to search string queue
		keywords2 = append(keywords2, keyword)
	}
	// construct sql search
	xsearch := strings.Join(keywords2, " & ")

	return xsearch

}

// SplitOnCaps splits and joins string on capital letters
// eg. (ThisIsFormType) becomes (This Is Form Type)
func SplitOnCaps(s string) string {
	re := regexp.MustCompile(`[A-Z][^A-Z]*`) // regex to split string on capital letters
	return strings.Join(re.FindAllString(s, -1), " ")
}

// SearchArgInt returns and converts a URL search argument to a int
func SearchArgInt(r *http.Request, key string) *int {
	str := r.URL.Query().Get(key)
	if str == "" {
		return nil
	}
	i, err := strconv.Atoi(str)
	if err != nil {
		return nil
	}
	return &i
}

// SearchArg returns URL search argument as a string pointer (returns nil if empty string)
func SearchArg(r *http.Request, key string) *string {
	str := r.URL.Query().Get(key)
	if str == "" {
		return nil
	}
	return &str
}

// StringPointer converts a string to a string pointer
func StringPointer(str string) *string {
	return &str
}

// FloatPointer converts a float32 to a float32 pointer
func FloatPointer(value float32) *float32 {
	return &value
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func RandStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

// TrimUsername removes misuse of invisible characters.
func TrimUsername(username string) string {
	// Check if entire string is nothing not non-printable characters
	isEmpty := true
	runes := []rune(username)
	for _, r := range runes {
		if unicode.IsPrint(r) && !unicode.IsSpace(r) {
			isEmpty = false
			break
		}
	}
	if isEmpty {
		return ""
	}

	// Remove Spaces like characters Around String (keep mark ones)
	output := strings.Trim(username, " \u00A0\u180E\u2000\u2001\u2002\u2003\u2004\u2005\u2006\u2007\u2008\u2009\u200A\u200B\u202F\u205F\u3000\uFEFF\u2423\u2422\u2420")

	// Enforce one Space like characters between words
	output = strings.Join(strings.Fields(output), " ")

	return output
}

func CheckAddressIsLocked(level string, user *types.User) bool {
	if level == "withdrawals" && user.WithdrawLock {
		return true
	}

	if level == "minting" && user.MintLock {
		return true
	}

	if level == "account" && user.TotalLock {
		return true
	}

	return false
}
