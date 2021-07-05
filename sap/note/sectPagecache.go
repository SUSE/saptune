package note

import (
	"github.com/SUSE/saptune/system"
	"strconv"
	"strings"
)

// section [pagecache]

// GetPagecacheVal initialise the pagecache structure with the current
// system settings
func GetPagecacheVal(key string, cur *LinuxPagingImprovements) string {
	val := ""
	currentPagecache, err := LinuxPagingImprovements{PagingConfig: cur.PagingConfig}.Initialise()
	if err != nil {
		return ""
	}
	current := currentPagecache.(LinuxPagingImprovements)

	switch key {
	case "ENABLE_PAGECACHE_LIMIT":
		if current.VMPagecacheLimitMB == 0 {
			val = "no"
		} else {
			val = "yes"
		}
	case system.SysctlPagecacheLimitIgnoreDirty:
		val = strconv.Itoa(current.VMPagecacheLimitIgnoreDirty)
	case "OVERRIDE_PAGECACHE_LIMIT_MB":
		if current.VMPagecacheLimitMB == 0 {
			val = ""
		} else {
			val = strconv.FormatUint(current.VMPagecacheLimitMB, 10)
		}
	}
	*cur = current
	return val
}

// OptPagecacheVal optimises the pagecache structure with the settings
// from the configuration file or with a calculation
//func OptPagecacheVal(key, cfgval string, cur *LinuxPagingImprovements, keyvalue map[string]map[string]txtparser.INIEntry) string {
func OptPagecacheVal(key, cfgval string, cur *LinuxPagingImprovements) string {
	val := strings.ToLower(cfgval)

	switch key {
	case "ENABLE_PAGECACHE_LIMIT":
		if val != "yes" && val != "no" {
			system.WarningLog("wrong selection for ENABLE_PAGECACHE_LIMIT. Now set to default 'no'")
			val = "no"
		}
	case system.SysctlPagecacheLimitIgnoreDirty:
		if val != "2" && val != "1" && val != "0" {
			system.WarningLog("wrong selection for %s. Now set to default '1'", system.SysctlPagecacheLimitIgnoreDirty)
			val = "1"
		}
		cur.VMPagecacheLimitIgnoreDirty, _ = strconv.Atoi(val)
	case "OVERRIDE_PAGECACHE_LIMIT_MB":
		opt, _ := cur.Optimise()
		if opt == nil {
			_ = system.ErrorLog("page cache optimise had problems reading the Note definition file '%s'. Please check", cur.PagingConfig)
			return ""
		}
		optval := opt.(LinuxPagingImprovements).VMPagecacheLimitMB
		if optval != 0 {
			cur.VMPagecacheLimitMB = optval
			val = strconv.FormatUint(optval, 10)
		} else {
			cur.VMPagecacheLimitMB = 0
			val = ""
		}
	}

	return val
}

// SetPagecacheVal applies the settings to the system
func SetPagecacheVal(key string, cur *LinuxPagingImprovements) error {
	var err error
	if key == "OVERRIDE_PAGECACHE_LIMIT_MB" {
		err = cur.Apply()
	}
	return err
}
