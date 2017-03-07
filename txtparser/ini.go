package txtparser

import (
	"strings"
	"fmt"
	"io/ioutil"
	"regexp"
	"path"
)

const VENDOR_DIR = "/etc/saptune/extra/"

var consecSpaces = regexp.MustCompile("[[:space:]]+")

/*
handle vendor specific sysconfig files
see /etc/saptune/extra/HPE-Recommended_OS_settings.conf
*/
type IniEntry struct {
	Section string
	Key string
	Value string
}

type Iniconf struct {
	AllValues []*IniEntry
	KeyValue  map[string]*IniEntry
}

func ParseIniFile(iniFile string) (*Iniconf, error) {
	contentBytes, err := ioutil.ReadFile(path.Join(VENDOR_DIR, iniFile))
	if err != nil {
		fmt.Errorf("failed to read vendor file: %v", err)
		return nil, err
	}

	var tunable, value string
	var fstart = 0
	section := "[-]"
	cont := &Iniconf{
		AllValues: make([]*IniEntry, 0, 0),
		KeyValue:  make(map[string]*IniEntry),
	}

	for _, line := range strings.Split(string(contentBytes), "\n") {
		fields := consecSpaces.Split(strings.TrimSpace(line), -1)
		if len(fields) == 0 || len(fields[0]) == 0 || fields[0][0] == '#' {
			continue // skip comments and empty lines
		}
		if len(fields) < 3 { // handle tuning lines without spaces
			fields = strings.Split(strings.TrimSpace(line), "=")
			if len(fields) == 1 {
				if strings.HasPrefix(fields[0], "[") && strings.HasSuffix(fields[0], "]") {
					section = fields[0]
				}
				continue
			}
			fstart = 1
		} else {
			if fields[1] != "=" {
				continue
			}
			fstart = 2
		}
		value = fields[fstart]
		for i := fstart + 1; i < len(fields); i++ { // handle tunables with more than one value
			value = value + " " + fields[i]
		}
		tunable = fields[0]
		skv := &IniEntry{
			Section: section,
			Key:     tunable,
			Value:   value,
		}
		cont.AllValues = append(cont.AllValues, skv)
		cont.KeyValue[tunable] = skv
	}
	return cont, nil
}

