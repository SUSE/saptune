package actions

import (
	"fmt"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"github.com/SUSE/saptune/sap/note"
	"github.com/SUSE/saptune/system"
)

type SetterFootnoteDescriptor struct {
	ActualValue   string
	ExpectedValue string
	MapKey        string
	Info          string
	Function      func(SetterFootnoteDescriptor, string, string, []string) (string, string, []string)
}

var footnotes = map[string]string{
	"1X86": " [1] setting is not supported by the system",
	"1IBM": " [1] setting is not relevant for the system",
	"1AZR": " [1] setting is not available on Azure instances (see SAP Note 2993054).",
	"1AWS": " [1] setting is not available on AWS instances (see SAP Note 1656250).",
	// set 'unsupported' footnote regarding the architecture
	"1":  " [1] setting is not supported by the system",
	"2":  " [2] setting is not available on the system",
	"3":  " [3] value is only checked, but NOT set",
	"4":  " [4] cpu idle state settings differ",
	"5":  " [5] expected value does not contain a supported scheduler",
	"6":  " [6] grub settings are mostly covered by other settings. See man page saptune-note(5) for details",
	"7":  " [7] parameter value is untouched by default",
	"8":  " [8] cannot set Perf Bias because SecureBoot is enabled",
	"9":  " [9] expected value limited to 'max_hw_sectors_kb'",
	"10": "[10] parameter is defined twice, see section SECT",
	"11": "[11] parameter is additional defined in SYSCTLLIST",
	"12": "[12] option FSOPT",
	"13": "[13] The SAP recommendation for nr_request does not work in the context of multiqueue block framework (scheduler=none).\n      Maximal supported value by the hardware is MAXVAL",
	"14": "[14] the parameter value exceeds the maximum possible number of open files. Check and increase fs.nr_open if really needed.",
	"15": "[15] the parameter is only used to calculate the size of tmpfs (/dev/shm)",
	"16": "[16] parameter not available on the system, setting not possible",
}

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

// initialization of all the setters for the footnote
func getAllFootnoteDescriptorSetters(comparison note.FieldComparison, inform string) []SetterFootnoteDescriptor {
	return []SetterFootnoteDescriptor{
		// set footnote for unsupported or not available parameter [1],[2]
		{comparison.ActualValue.(string), "", "", "", setUsNa},
		// set footnote for rpm or grub parameter [3],[6]
		{"", "", comparison.ReflectMapKey, "", setRpmGrub},
		// set footnote for diffs in force_latency parameter [4]
		{"", "", comparison.ReflectMapKey, inform, setFLdiffs},
		// set footnote for unsupported scheduler [5]
		{"", "", comparison.ReflectMapKey, inform, setUnSched},
		// set footnote for untouched parameter [7]
		{"", comparison.ExpectedValue.(string), "", "", setUntouched},
		// set footnote for secure boot [8]
		{"", "", comparison.ReflectMapKey, "", setSecBoot},
		// set footnote for limited parameter value [9]
		{"", "", comparison.ReflectMapKey, inform, setLimited},
		// set footnote for double defined parameters [10]
		{"", "", comparison.ReflectMapKey, inform, setDouble},
		// set footnote for system wide (global) defines sysctl parameter [11]
		{"", "", "", inform, setSysctlGlobal},
		// set footnote for filesystem options [12]
		{comparison.ActualValue.(string), "", comparison.ReflectMapKey, inform, setFSOptions},
		// set footnote for unsupported nr_request value [13]
		{"", "", comparison.ReflectMapKey, inform, setUnNRR},
		// set footnote for unsupported nofile limit value [14]
		{"", "", comparison.ReflectMapKey, inform, setNofile},
		// set footnote for VSZ_TMPFS_PERCENT parameter from mem section
		{"", "", comparison.ReflectMapKey, "", setMem},
	}
}

