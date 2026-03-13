package txtparser

import (
	"fmt"
	"github.com/SUSE/saptune/system"
	"regexp"
	"runtime"
	"strconv"
	"strings"
)

// isTagAvail checks, if a special tag is available in the section Fields
func isTagAvail(tag string, secFields []string) bool {
	cnt := 0
	for _, secTag := range secFields {
		if cnt == 0 {
			// skip section name
			cnt = cnt + 1
			continue
		}
		tagField := strings.Split(secTag, "=")
		if len(tagField) != 2 {
			return false
		}
		if tag == tagField[0] {
			return true
		}
	}
	return false
}

// chkSecTags checks, if the tags of a section are valid
func chkSecTags(secFields, blkDev []string) (bool, []string) {
	ret := true
	cnt := 0
	for _, secTag := range secFields {
		if cnt == 0 {
			// skip section name
			cnt = cnt + 1
			continue
		}
		if secTag == "" {
			// support empty tags
			continue
		}
		tagField := strings.Split(secTag, "=")
		if len(tagField) != 2 {
			system.WarningLog("wrong syntax of section tag '%s', skipping whole section '%v'. Please check. ", secTag, secFields)
			return false, blkDev
		}
		switch tagField[0] {
		case "os":
			ret = chkOsTags(tagField[1], secFields)
		case "arch":
			ret = chkArchTags(tagField[1], secFields)
		case "csp":
			ret = chkCspTags(tagField[1], secFields)
		case "virt":
			ret = chkVirtTags(tagField[1], secFields)
		case "blkvendor", "blkmodel", "blkpat":
			ret, blkDev = chkBlkTags(tagField[0], tagField[1], secFields, blkDev)
		case "vendor", "model":
			ret = chkHWTags(tagField[0], tagField[1], secFields)
		case "pmu_name":
			ret = chkCPUTags(tagField[1], secFields)
		default:
			ret = chkOtherTags(tagField[0], tagField[1], secFields)
		}
		if !ret {
			break
		}
	}
	return ret, blkDev
}

// chkOsTags checks if the os section tag is valid or not
func chkOsTags(tagField string, secFields []string) bool {
	ret := true
	osWild := regexp.MustCompile(`(.*)[-\.]S?P?(\*)`)
	osw := osWild.FindStringSubmatch(tagField)
	if len(osw) != 3 {
		// check for multiple os releases
		return chkMultiRelease(tagField, secFields)
	}
	// len == 3
	if osw[2] == "*" {
		// wildcard 15-* or 15.*
		// check for supported os version
		if osw[1] != "12" && osw[1] != "15" && osw[1] != "16" {
			system.InfoLog("unsupported os version '%s' in section definition '%v'. Skipping whole section with all lines till next valid section definition", osw[1], secFields)
			ret = false
		}
		// check runing os version
		if !system.IsSLE(osw[1]) {
			system.InfoLog("os version '%s' in section definition '%v' does not match running os version '%s'. Skipping whole section with all lines till next valid section definition", tagField, secFields, system.GetOsVers())
			ret = false
		}
	} else {
		// wrong syntax
		system.InfoLog("wrong syntax for tag in section definition '%v'.  Skipping whole section with all lines till next valid section definition", secFields)
		ret = false
	}
	return ret
}

// chkMultiRelease checks if the os section tag contains multiple os releases
func chkMultiRelease(tagField string, secFields []string) bool {
	// 15-[2,4-5,7-]
	osMultiSPs := regexp.MustCompile(`(.*)[-\.]\[(.*)\]`)
	osmsps := osMultiSPs.FindStringSubmatch(tagField)
	if len(osmsps) > 1 {
		// check for supported os version
		if osmsps[1] != "12" && osmsps[1] != "15" && osmsps[1] != "16" {
			system.InfoLog("unsupported os version '%s' in section definition '%v'. Skipping whole section with all lines till next valid section definition", osmsps[1], secFields)
			return false
		}
	}
	if len(osmsps) != 3 {
		// len == -> 15-SP6, 16.0
		if tagField != system.GetOsVers() {
			// os version does not match
			system.InfoLog("os version '%s' in section definition '%v' does not match running os version '%s'. Skipping whole section with all lines till next valid section definition", tagField, secFields, system.GetOsVers())
			return false
		}
		return true
	}
	// len == 3 - check releases
	return chkRelease(osmsps[2], secFields)
}

