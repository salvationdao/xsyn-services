package helpers

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/microcosm-cc/bluemonday"
	"github.com/ninja-software/terror/v2"
	"github.com/volatiletech/null/v8"
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
