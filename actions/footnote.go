package actions

import (
	"github.com/SUSE/saptune/sap/note"
	"github.com/SUSE/saptune/system"
	"regexp"
	"runtime"
	"strconv"
	"strings"
)

//define footnote texts
const (
	footnote1X86 = " [1] setting is not supported by the system"
	footnote1IBM = " [1] setting is not relevant for the system"
	footnote1AZR = " [1] setting is not available on Azure instances (see SAP Note 2993054)."
	footnote1AWS = " [1] setting is not available on AWS instances (see SAP Note 1656250)."
	footnote2    = " [2] setting is not available on the system"
	footnote3    = " [3] value is only checked, but NOT set"
	footnote4    = " [4] cpu idle state settings differ"
	footnote5    = " [5] expected value does not contain a supported scheduler"
	footnote6    = " [6] grub settings are mostly covered by other settings. See man page saptune-note(5) for details"
	footnote7    = " [7] parameter value is untouched by default"
	footnote8    = " [8] cannot set Perf Bias because SecureBoot is enabled"
	footnote9    = " [9] expected value limited to 'max_hw_sectors_kb'"
	footnote10   = "[10] parameter is defined twice, see section SECT"
	footnote11   = "[11] parameter is additional defined in SYSCTLLIST"
	footnote12   = "[12] option FSOPT"
	footnote13   = "[13] The SAP recommendation for nr_request does not work in the context of multiqueue block framework (scheduler=none).\n      Maximal supported value by the hardware is MAXVAL"
	footnote14   = "[14] the parameter value exceeds the maximum possible number of open files. Check and increase fs.nr_open if really needed."
	footnote15   = "[15] the parameter is only used to calculate the size of tmpfs (/dev/shm)"
)

// set 'unsupported' footnote regarding the architecture
var footnote1 = footnote1X86

// prepFN checks, if we need to prepare the footnote for a parameter
// if the command line flage '--show-non-compliant' is used only non compliant
// parameter rows will be printed and that has to be reflected to the footnotes
// too.
func prepFN(comparison note.FieldComparison, compliant, inform string) bool {
	prep := true
	if system.IsFlagSet("show-non-compliant") {
		// skip preparation of footnotes, if compliant state is not 'no'
		if strings.Contains(compliant, "yes") || strings.Contains(compliant, "-") {
			prep = false
		}
		// and now define the exceptions, which need a special handling
		if comparison.ReflectMapKey == "force_latency" && inform == "hasDiffs" {
			// compliant will be set to 'no' during footnote preparation
			prep = true
		}
		if comparison.ReflectMapKey == "VSZ_TMPFS_PERCENT" {
			// compliant will be set to '-' during footnote preparation
			prep = true
		}
		if strings.Contains(comparison.ReflectMapKey, "xfsopt_") {
			prep = true
		}
	}
	return prep
}

// prepareFootnote prepares the content of the last column and the
// corresponding footnotes
func prepareFootnote(comparison note.FieldComparison, compliant, comment, inform string, footnote []string) (string, string, []string) {
	if !prepFN(comparison, compliant, inform) {
		return compliant, comment, footnote
	}
	// set 'unsupported' footnote regarding the architecture
	if runtime.GOARCH == "ppc64le" {
		footnote1 = footnote1IBM
	}
	if system.GetCSP() == "azure" {
		footnote1 = footnote1AZR
	}
	if system.GetCSP() == "aws" {
		footnote1 = footnote1AWS
	}
	// set footnote for unsupported or not available parameter [1],[2]
	compliant, comment, footnote = setUsNa(comparison.ActualValue.(string), compliant, comment, footnote)
	// set footnote for rpm or grub parameter [3],[6]
	compliant, comment, footnote = setRpmGrub(comparison.ReflectMapKey, compliant, comment, footnote)
	// set footnote for diffs in force_latency parameter [4]
	compliant, comment, footnote = setFLdiffs(comparison.ReflectMapKey, compliant, comment, inform, footnote)
	// set footnote for unsupported scheduler [5]
	compliant, comment, footnote = setUnSched(comparison.ReflectMapKey, compliant, comment, inform, footnote)
	// set footnote for untouched parameter [7]
	compliant, comment, footnote = setUntouched(comparison.ExpectedValue.(string), compliant, comment, footnote)
	// set footnote for secure boot [8]
	compliant, comment, footnote = setSecBoot(comparison.ReflectMapKey, compliant, comment, footnote)
	// set footnote for limited parameter value [9]
	compliant, comment, footnote = setLimited(comparison.ReflectMapKey, compliant, comment, inform, footnote)
	// set footnote for double defined parameters [10]
	compliant, comment, footnote = setDouble(comparison.ReflectMapKey, compliant, comment, inform, footnote)
	// set footnote for system wide (global) defines sysctl parameter [11]
	compliant, comment, footnote = setSysctlGlobal(compliant, comment, inform, footnote)
	// set footnote for filesystem options [12]
	compliant, comment, footnote = setFSOptions(comparison, compliant, comment, inform, footnote)
	// set footnote for unsupported nr_request value [13]
	compliant, comment, footnote = setUnNRR(comparison.ReflectMapKey, compliant, comment, inform, footnote)
	// set footnote for unsupported nofile limit value [14]
	compliant, comment, footnote = setNofile(comparison.ReflectMapKey, compliant, comment, inform, footnote)
	// set footnote for VSZ_TMPFS_PERCENT parameter from mem section
	compliant, comment, footnote = setMem(comparison.ReflectMapKey, compliant, comment, footnote)
	return compliant, comment, footnote
}