// chkRelease checks if the os section tag contains a release matching the
// running system
func chkRelease(relList string, secFields []string) bool {
	ret := false
	chkrel, _ := strconv.Atoi(system.GetOsRel())
	// [2,4-5,7-]
	for _, rel := range strings.Split(relList, ",") {
		relrange := strings.Split(rel, "-")
		if len(relrange) == 1 {
			// [2]
			if rel != system.GetOsRel() {
				system.InfoLog("os release '%s' in section definition '%v' does not match running os release '%s'. Skipping whole section with all lines till next valid section definition", rel, secFields, system.GetOsRel())
				continue
			} else {
				ret = true
				break
			}
		}

		if relrange[0] == "" && relrange[1] == "" {
			// wrong syntax [-]
			system.InfoLog("wrong syntax for tag in section definition '%v'.  Skipping whole section with all lines till next valid section definition", secFields)
			continue
		}
		if relrange[0] == "" {
			//[-4] --> SP0, SP1, SP2, SP3, SP4
			chkrelrng, _ := strconv.Atoi(relrange[1])
			if chkrelrng >= chkrel {
				ret = true
				break
			}
		}
		if relrange[1] == "" {
			// [5-] --> SP5, SP6, SP7
			chkrelrng, _ := strconv.Atoi(relrange[0])
			if chkrelrng <= chkrel {
				ret = true
				break
			}
		}
		if relrange[0] != "" && relrange[1] != "" {
			// [4-6] --> SP4, SP5, SP6
			relstart, _ := strconv.Atoi(relrange[0])
			relend, _ := strconv.Atoi(relrange[1])
			if relstart <= chkrel && relend >= chkrel {
				ret = true
				break
			}
		}
	}
	return ret
}

// chkArchTags checks if the arch section tag is valid or not
func chkArchTags(tagField string, secFields []string) bool {
	ret := true
	chkArch := runtime.GOARCH
	if chkArch == "amd64" {
		// map architecture to 'uname -i' output
		chkArch = "x86_64"
	}
	if tagField != chkArch {
		// arch does not match
		system.InfoLog("system architecture '%s' in section definition '%v' does not match the architecture of the running system '%s'. Skipping whole section with all lines till next valid section definition", tagField, secFields, chkArch)
		ret = false
	}
	return ret
}

// chkCsp checks if the csp section tag is valid or not
func chkCspTags(tagField string, secFields []string) bool {
	ret := true
	chkCsp := system.GetCSP()
	if tagField != chkCsp {
		// csp does not match
		if chkCsp == "" {
			chkCsp = "not a cloud"
		}
		system.InfoLog("cloud service provider '%s' in section definition '%v' does not match the cloud service provider of the running system ('%s'). Skipping whole section with all lines till next valid section definition", tagField, secFields, chkCsp)
		ret = false
	}
	return ret
}

// chkVirtTags checks if the virtualization type section tag is valid or not
func chkVirtTags(tagField string, secFields []string) bool {
	ret := true
	vopt := ""
	switch tagField {
	case "vm":
		vopt = "-v"
	case "chroot":
		vopt = "-r"
	case "container":
		vopt = "-c"
	}
	virt, chkVirt, err := system.SystemdDetectVirt(vopt)
	if vopt != "" {
		if !virt {
			ret = false
			system.InfoLog("virtualization class type '%s' in section definition '%v' does not match the virtualization class type of the running system. Skipping whole section with all lines till next valid section definition", tagField, secFields)
		}
		return ret
	}
	// order of vopt and err check is by intention
	if err != nil {
		system.InfoLog("No virtualization detected - error with systemd-detect-virt. Skipping whole section '%v' with all lines till next valid section definition", secFields)
		return false
	}
	if tagField != chkVirt {
		// virtualization type does not match
		system.InfoLog("virtualization type '%s' in section definition '%v' does not match the virtualization type of the running system ('%s'). Skipping whole section with all lines till next valid section definition", tagField, secFields, chkVirt)
		ret = false
	}
	return ret
}

