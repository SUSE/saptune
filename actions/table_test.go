package actions

import (
	"bytes"
	"fmt"
	"github.com/SUSE/saptune/sap/note"
	"testing"
)

func TestSetWidthOfColums(t *testing.T) {
	compare := note.FieldComparison{ReflectFieldName: "SysctlParams", ReflectMapKey: "IO_SCHEDULER_sr0", ActualValueJS: "cfq", ExpectedValueJS: "cfq"}
	w1 := 2
	w2 := 3
	w3 := 4
	w4 := 5
	v1, v2, v3, v4 := setWidthOfColums(compare, w1, w2, w3, w4)
	if v1 != w1 {
		t.Fatal(v1, w1)
	}
	if v2 != 16 {
		t.Fatal(v2, w2)
	}
	if v3 != w3 || v4 != w4 {
		t.Fatal(v3, w3, v4, w4)
	}
	compare = note.FieldComparison{ReflectFieldName: "OverrideParams", ReflectMapKey: "IO_SCHEDULER_sr0", ActualValueJS: "cfq", ExpectedValueJS: "cfq"}
	v1, v2, v3, v4 = setWidthOfColums(compare, w1, w2, w3, w4)
	if v1 != 3 {
		t.Fatal(v1, w1)
	}
	if v2 != w2 || v3 != w3 || v4 != w4 {
		t.Fatal(v2, w2, v3, w3, v4, w4)
	}
	compare = note.FieldComparison{ReflectFieldName: "SysctlParams", ReflectMapKey: "governor", ActualValueJS: "all-none", ExpectedValueJS: "all-performance"}
	v1, v2, v3, v4 = setWidthOfColums(compare, w1, w2, w3, w4)
	if v1 != w1 {
		t.Fatal(v1, w1)
	}
	if v2 != 8 {
		t.Fatal(v2, w2)
	}
	if v3 != 15 {
		t.Fatal(v3, w3)
	}
	if v4 != 8 {
		t.Fatal(v4, w4)
	}
	compare = note.FieldComparison{ReflectFieldName: "SysctlParams", ReflectMapKey: "", ActualValueJS: "all-none", ExpectedValueJS: "all-performance"}
	v1, v2, v3, v4 = setWidthOfColums(compare, w1, w2, w3, w4)
	if v1 != w1 || v2 != w2 || v3 != w3 || v4 != w4 {
		t.Fatal(v1, w1, v2, w2, v3, w3, v4, w4)
	}
}