// setUsNa sets footnote for unsupported or not available parameter
func setUsNa(actVal, compliant, comment string, footnote []string) (string, string, []string) {
	switch actVal {
	case "all:none":
		compliant = compliant + " [1]"
		comment = comment + " [1]"
		footnote[0] = footnote1
	case "NA", "":
		compliant = compliant + " [2]"
		comment = comment + " [2]"
		footnote[1] = footnote2
	}
	return compliant, comment, footnote
}

// setRpmGrub sets footnote for rpm or grub parameter
func setRpmGrub(mapKey, compliant, comment string, footnote []string) (string, string, []string) {
	if strings.Contains(mapKey, "rpm") || strings.Contains(mapKey, "grub") {
		compliant = compliant + " [3]"
		comment = comment + " [3]"
		footnote[2] = footnote3
	}
	if strings.Contains(mapKey, "grub") {
		compliant = compliant + " [6]"
		comment = comment + " [6]"
		footnote[5] = footnote6
	}
	return compliant, comment, footnote
}

// setUntouched sets footnote for untouched parameter
func setUntouched(expVal, compliant, comment string, footnote []string) (string, string, []string) {
	if expVal == "" {
		compliant = compliant + " [7]"
		comment = comment + " [7]"
		footnote[6] = footnote7
	}
	return compliant, comment, footnote
}

// setSecBoot sets footnote for secure boot affected parameter
func setSecBoot(mapKey, compliant, comment string, footnote []string) (string, string, []string) {
	if mapKey == "energy_perf_bias" && system.SecureBootEnabled() {
		compliant = compliant + " [8]"
		comment = comment + " [8]"
		footnote[7] = footnote8
	}
	return compliant, comment, footnote
}

// setFLdiffs sets footnote for diffs in force_latency parameter
func setFLdiffs(mapKey, compliant, comment, info string, footnote []string) (string, string, []string) {
	if mapKey == "force_latency" && info == "hasDiffs" {
		compliant = "no  [4]"
		comment = comment + " [4]"
		footnote[3] = footnote4
	}
	return compliant, comment, footnote
}

// setUnSched sets footnote for unsupported scheduler
func setUnSched(mapKey, compliant, comment, info string, footnote []string) (string, string, []string) {
	if system.IsSched.MatchString(mapKey) && strings.Contains(info, "NA") {
		compliant = compliant + " [5]"
		comment = comment + " [5]"
		footnote[4] = footnote5
	}
	return compliant, comment, footnote
}

// setUnNRR sets footnote for unsupported nr_request values
func setUnNRR(mapKey, compliant, comment, info string, footnote []string) (string, string, []string) {
	if system.IsNrreq.MatchString(mapKey) && strings.Contains(info, "wrongVal") {
		compliant = compliant + " [13]"
		comment = comment + " [13]"
		maxVal, _, _ := system.GetNrTags(mapKey)
		footnote[12] = writeFN(footnote[12], footnote13, strconv.Itoa(maxVal), "MAXVAL")
	}
	return compliant, comment, footnote
}