// prepareFootnote prepares the content of the last column and the
// corresponding footnotes
func prepareFootnote(comparison note.FieldComparison, compliant, comment, inform string, footnote []string) (string, string, []string) {
	if !prepFN(comparison, compliant, inform) {
		return compliant, comment, footnote
	}
	// set 'unsupported' footnote regarding the architecture
	if runtime.GOARCH == "ppc64le" {
		footnotes["1"] = footnotes["1IBM"]
	}
	if system.GetCSP() == "azure" {
		footnotes["1"] = footnotes["1AZR"]
	}
	if system.GetCSP() == "aws" {
		footnotes["1"] = footnotes["1AWS"]
	}

	for _, setter := range getAllFootnoteDescriptorSetters(comparison, inform) {
		compliant, comment, footnote = setter.Function(setter, compliant, comment, footnote)
	}

	return compliant, comment, footnote
}

func setCompliantCommentFootnote(footnoteNumber int8, compliant string, comment string, footnote []string, updateFootnote bool) (string, string, []string) {
	number := fmt.Sprint(footnoteNumber)
	compliant = compliant + " [" + number + "]"
	comment = comment + " [" + number + "]"
	if updateFootnote {
		footnote[footnoteNumber-1] = footnotes[number]
	}
	return compliant, comment, footnote
}

// setUsNa sets footnote for unsupported or not available parameter
func setUsNa(comparison SetterFootnoteDescriptor, compliant, comment string, footnote []string) (string, string, []string) {
	switch comparison.ActualValue {
	case "all:none":
		compliant, comment, footnote = setCompliantCommentFootnote(1, compliant, comment, footnote, true)
	case "NA":
		compliant, comment, footnote = setCompliantCommentFootnote(2, compliant, comment, footnote, true)
	case "PNA":
		compliant, comment, footnote = setCompliantCommentFootnote(16, compliant, comment, footnote, !system.IsFlagSet("show-non-compliant"))
		compliant = strings.Replace(compliant, "no ", " - ", 1)
	}
	return compliant, comment, footnote
}

// setRpmGrub sets footnote for rpm or grub parameter
func setRpmGrub(comparison SetterFootnoteDescriptor, compliant, comment string, footnote []string) (string, string, []string) {
	mapKey := comparison.MapKey
	if strings.Contains(mapKey, "rpm") || strings.Contains(mapKey, "grub") {
		compliant, comment, footnote = setCompliantCommentFootnote(3, compliant, comment, footnote, true)
	}
	if strings.Contains(mapKey, "grub") {
		compliant, comment, footnote = setCompliantCommentFootnote(6, compliant, comment, footnote, true)
	}
	return compliant, comment, footnote
}

// setUntouched sets footnote for untouched parameter
func setUntouched(comparison SetterFootnoteDescriptor, compliant, comment string, footnote []string) (string, string, []string) {
	if comparison.ExpectedValue == "" {
		compliant, comment, footnote = setCompliantCommentFootnote(7, compliant, comment, footnote, true)
	}
	return compliant, comment, footnote
}

// setSecBoot sets footnote for secure boot affected parameter
func setSecBoot(comparison SetterFootnoteDescriptor, compliant, comment string, footnote []string) (string, string, []string) {
	if comparison.MapKey == "energy_perf_bias" && system.SecureBootEnabled() {
		compliant, comment, footnote = setCompliantCommentFootnote(8, compliant, comment, footnote, true)
	}
	return compliant, comment, footnote
}

// setFLdiffs sets footnote for diffs in force_latency parameter
func setFLdiffs(comparison SetterFootnoteDescriptor, compliant string, comment string, footnote []string) (string, string, []string) {
	if comparison.MapKey == "force_latency" && comparison.Info == "hasDiffs" {
		compliant, comment, footnote = setCompliantCommentFootnote(4, compliant, comment, footnote, true)
		compliant = strings.Replace(compliant, " - ", "no ", 1)
	}
	return compliant, comment, footnote
}

// setUnSched sets footnote for unsupported scheduler
func setUnSched(comparison SetterFootnoteDescriptor, compliant string, comment string, footnote []string) (string, string, []string) {
	if system.IsSched.MatchString(comparison.MapKey) && strings.Contains(comparison.Info, "NA") {
		compliant, comment, footnote = setCompliantCommentFootnote(5, compliant, comment, footnote, true)
	}
	return compliant, comment, footnote
}