func TestPrintNoteFields(t *testing.T) {
	var printMatchText1 = `
941735 - Configuration drop in for simple tests
			Version 1 from 09.07.2019  

   SAPNote, Version | Parameter                  | Expected             | Override  | Actual               | Compliant
--------------------+----------------------------+----------------------+-----------+----------------------+-----------
   941735, 1        | IO_SCHEDULER_sdb           |                      |           | all:none             |  -  [1] [5] [7]
   941735, 1        | IO_SCHEDULER_sdc           |                      |           | bfq                  | no  [7]
   941735, 1        | IO_SCHEDULER_sdd           |                      |           | bfq                  | no  [7] [10]
   941735, 1        | IO_SCHEDULER_vda           | noop                 |           | all:none             |  -  [1] [5]
   941735, 1        | ShmFileSystemSizeMB        | 1714                 |           | 488                  | no 
   941735, 1        | force_latency              | 70                   |           | all:none             | no  [4]
   941735, 1        | grub:intel_idle.max_cstate | 1                    |           | NA                   | no  [2] [3] [6]
   941735, 1        | kernel.shmmax              | 18446744073709551615 |           | 18446744073709551615 | yes

  [1] setting is not supported by the system
  [2] setting is not available on the system
  [3] value is only checked, but NOT set
  [4] cpu idle state settings differ
  [5] expected value does not contain a supported scheduler
  [6] grub settings are mostly covered by other settings. See man page saptune-note(5) for details
  [7] parameter value is untouched by default
 [10] parameter is defined twice, see section [sys] 'sys:block.sdd.queue.scheduler' from the other applied notes

`
	var printMatchText2 = `
941735 - Configuration drop in for simple tests
			Version 1 from 09.07.2019  

   Parameter                  | Value set            | Value expected       | Override  | Comment
------------------------------+----------------------+----------------------+-----------+--------------
   IO_SCHEDULER_sdb           | all:none             |                      |           |  [1] [5] [7]
   IO_SCHEDULER_sdc           | bfq                  |                      |           |  [7]
   IO_SCHEDULER_sdd           | bfq                  |                      |           |  [7] [10]
   IO_SCHEDULER_vda           | all:none             | noop                 |           |  [1] [5]
   ShmFileSystemSizeMB        | 488                  | 1714                 |           |   
   force_latency              | all:none             | 70                   |           |  [1] [4]
   grub:intel_idle.max_cstate | NA                   | 1                    |           |  [2] [3] [6]
   kernel.shmmax              | 18446744073709551615 | 18446744073709551615 |           |   

  [1] setting is not supported by the system
  [2] setting is not available on the system
  [3] value is only checked, but NOT set
  [4] cpu idle state settings differ
  [5] expected value does not contain a supported scheduler
  [6] grub settings are mostly covered by other settings. See man page saptune-note(5) for details
  [7] parameter value is untouched by default
 [10] parameter is defined twice, see section [sys] 'sys:block.sdd.queue.scheduler' from the other applied notes

`
	var printMatchText3 = `
   SAPNote, Version | Parameter                  | Expected             | Override  | Actual               | Compliant
--------------------+----------------------------+----------------------+-----------+----------------------+-----------
   941735, 1        | IO_SCHEDULER_sdb           |                      |           | all:none             |  -  [1] [5] [7]
   941735, 1        | IO_SCHEDULER_sdc           |                      |           | bfq                  | no  [7]
   941735, 1        | IO_SCHEDULER_sdd           |                      |           | bfq                  | no  [7] [10]
   941735, 1        | IO_SCHEDULER_vda           | noop                 |           | all:none             |  -  [1] [5]
   941735, 1        | ShmFileSystemSizeMB        | 1714                 |           | 488                  | no 
   941735, 1        | force_latency              | 70                   |           | all:none             | no  [4]
   941735, 1        | grub:intel_idle.max_cstate | 1                    |           | NA                   | no  [2] [3] [6]
   941735, 1        | kernel.shmmax              | 18446744073709551615 |           | 18446744073709551615 | yes

  [1] setting is not supported by the system
  [2] setting is not available on the system
  [3] value is only checked, but NOT set
  [4] cpu idle state settings differ
  [5] expected value does not contain a supported scheduler
  [6] grub settings are mostly covered by other settings. See man page saptune-note(5) for details
  [7] parameter value is untouched by default
 [10] parameter is defined twice, see section [sys] 'sys:block.sdd.queue.scheduler' from the other applied notes

`
	var printMatchText4 = `
   Parameter                  | Value set            | Value expected       | Override  | Comment
------------------------------+----------------------+----------------------+-----------+--------------
   IO_SCHEDULER_sdb           | all:none             |                      |           |  [1] [5] [7]
   IO_SCHEDULER_sdc           | bfq                  |                      |           |  [7]
   IO_SCHEDULER_sdd           | bfq                  |                      |           |  [7] [10]
   IO_SCHEDULER_vda           | all:none             | noop                 |           |  [1] [5]
   ShmFileSystemSizeMB        | 488                  | 1714                 |           |   
   force_latency              | all:none             | 70                   |           |  [1] [4]
   grub:intel_idle.max_cstate | NA                   | 1                    |           |  [2] [3] [6]
   kernel.shmmax              | 18446744073709551615 | 18446744073709551615 |           |   

  [1] setting is not supported by the system
  [2] setting is not available on the system
  [3] value is only checked, but NOT set
  [4] cpu idle state settings differ
  [5] expected value does not contain a supported scheduler
  [6] grub settings are mostly covered by other settings. See man page saptune-note(5) for details
  [7] parameter value is untouched by default
 [10] parameter is defined twice, see section [sys] 'sys:block.sdd.queue.scheduler' from the other applied notes

`
	checkCorrectMessage := func(t *testing.T, got, want string) {
		t.Helper()
		if got != want {
			fmt.Println("==============")
			fmt.Println(got)
			fmt.Println("==============")
			fmt.Println(want)
			fmt.Println("==============")
			t.Errorf("Output differs from expected one")
		}
	}

	//fcomp1 := note.FieldComparison{ReflectFieldName: "ConfFilePath", ReflectMapKey: "", ActualValue: "/usr/share/saptune/notes/941735", ExpectedValue: "/usr/share/saptune/notes/941735", ActualValueJS: "/usr/share/saptune/notes/941735", ExpectedValueJS: "/usr/share/saptune/notes/941735", MatchExpectation: true}
	cfgFile := fmt.Sprintf("%ssimpleNote.conf", ExtraFilesInGOPATH)
	fcomp1 := note.FieldComparison{ReflectFieldName: "ConfFilePath", ReflectMapKey: "", ActualValue: cfgFile, ExpectedValue: cfgFile, ActualValueJS: cfgFile, ExpectedValueJS: cfgFile, MatchExpectation: true}
	fcomp2 := note.FieldComparison{ReflectFieldName: "ID", ReflectMapKey: "", ActualValue: "941735", ExpectedValue: "941735", ActualValueJS: "941735", ExpectedValueJS: "941735", MatchExpectation: true}
	fcomp3 := note.FieldComparison{ReflectFieldName: "DescriptiveName", ReflectMapKey: "", ActualValue: "", ExpectedValue: "", ActualValueJS: "", ExpectedValueJS: "", MatchExpectation: true}
	fcomp4 := note.FieldComparison{ReflectFieldName: "SysctlParams", ReflectMapKey: "ShmFileSystemSizeMB", ActualValue: "488", ExpectedValue: "1714", ActualValueJS: "488", ExpectedValueJS: "1714", MatchExpectation: false}
	fcomp5 := note.FieldComparison{ReflectFieldName: "SysctlParams", ReflectMapKey: "kernel.shmmax", ActualValue: "18446744073709551615", ExpectedValue: "18446744073709551615", ActualValueJS: "18446744073709551615", ExpectedValueJS: "18446744073709551615", MatchExpectation: true}
	fcomp6 := note.FieldComparison{ReflectFieldName: "SysctlParams", ReflectMapKey: "IO_SCHEDULER_vda", ActualValue: "all:none", ExpectedValue: "noop", ActualValueJS: "all:none", ExpectedValueJS: "noop", MatchExpectation: false}
	fcomp7 := note.FieldComparison{ReflectFieldName: "SysctlParams", ReflectMapKey: "grub:intel_idle.max_cstate", ActualValue: "NA", ExpectedValue: "1", ActualValueJS: "NA", ExpectedValueJS: "1", MatchExpectation: false}
	fcomp8 := note.FieldComparison{ReflectFieldName: "SysctlParams", ReflectMapKey: "force_latency", ActualValue: "all:none", ExpectedValue: "70", ActualValueJS: "all:none", ExpectedValueJS: "70", MatchExpectation: false}
	fcomp9 := note.FieldComparison{ReflectFieldName: "Inform", ReflectMapKey: "force_latency", ActualValue: "hasDiffs", ExpectedValue: "hasDiffs", ActualValueJS: "hasDiffs", ExpectedValueJS: "hasDiffs", MatchExpectation: true}
	fcomp10 := note.FieldComparison{ReflectFieldName: "Inform", ReflectMapKey: "IO_SCHEDULER_vda", ActualValue: "NA", ExpectedValue: "NA", ActualValueJS: "NA", ExpectedValueJS: "NA", MatchExpectation: true}
	fcomp11 := note.FieldComparison{ReflectFieldName: "SysctlParams", ReflectMapKey: "IO_SCHEDULER_sdb", ActualValue: "all:none", ExpectedValue: "", ActualValueJS: "all:none", ExpectedValueJS: "", MatchExpectation: false}
	fcomp12 := note.FieldComparison{ReflectFieldName: "Inform", ReflectMapKey: "IO_SCHEDULER_sdb", ActualValue: "NA", ExpectedValue: "", ActualValueJS: "NA", ExpectedValueJS: "", MatchExpectation: false}
	fcomp13 := note.FieldComparison{ReflectFieldName: "SysctlParams", ReflectMapKey: "IO_SCHEDULER_sdc", ActualValue: "bfq", ExpectedValue: "", ActualValueJS: "bfq", ExpectedValueJS: "", MatchExpectation: false}
	fcomp14 := note.FieldComparison{ReflectFieldName: "Inform", ReflectMapKey: "IO_SCHEDULER_sdc", ActualValue: "", ExpectedValue: "bfq", ActualValueJS: "", ExpectedValueJS: "bfq", MatchExpectation: false}
	fcomp15 := note.FieldComparison{ReflectFieldName: "SysctlParams", ReflectMapKey: "IO_SCHEDULER_sdd", ActualValue: "bfq", ExpectedValue: "", ActualValueJS: "bfq", ExpectedValueJS: "", MatchExpectation: false}
	fcomp16 := note.FieldComparison{ReflectFieldName: "Inform", ReflectMapKey: "IO_SCHEDULER_sdd", ActualValue: "", ExpectedValue: "[sys] 'sys:block.sdd.queue.scheduler' from the other applied notes", ActualValueJS: "", ExpectedValueJS: "[sys] 'sys:block.sdd.queue.scheduler' from the other applied notes", MatchExpectation: false}

	map941735 := map[string]note.FieldComparison{"ConfFilePath": fcomp1, "ID": fcomp2, "DescriptiveName": fcomp3, "SysctlParams[ShmFileSystemSizeMB]": fcomp4, "SysctlParams[kernel.shmmax]": fcomp5, "SysctlParams[IO_SCHEDULER_vda]": fcomp6, "SysctlParams[grub:intel_idle.max_cstate]": fcomp7, "SysctlParams[force_latency]": fcomp8, "Inform[force_latency]": fcomp9, "Inform[IO_SCHEDULER_vda]": fcomp10, "SysctlParams[IO_SCHEDULER_sdb]": fcomp11, "Inform[IO_SCHEDULER_sdb]": fcomp12, "SysctlParams[IO_SCHEDULER_sdc]": fcomp13, "Inform[IO_SCHEDULER_sdc]": fcomp14, "SysctlParams[IO_SCHEDULER_sdd]": fcomp15, "Inform[IO_SCHEDULER_sdd]": fcomp16}
	noteComp := map[string]map[string]note.FieldComparison{"941735": map941735}

	t.Run("verify with header", func(t *testing.T) {
		buffer := bytes.Buffer{}
		PrintNoteFields(&buffer, "HEAD", noteComp, true)
		txt := buffer.String()
		checkCorrectMessage(t, txt, printMatchText1)
	})
	t.Run("simulate with header", func(t *testing.T) {
		buffer := bytes.Buffer{}
		PrintNoteFields(&buffer, "HEAD", noteComp, false)
		txt := buffer.String()
		checkCorrectMessage(t, txt, printMatchText2)
	})
	t.Run("verify without header", func(t *testing.T) {
		buffer := bytes.Buffer{}
		PrintNoteFields(&buffer, "NONE", noteComp, true)
		txt := buffer.String()
		checkCorrectMessage(t, txt, printMatchText3)
	})
	t.Run("simulate without header", func(t *testing.T) {
		buffer := bytes.Buffer{}
		PrintNoteFields(&buffer, "NONE", noteComp, false)
		txt := buffer.String()
		checkCorrectMessage(t, txt, printMatchText4)
	})
}