// setLimit sets footnote for limited parameter value
func setLimited(mapKey, compliant, comment, info string, footnote []string) (string, string, []string) {
	if system.IsMsect.MatchString(mapKey) && strings.Contains(info, "limited") {
		compliant = compliant + " [9]"
		comment = comment + " [9]"
		footnote[8] = footnote9
	}
	return compliant, comment, footnote
}

// setDouble sets footnote for double defined sys parameters
func setDouble(mapKey, compliant, comment, info string, footnote []string) (string, string, []string) {
	if (system.IsSched.MatchString(mapKey) || system.IsNrreq.MatchString(mapKey) || system.IsRahead.MatchString(mapKey) || system.IsMsect.MatchString(mapKey)) && info != "" {
		// check for double defined parameters
		sect := regexp.MustCompile(`.*\[\w+\].*`)
		inf := strings.Split(info, "§")
		if len(inf) > 1 {
			if inf[0] != "limited" && inf[0] != "NA" {
				info = inf[0]
			} else {
				info = inf[1]
			}
		}
		if info != "limited" && info != "NA" && sect.MatchString(info) {
			compliant = compliant + " [10]"
			comment = comment + " [10]"
			footnote[9] = writeFN(footnote[9], footnote10, info, "SECT")
		}
	}
	if (strings.Contains(mapKey, "THP") || strings.Contains(mapKey, "KSM")) && info != "" {
		compliant = compliant + " [10]"
		comment = comment + " [10]"
		footnote[9] = writeFN(footnote[9], footnote10, info, "SECT")
	}
	if strings.Contains(mapKey, "sys:") && info != "" {
		compliant = compliant + " [10]"
		comment = comment + " [10]"
		footnote[9] = writeFN(footnote[9], footnote10, info, "SECT")
	}
	return compliant, comment, footnote
}

// setSysctlGlobal sets footnote for global defined sysctl parameters
func setSysctlGlobal(compliant, comment, info string, footnote []string) (string, string, []string) {
	// check if the sysctl parameter is additional set in a sysctl system
	// configuration file
	if strings.HasPrefix(info, "sysctl config file ") {
		// sysctl info
		compliant = compliant + " [11]"
		comment = comment + " [11]"
		footnote[10] = writeFN(footnote[10], footnote11, info, "SYSCTLLIST")
	}
	return compliant, comment, footnote
}

// setFSOptions sets footnote for not matching filesystem options
func setFSOptions(comparison note.FieldComparison, compliant, comment, info string, footnote []string) (string, string, []string) {
	// check if there are mount points with wrong FS option settings
	if strings.Contains(comparison.ReflectMapKey, "xfsopt_") {
		if !system.IsFlagSet("show-non-compliant") && info != "" {
			// fs option info
			compliant = compliant + " [12]"
			comment = comment + " [12]"
			footnote[11] = writeFN(footnote[11], footnote12, info, "FSOPT")
		}
		if comparison.ActualValue.(string) == "NA" {
			compliant = strings.Replace(compliant, "no ", " - ", 1)
		}
	}
	return compliant, comment, footnote
}

// setNofile sets footnote for unsupported nofile limit value
func setNofile(mapKey, compliant, comment, info string, footnote []string) (string, string, []string) {
	if strings.Contains(mapKey, "LIMIT_") && info == "limit_exceeded" {
		compliant = compliant + " [14]"
		comment = comment + " [14]"
		footnote[13] = footnote14
	}
	return compliant, comment, footnote
}

// setMem sets footnote for VSZ_TMPFS_PERCENT parameter from mem section
func setMem(mapKey, compliant, comment string, footnote []string) (string, string, []string) {
	if mapKey == "VSZ_TMPFS_PERCENT" {
		if system.IsFlagSet("show-non-compliant") {
			compliant = " - "
		} else {
			compliant = " -  [15]"
			comment = comment + " [15]"
			footnote[14] = footnote15
		}
	}
	return compliant, comment, footnote
}

// writeFN customizes the text for footnotes by replacing strings/placeholder
func writeFN(footnote, fntxt, info, pat string) string {
	if footnote == "" {
		footnote = strings.Replace(fntxt, pat, info, 1)
	} else {
		footnote = footnote + "\n " + strings.Replace(fntxt, pat, info, 1)
	}
	return footnote
}