// setUnNRR sets footnote for unsupported nr_request values
func setUnNRR(comparison SetterFootnoteDescriptor, compliant string, comment string, footnote []string) (string, string, []string) {
	mapKey := comparison.MapKey
	if system.IsNrreq.MatchString(mapKey) && strings.Contains(comparison.Info, "wrongVal") {
		compliant, comment, footnote = setCompliantCommentFootnote(13, compliant, comment, footnote, false)
		maxVal, _, _ := system.GetNrTags(mapKey)
		footnote[12] = writeFN(footnote[12], footnotes["13"], strconv.Itoa(maxVal), "MAXVAL")
	}
	return compliant, comment, footnote
}

// setLimit sets footnote for limited parameter value
func setLimited(comparison SetterFootnoteDescriptor, compliant string, comment string, footnote []string) (string, string, []string) {
	if system.IsMsect.MatchString(comparison.MapKey) && strings.Contains(comparison.Info, "limited") {
		compliant, comment, footnote = setCompliantCommentFootnote(9, compliant, comment, footnote, true)
	}
	return compliant, comment, footnote
}

// setDouble sets footnote for double defined sys parameters
func setDouble(comparison SetterFootnoteDescriptor, compliant string, comment string, footnote []string) (string, string, []string) {
	mapKey := comparison.MapKey
	info := comparison.Info
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
			compliant, comment, footnote = setCompliantCommentFootnote(10, compliant, comment, footnote, false)
			footnote[9] = writeFN(footnote[9], footnotes["10"], info, "SECT")
		}
	}
	if (strings.Contains(mapKey, "THP") || strings.Contains(mapKey, "KSM")) && info != "" {
		compliant, comment, footnote = setCompliantCommentFootnote(10, compliant, comment, footnote, false)
		footnote[9] = writeFN(footnote[9], footnotes["10"], info, "SECT")
	}
	if strings.Contains(mapKey, "sys:") && info != "" {
		compliant, comment, footnote = setCompliantCommentFootnote(10, compliant, comment, footnote, false)
		footnote[9] = writeFN(footnote[9], footnotes["10"], info, "SECT")
	}
	return compliant, comment, footnote
}

// setSysctlGlobal sets footnote for global defined sysctl parameters
func setSysctlGlobal(comparison SetterFootnoteDescriptor, compliant string, comment string, footnote []string) (string, string, []string) {
	// check if the sysctl parameter is additional set in a sysctl system
	// configuration file
	info := comparison.Info
	if strings.HasPrefix(info, "sysctl config file ") {
		// sysctl info
		compliant, comment, footnote = setCompliantCommentFootnote(11, compliant, comment, footnote, false)
		footnote[10] = writeFN(footnote[10], footnotes["11"], info, "SYSCTLLIST")
	}
	return compliant, comment, footnote
}

// setFSOptions sets footnote for not matching filesystem options
func setFSOptions(comparison SetterFootnoteDescriptor, compliant string, comment string, footnote []string) (string, string, []string) {
	info := comparison.Info
	// check if there are mount points with wrong FS option settings
	if strings.Contains(comparison.MapKey, "xfsopt_") {
		if !system.IsFlagSet("show-non-compliant") && info != "" {
			// fs option info
			compliant, comment, footnote = setCompliantCommentFootnote(12, compliant, comment, footnote, false)
			footnote[11] = writeFN(footnote[11], footnotes["12"], info, "FSOPT")
		}
		if comparison.ActualValue == "NA" {
			compliant = strings.Replace(compliant, "no ", " - ", 1)
		}
	}
	return compliant, comment, footnote
}

// setNofile sets footnote for unsupported nofile limit value
func setNofile(comparison SetterFootnoteDescriptor, compliant string, comment string, footnote []string) (string, string, []string) {
	if strings.Contains(comparison.MapKey, "LIMIT_") && comparison.Info == "limit_exceeded" {
		compliant, comment, footnote = setCompliantCommentFootnote(14, compliant, comment, footnote, true)
	}
	return compliant, comment, footnote
}

// setMem sets footnote for VSZ_TMPFS_PERCENT parameter from mem section
func setMem(comparison SetterFootnoteDescriptor, compliant, comment string, footnote []string) (string, string, []string) {
	if comparison.MapKey == "VSZ_TMPFS_PERCENT" {
		if system.IsFlagSet("show-non-compliant") {
			compliant = " - "
		} else {
			_, comment, footnote = setCompliantCommentFootnote(15, compliant, comment, footnote, true)
			compliant = " -  [15]"
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
