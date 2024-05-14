package txtparser

// Implement a universal parser for /etc/sysconfig files.

import (
	"bytes"
	"fmt"
	"github.com/SUSE/saptune/system"
	"regexp"
	"strconv"
	"strings"
)

var consecutiveSpaces = regexp.MustCompile("[[:space:]]+")

// SysconfigEntry contains a single key-value pair in sysconfig file.
type SysconfigEntry struct {
	LeadingComments []string // The comment lines leading to the key-value pair, including prefix '#', excluding end-of-line.
	Key             string   // The key.
	Value           string   // The value, excluding '=' character and double-quotes. Values will always come in double-quotes when converted to text.
}

// Sysconfig contains key-value pairs of a sysconfig file.
// It is able to convert back to original text in the original key order.
type Sysconfig struct {
	AllValues []*SysconfigEntry // All key-value pairs in the orignal order.
	KeyValue  map[string]*SysconfigEntry
}

// ParseSysconfigFile read sysconfig file and parse the file content into
// memory structures.
func ParseSysconfigFile(fileName string, autoCreate bool) (*Sysconfig, error) {
	content, err := system.ReadConfigFile(fileName, autoCreate)
	if err != nil {
		return nil, err
	}
	return ParseSysconfig(string(content))
}

// ParseSysconfig read sysconfig text and parse the text into memory structures.
func ParseSysconfig(input string) (*Sysconfig, error) {
	conf := &Sysconfig{
		AllValues: make([]*SysconfigEntry, 0),
		KeyValue:  make(map[string]*SysconfigEntry),
	}
	leadingComments := make([]string, 0)
	for _, line := range strings.Split(input, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "#") {
			// Line is a comment
			leadingComments = append(leadingComments, line)
		} else if eqChar := strings.IndexRune(line, '='); eqChar != -1 {
			// Line is a key-value pair
			// remove trailing comments from line
			line = system.StripComment(line, `\s#[^#]|"\s#[^#]`)
			key := strings.TrimSpace(line[0:eqChar])
			value := strings.Trim(strings.TrimSpace(line[eqChar+1:]), `"`)
			kv := &SysconfigEntry{
				LeadingComments: leadingComments,
				Key:             key,
				Value:           value,
			}
			conf.AllValues = append(conf.AllValues, kv)
			conf.KeyValue[key] = kv
			// Clear comments to be ready for the next key-value pair
			leadingComments = make([]string, 0)
		} else {
			// Consider other lines (such as blank lines) as comments
			leadingComments = append(leadingComments, line)
		}
	}
	return conf, nil
}

// Set value for a key. If the key does not yet exist, it is created.
func (conf *Sysconfig) Set(key string, value interface{}) {
	kv, exists := conf.KeyValue[key]
	if exists {
		kv.Value = fmt.Sprint(value)
	} else {
		kv = &SysconfigEntry{
			LeadingComments: nil,
			Key:             key,
			Value:           fmt.Sprint(value),
		}
		// When converted back into text, the new value will be appended at the end.
		conf.AllValues = append(conf.AllValues, kv)
	}
	conf.KeyValue[key] = kv
}

// SetIntArray give a space-separated integer array value to a key.
// If the key does not yet exist, it is created.
func (conf *Sysconfig) SetIntArray(key string, values []int) {
	strs := make([]string, len(values))
	for i, val := range values {
		strs[i] = strconv.Itoa(val)
	}
	conf.Set(key, strings.Join(strs, " "))
}

// SetStrArray give a space-separated string array value to a key.
// If the key does not yet exist, it is created.
func (conf *Sysconfig) SetStrArray(key string, values []string) {
	conf.Set(key, strings.Join(values, " "))
}

// GetInt return integer value that belongs to the key, or the default value
// if the key does not exist or value is not an integer.
func (conf *Sysconfig) GetInt(key string, defaultValue int) int {
	entry, exists := conf.KeyValue[key]
	if !exists {
		return defaultValue
	}
	intValue, err := strconv.Atoi(entry.Value)
	if err != nil {
		return defaultValue
	}
	return intValue
}

// GetUint64 return uint64 value that belongs to the key, or the default value
// if the key does not exist or value is not an integer.
func (conf *Sysconfig) GetUint64(key string, defaultValue uint64) uint64 {
	entry, exists := conf.KeyValue[key]
	if !exists {
		return defaultValue
	}
	intValue, err := strconv.ParseUint(entry.Value, 10, 64)
	if err != nil {
		return defaultValue
	}
	return intValue
}

// GetString return string value that belongs to the key, or the default value
// if the key does not exist.
func (conf *Sysconfig) GetString(key, defaultValue string) string {
	entry, exists := conf.KeyValue[key]
	if !exists || strings.TrimSpace(entry.Value) == "" {
		return defaultValue
	}
	return strings.TrimSpace(entry.Value)
}

// GetStringArray assume the key carries a space-separated array value,
// return the value array.
func (conf *Sysconfig) GetStringArray(key string, defaultValue []string) (ret []string) {
	entry, exists := conf.KeyValue[key]
	if !exists {
		return defaultValue
	}
	split := consecutiveSpaces.Split(strings.TrimSpace(entry.Value), -1)
	ret = make([]string, 0, len(split))
	for _, val := range split {
		if val != "" {
			ret = append(ret, val)
		}
	}
	return
}

// GetIntArray assume the key carries a space-separated array of integers,
// return the array. Discard malformed integers.
func (conf *Sysconfig) GetIntArray(key string, defaultValue []int) (ret []int) {
	entry, exists := conf.KeyValue[key]
	if !exists {
		return defaultValue
	}
	split := consecutiveSpaces.Split(strings.TrimSpace(entry.Value), -1)
	ret = make([]int, 0, len(split))
	for _, val := range split {
		iVal, err := strconv.Atoi(val)
		if err == nil {
			ret = append(ret, iVal)
		}
	}
	return
}

// GetBool return bool value that belongs to the key, or the default value
// if key does not exist.
// True values are "yes" or "true".
func (conf *Sysconfig) GetBool(key string, defaultValue bool) bool {
	defaultValStr := "no"
	if defaultValue {
		defaultValStr = "yes"
	}
	value := strings.ToLower(conf.GetString(key, defaultValStr))
	return (value == "yes" || value == "true")
}

// ToText convert key-value pairs back into text.
// Values are always surrounded by double-quotes.
func (conf *Sysconfig) ToText() string {
	var ret bytes.Buffer
	for _, kv := range conf.AllValues {
		if kv.LeadingComments != nil && len(kv.LeadingComments) > 0 {
			ret.WriteString(strings.Join(kv.LeadingComments, "\n"))
			ret.WriteRune('\n')
		}
		ret.WriteString(fmt.Sprintf("%s=\"%s\"\n", kv.Key, kv.Value))
	}
	return ret.String()
}

// IsKeyAvail return true, if the key is available.
// false, if the key does not exist.
func (conf *Sysconfig) IsKeyAvail(key string) bool {
	_, exists := conf.KeyValue[key]
	return exists
}
