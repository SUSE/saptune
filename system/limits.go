// Implement a parser for /etc/security/limits.conf.
package system

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"regexp"
	"strconv"
	"strings"
)

var consecutiveSpaces = regexp.MustCompile("[[:space:]]+")

type SecurityLimitInt int // SecurityLimitInt is an integer number where -1 represents unlimited value.

// SecurityLimitUnlimitedValue is the constant integer value that represents unrestricted limit.
const SecurityLimitUnlimitedValue = SecurityLimitInt(-1)

// SecurityLimitUnlimitedString are the string constants that represent unrestricted limit.
var SecurityLimitUnlimitedString = []string{"unlimited", "infinity"}

func (limit SecurityLimitInt) String() string {
	if limit == SecurityLimitUnlimitedValue {
		return SecurityLimitUnlimitedString[0]
	}
	return strconv.Itoa(int(limit))
}

/*
ToSecurityLimitInt interprets integer limit number from input string. If the input cannot be parsed successfully, it
will return a default 0 value.
*/
func ToSecurityLimitInt(in string) SecurityLimitInt {
	in = strings.TrimSpace(in)
	for _, match := range SecurityLimitUnlimitedString {
		if match == in {
			return SecurityLimitUnlimitedValue
		}
	}
	i, _ := strconv.Atoi(in)
	return SecurityLimitInt(i)
}

// A single entry in security/limits.conf file.
type SecLimitsEntry struct {
	LeadingComments    []string // The comment lines leading to the key-value pair, including prefix '#', excluding end-of-line.
	Domain, Type, Item string
	Value              string
}

// Entries of security/limits.conf file. It is able to convert back to original text in the original entry order.
type SecLimits struct {
	Entries []*SecLimitsEntry
}

// Read limits.conf and parse the file content into memory structures.
func ParseSecLimitsFile() (*SecLimits, error) {
	content, err := ioutil.ReadFile("/etc/security/limits.conf")
	if err != nil {
		return nil, fmt.Errorf("failed to open limits.conf: %v", err)
	}
	return ParseSecLimits(string(content)), nil
}

// Read limits.conf text and parse the text into memory structures.
func ParseSecLimits(input string) *SecLimits {
	limits := &SecLimits{Entries: make([]*SecLimitsEntry, 0, 0)}
	leadingComments := make([]string, 0, 0)
	splitInput := strings.Split(input, "\n")
	noOfLines := len(splitInput)
	for lineNo, line := range splitInput {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "#") {
			// Line is a comment
			leadingComments = append(leadingComments, line)
		} else if fields := consecutiveSpaces.Split(line, -1); len(fields) == 4 || len(fields) == 3 {
			val := ""
			if len(fields) == 4 {
				val = fields[3]
			}
			// Line is an entry
			entry := &SecLimitsEntry{
				LeadingComments: leadingComments,
				Domain:          fields[0],
				Type:            fields[1],
				Item:            fields[2],
				Value:           val,
			}
			limits.Entries = append(limits.Entries, entry)
			// Get ready for the next entry by clearing comments
			leadingComments = make([]string, 0, 0)
		} else {
			// Consider other lines (such as blank lines) as comments
			// seems that strings.Split(input, "\n") adds an additional new line to the split result, which should not end up in the resulting SecLimits structure
			if lineNo < (noOfLines - 1) {
				leadingComments = append(leadingComments, line)
			}
		}
	}
	// add the comment section. Needed, if the file only contains
	// comments, but no entries to not loose this comments
	entry := &SecLimitsEntry{
		LeadingComments: leadingComments,
	}
	limits.Entries = append(limits.Entries, entry)
	return limits
}

// Return string value that belongs to the entry.
func (limits *SecLimits) Get(domain, typeName, item string) (string, bool) {
	for _, entry := range limits.Entries {
		if entry.Domain == domain && entry.Type == typeName && entry.Item == item {
			return entry.Value, true
		}
	}
	return "0", false
}

/*
GetOrUnlimited retrieves an integer limit value and return. If the value is not specified or cannot be parsed correctly,
the 0 value will be returned.
*/
func (limits *SecLimits) GetOr0(domain, typeName, item string) SecurityLimitInt {
	val, _ := limits.Get(domain, typeName, item)
	return ToSecurityLimitInt(val)
}

// Set value for an entry. If the entry does not yet exist, it is created.
func (limits *SecLimits) Set(domain, typeName, item, value string) {
	if value == "0" {
		value = ""
	}
	for _, entry := range limits.Entries {
		if entry.Domain == domain && entry.Type == typeName && entry.Item == item {
			entry.Value = value
			return
		}
	}
	// Create a new entry
	limits.Entries = append(limits.Entries, &SecLimitsEntry{
		Domain: domain,
		Type:   typeName,
		Item:   item,
		Value:  value,
	})
}

// Convert the entries back into text.
func (limits *SecLimits) ToText() string {
	var ret bytes.Buffer
	for _, entry := range limits.Entries {
		if entry.LeadingComments != nil && len(entry.LeadingComments) > 0 {
			ret.WriteString(strings.Join(entry.LeadingComments, "\n"))
			ret.WriteRune('\n')
		}
		// prevent useless empty lines
		if entry.Domain != "" && entry.Type != "" && entry.Item != "" && entry.Value != "" {
			ret.WriteString(fmt.Sprintf("%s %s %s %s\n", entry.Domain, entry.Type, entry.Item, entry.Value))
		}
	}
	return ret.String()
}

// Overwrite /etc/security/limits.conf with the content of this structure.
func (limits *SecLimits) Apply() error {
	return ioutil.WriteFile("/etc/security/limits.conf", []byte(limits.ToText()), 0644)
}
