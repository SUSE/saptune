package txtparser

import (
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"strings"
)

const (
	OperatorLessThan = "<"
	OperatorMoreThan = ">"
	OperatorEqual    = "="
)

type Operator string // The comparison or assignment operator used in an INI file entry

var RegexKeyOperatorValue = regexp.MustCompile(`([\w._-]+)\s*([<=>])\s*(.*)`) // Break up a line into key, operator, value.

// A single key-value pair in INI file.
type INIEntry struct {
	Section  string
	Key      string
	Operator Operator
	Value    string
}

// All key-value pairs of an INI file.
type INIFile struct {
	AllValues []INIEntry
	KeyValue  map[string]map[string]INIEntry
}

func ParseINIFile(fileName string, autoCreate bool) (*INIFile, error) {
	content, err := ioutil.ReadFile(fileName)
	if os.IsNotExist(err) && autoCreate {
		err = os.MkdirAll(path.Dir(fileName), 0755)
		if err != nil {
			return nil, err
		}
		err = ioutil.WriteFile(fileName, []byte{}, 0644)
		content = []byte{}
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}
	return ParseINI(string(content)), nil
}

func ParseINI(input string) *INIFile {
	ret := &INIFile{
		AllValues: make([]INIEntry, 0, 64),
		KeyValue:  make(map[string]map[string]INIEntry),
	}

	currentSection := ""
	currentEntriesArray := make([]INIEntry, 0, 8)
	currentEntriesMap := make(map[string]INIEntry)
	for _, line := range strings.Split(input, "\n") {
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		if strings.HasPrefix(line, "#") {
			// Skip comments. Need to be done before
			// 'break apart the line into key, operator, value'
			// to support comments like # something (default = 60)
			// without side effects
			continue
		}

		if line[0] == '[' {
			// Save previous section
			if currentSection != "" {
				ret.KeyValue[currentSection] = currentEntriesMap
				ret.AllValues = append(ret.AllValues, currentEntriesArray...)
			}
			// Start a new section
			currentSection = line[1 : len(line)-1]
			currentEntriesArray = make([]INIEntry, 0, 8)
			currentEntriesMap = make(map[string]INIEntry)
			continue
		}
		// Break apart a line into key, operator, value.
		kov := RegexKeyOperatorValue.FindStringSubmatch(line)
		if kov == nil {
			// Skip comments, empty, and irregular lines.
			continue
		}
		// handle tunables with more than one value
		value := strings.Replace(kov[3], " ", "\t", -1)
		entry := INIEntry{
			Section:  currentSection,
			Key:      kov[1],
			Operator: Operator(kov[2]),
			Value:    value,
		}
		currentEntriesArray = append(currentEntriesArray, entry)
		currentEntriesMap[entry.Key] = entry
	}
	// Save last section
	if currentSection != "" {
		ret.KeyValue[currentSection] = currentEntriesMap
		ret.AllValues = append(ret.AllValues, currentEntriesArray...)
	}
	return ret
}
