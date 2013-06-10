package plate

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"
)

func WebTime(t time.Time) string {
	ftime := t.Format(time.RFC1123)
	if strings.HasSuffix(ftime, "UTC") {
		ftime = ftime[0:len(ftime)-3] + "GMT"
	}
	return ftime
}

func DirExists(dir string) bool {
	d, e := os.Stat(dir)
	switch {
	case e != nil:
		return false
	case !d.IsDir():
		return false
	}

	return true
}

func FileExists(dir string) bool {
	info, err := os.Stat(dir)
	if err != nil {
		return false
	}

	return !info.IsDir()
}

// Urlencode is a helper method that converts a map into URL-encoded form data.
// It is a useful when constructing HTTP POST requests.
func Urlencode(data map[string]string) string {
	var buf bytes.Buffer
	for k, v := range data {
		buf.WriteString(url.QueryEscape(k))
		buf.WriteByte('=')
		buf.WriteString(url.QueryEscape(v))
		buf.WriteByte('&')
	}
	s := buf.String()
	return s[0 : len(s)-1]
}

var slugRegex = regexp.MustCompile(`(?i:[^a-z0-9\-_])`)

// Slug is a helper function that returns the URL slug for string s.
// It's used to return clean, URL-friendly strings that can be
// used in routing.
func Slug(s string, sep string) string {
	if s == "" {
		return ""
	}
	slug := slugRegex.ReplaceAllString(s, sep)
	if slug == "" {
		return ""
	}
	quoted := regexp.QuoteMeta(sep)
	sepRegex := regexp.MustCompile("(" + quoted + "){2,}")
	slug = sepRegex.ReplaceAllString(slug, sep)
	sepEnds := regexp.MustCompile("^" + quoted + "|" + quoted + "$")
	slug = sepEnds.ReplaceAllString(slug, "")
	return strings.ToLower(slug)
}

// NewCookie is a helper method that returns a new http.Cookie object.
// Duration is specified in seconds. If the duration is zero, the cookie is permanent.
// This can be used in conjunction with ctx.SetCookie.
func NewCookie(name string, value string, age int64) *http.Cookie {
	var utctime time.Time
	if age == 0 {
		// 2^31 - 1 seconds (roughly 2038)
		utctime = time.Unix(2147483647, 0)
	} else {
		utctime = time.Unix(time.Now().Unix()+age, 0)
	}
	return &http.Cookie{Name: name, Value: value, Expires: utctime}
}

func Substr(s string, start, length int) string {
	bt := []rune(s)
	if start < 0 {
		start = 0
	}
	var end int
	if (start + length) > (len(bt) - 1) {
		end = len(bt) - 1
	} else {
		end = start + length
	}
	return string(bt[start:end])
}

// Html2str() returns escaping text convert from html
func Html2str(html string) string {
	src := string(html)

	// Lower HTML
	re, _ := regexp.Compile("\\<[\\S\\s]+?\\>")
	src = re.ReplaceAllStringFunc(src, strings.ToLower)

	//Style
	re, _ = regexp.Compile("\\<style[\\S\\s]+?\\</style\\>")
	src = re.ReplaceAllString(src, "")

	//Script
	re, _ = regexp.Compile("\\<script[\\S\\s]+?\\</script\\>")
	src = re.ReplaceAllString(src, "")

	// HTML
	re, _ = regexp.Compile("\\<[\\S\\s]+?\\>")
	src = re.ReplaceAllString(src, "\n")

	// String
	re, _ = regexp.Compile("\\s{2,}")
	src = re.ReplaceAllString(src, "\n")

	return strings.TrimSpace(src)
}

// DateFormat takes a time and a layout string and returns a string with the formatted date. Used by the template parser as "dateformat"
func DateFormat(t time.Time, layout string) (datestring string) {
	datestring = t.Format(layout)
	return
}

// Date takes a PHP like date func to Go's time fomate
func Date(t time.Time, format string) (datestring string) {
	patterns := []string{
		// year
		"Y", "2006", // A full numeric representation of a year, 4 digits	Examples: 1999 or 2003
		"y", "06", //A two digit representation of a year	Examples: 99 or 03

		// month
		"m", "01", // Numeric representation of a month, with leading zeros	01 through 12
		"n", "1", // Numeric representation of a month, without leading zeros	1 through 12
		"M", "Jan", // A short textual representation of a month, three letters	Jan through Dec
		"F", "January", // A full textual representation of a month, such as January or March	January through December

		// day
		"d", "02", // Day of the month, 2 digits with leading zeros	01 to 31
		"j", "2", // Day of the month without leading zeros	1 to 31

		// week
		"D", "Mon", // A textual representation of a day, three letters	Mon through Sun
		"l", "Monday", // A full textual representation of the day of the week	Sunday through Saturday

		// time
		"g", "3", // 12-hour format of an hour without leading zeros	1 through 12
		"G", "15", // 24-hour format of an hour without leading zeros	0 through 23
		"h", "03", // 12-hour format of an hour with leading zeros	01 through 12
		"H", "15", // 24-hour format of an hour with leading zeros	00 through 23

		"a", "pm", // Lowercase Ante meridiem and Post meridiem	am or pm
		"A", "PM", // Uppercase Ante meridiem and Post meridiem	AM or PM

		"i", "04", // Minutes with leading zeros	00 to 59
		"s", "05", // Seconds, with leading zeros	00 through 59
	}
	replacer := strings.NewReplacer(patterns...)
	format = replacer.Replace(format)
	datestring = t.Format(format)
	return
}

// Compare is a quick and dirty comparison function. It will convert whatever you give it to strings and see if the two values are equal.
// Whitespace is trimmed. Used by the template parser as "eq"
func Compare(a, b interface{}) (equal bool) {
	equal = false
	if strings.TrimSpace(fmt.Sprintf("%v", a)) == strings.TrimSpace(fmt.Sprintf("%v", b)) {
		equal = true
	}
	return
}

func Str2html(raw string) template.HTML {
	return template.HTML(raw)
}

// Encodes `text` for raw use in HTML.
// >>> htmlquote("<'&\\">")
// 		'&lt;&#39;&amp;&quot;&gt;'
func Htmlquote(src string) string {
	text := string(src)

	text = strings.Replace(text, "&", "&amp;", -1) // Must be done first!
	text = strings.Replace(text, "<", "&lt;", -1)
	text = strings.Replace(text, ">", "&gt;", -1)
	text = strings.Replace(text, "'", "&#39;", -1)
	text = strings.Replace(text, "\"", "&quot;", -1)
	text = strings.Replace(text, "“", "&ldquo;", -1)
	text = strings.Replace(text, "”", "&rdquo;", -1)
	text = strings.Replace(text, " ", "&nbsp;", -1)

	return strings.TrimSpace(text)
}

// Decodes `text` that's HTML quoted.
// >>> HtmlUnquote('&lt;&#39;&amp;&quot;&gt;')
//	       '<\\'&">'
func Htmlunquote(src string) string {
	// strings.Replace(s, old, new, n)

	text := string(src)
	text = strings.Replace(text, "&nbsp;", " ", -1)
	text = strings.Replace(text, "&rdquo;", "”", -1)
	text = strings.Replace(text, "&ldquo;", "“", -1)
	text = strings.Replace(text, "&quot;", "\"", -1)
	text = strings.Replace(text, "&#39;", "'", -1)
	text = strings.Replace(text, "&gt;", ">", -1)
	text = strings.Replace(text, "&lt;", "<", -1)
	text = strings.Replace(text, "&amp;", "&", -1) // Must be done last!

	return strings.TrimSpace(text)
}
