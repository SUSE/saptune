// Implement a parser for /etc/security/limits.conf.
package system

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"
)

var consecutiveSpaces = regexp.MustCompile("[[:space:]]+")

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
	for _, line := range strings.Split(input, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "#") {
			// Line is a comment
			leadingComments = append(leadingComments, line)
		} else if fields := consecutiveSpaces.Split(line, -1); len(fields) == 4 {
			// Line is an entry
			entry := &SecLimitsEntry{
				LeadingComments: leadingComments,
				Domain:          fields[0],
				Type:            fields[1],
				Item:            fields[2],
				Value:           fields[3],
			}
			limits.Entries = append(limits.Entries, entry)
			// Get ready for the next entry by clearing comments
			leadingComments = make([]string, 0, 0)
		} else {
			// Consider other lines (such as blank lines) as comments
			leadingComments = append(leadingComments, line)
		}
	}
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

// Set value for an entry. If the entry does not yet exist, it is created.
func (limits *SecLimits) Set(domain, typeName, item, value string) {
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
		ret.WriteString(fmt.Sprintf("%s %s %s %s\n", entry.Domain, entry.Type, entry.Item, entry.Value))
	}
	return ret.String()
}

// Overwrite /etc/security/limits.conf with the content of this structure.
func (limits *SecLimits) Apply() error {
	return ioutil.WriteFile("/etc/security/limits.conf", []byte(limits.ToText()), 0644)
}