// chkOtherTags checks, if the tag is a valid tag (file exists in
// /sys/class/dmi/id) and the contents matches the tag value
// future use possible by simply look for files in an additional location.
func chkOtherTags(file, tagField string, secFields []string) bool {
	ret := true
	tagExpr := fmt.Sprintf(".*%s.*", tagField)
	// check filenames in /sys/class/dmi/id
	chkDmi, err := system.GetDmiID(file)
	if err != nil {
		// file does not exist
		system.WarningLog("skip unknown section tag '%v'.", file)
		ret = false
	} else {
		match, _ := regexp.MatchString(tagExpr, chkDmi)
		if !match {
			// content of file does not match
			system.InfoLog("the string '%s' in section definition '%v' does not match the content of the file '/sys/class/dmi/id/%s' ('%s'). Skipping whole section with all lines till next valid section definition", tagField, secFields, file, chkDmi)
			ret = false
		}
	}
	return ret
}

// chkCPUTags checks, if a cpu related section tag is valid or not
// currently we only support the CPU platform (pmu_name)
func chkCPUTags(tagField string, secFields []string) bool {
	ret := true
	chkCPUpf := system.CPUPlatform()
	if tagField != chkCPUpf {
		// CPU platform does not match
		system.InfoLog("CPU platform '%s' in section definition '%v' does not match the CPU platform of the running system '%s'. Skipping whole section with all lines till next valid section definition", tagField, secFields, chkCPUpf)
		ret = false
	}
	return ret
}

// chkHWTags checks, if the vendor or model section tag is valid or not
// the files to identify the hardware vendor or the hardware model may
// need
func chkHWTags(info, tagField string, secFields []string) bool {
	ret := true
	tagExpr := fmt.Sprintf(".*%s.*", tagField)
	chkHW, err := system.GetHWIdentity(info)
	if err != nil {
		// file to identify the hardware is not available
		system.WarningLog("hardware identification failed. Skipping whole section")
		ret = false
	} else {
		match, _ := regexp.MatchString(tagExpr, chkHW)
		if !match {
			system.InfoLog("hardware %s '%s' in section definition '%v' does not match the hardware %s of the running system ('%s'). Skipping whole section with all lines till next valid section definition", info, tagField, secFields, info, chkHW)
			ret = false
		}
	}
	return ret
}

// chkBlkTags checks if the blkvendor or blkmodel section tag is valid or not
// and returns a list of valid block devices or uses a special device
// pattern to return a list of valid block devices
func chkBlkTags(info, tagField string, secFields, actbdev []string) (bool, []string) {
	ret := false
	info = strings.TrimPrefix(info, "blk")
	tagExpr := fmt.Sprintf(".*%s.*", tagField)
	// vendor or model
	blkInfo := strings.ToUpper(info)
	if info == "pat" {
		// pattern
		blkInfo = info
	}
	bdev := system.GetAvailBlockInfo(blkInfo, tagExpr)
	if len(bdev) == 0 {
		// pattern, vendor or model does not match
		system.InfoLog("%s '%s' in section definition '%v' does not match any available block device %s of the running system. Skipping whole section with all lines till next valid section definition", info, tagField, secFields, info)
	} else {
		// as it is possible to have more than one tag in a
		// section (vendor and module) we need the overlap for
		// a valid result
		if len(actbdev) == 0 {
			// paranoia, as this should never happens because of the 'if'
			// 8 lines above
			// a former call has returned an empty list of valid block
			// devices. So we return an empty list even that this
			// tag has returned a non-empty list.
			bdev = actbdev
		} else {
			// a former call has returned a list of valid block
			// devices and this call also returned a list of valid
			// block devices - get overlap or empty
			newbdev := []string{}
			for _, a := range actbdev {
				for _, b := range bdev {
					if a == b {
						newbdev = append(newbdev, a)
					}
				}
			}
			bdev = newbdev
			if len(newbdev) != 0 {
				ret = true
			}
		}
	}
	return ret, bdev
}
