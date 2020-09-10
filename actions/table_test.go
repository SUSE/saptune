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
941735 -  

   SAPNote, Version | Parameter           | Expected             | Override  | Actual               | Compliant
--------------------+---------------------+----------------------+-----------+----------------------+-----------
   941735,          | ShmFileSystemSizeMB | 1714                 |           | 488                  | no 
   941735,          | kernel.shmmax       | 18446744073709551615 |           | 18446744073709551615 | yes


`
	var printMatchText2 = `
941735 -  

   Parameter           | Value set            | Value expected       | Override  | Comment
-----------------------+----------------------+----------------------+-----------+--------------
   ShmFileSystemSizeMB | 488                  | 1714                 |           |   
   kernel.shmmax       | 18446744073709551615 | 18446744073709551615 |           |   


`
	var printMatchText3 = `   SAPNote, Version | Parameter           | Expected             | Override  | Actual               | Compliant
--------------------+---------------------+----------------------+-----------+----------------------+-----------
   941735,          | ShmFileSystemSizeMB | 1714                 |           | 488                  | no 
   941735,          | kernel.shmmax       | 18446744073709551615 |           | 18446744073709551615 | yes


`
	var printMatchText4 = `   Parameter           | Value set            | Value expected       | Override  | Comment
-----------------------+----------------------+----------------------+-----------+--------------
   ShmFileSystemSizeMB | 488                  | 1714                 |           |   
   kernel.shmmax       | 18446744073709551615 | 18446744073709551615 |           |   


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

	fcomp1 := note.FieldComparison{ReflectFieldName: "ConfFilePath", ReflectMapKey: "", ActualValue: "/usr/share/saptune/notes/941735", ExpectedValue: "/usr/share/saptune/notes/941735", ActualValueJS: "/usr/share/saptune/notes/941735", ExpectedValueJS: "/usr/share/saptune/notes/941735", MatchExpectation: true}
	fcomp2 := note.FieldComparison{ReflectFieldName: "ID", ReflectMapKey: "", ActualValue: "941735", ExpectedValue: "941735", ActualValueJS: "941735", ExpectedValueJS: "941735", MatchExpectation: true}
	fcomp3 := note.FieldComparison{ReflectFieldName: "DescriptiveName", ReflectMapKey: "", ActualValue: "", ExpectedValue: "", ActualValueJS: "", ExpectedValueJS: "", MatchExpectation: true}
	fcomp4 := note.FieldComparison{ReflectFieldName: "SysctlParams", ReflectMapKey: "ShmFileSystemSizeMB", ActualValue: "488", ExpectedValue: "1714", ActualValueJS: "488", ExpectedValueJS: "1714", MatchExpectation: false}
	fcomp5 := note.FieldComparison{ReflectFieldName: "SysctlParams", ReflectMapKey: "kernel.shmmax", ActualValue: "18446744073709551615", ExpectedValue: "18446744073709551615", ActualValueJS: "18446744073709551615", ExpectedValueJS: "18446744073709551615", MatchExpectation: true}
	map941735 := map[string]note.FieldComparison{"ConfFilePath": fcomp1, "ID": fcomp2, "DescriptiveName": fcomp3, "SysctlParams[ShmFileSystemSizeMB]": fcomp4, "SysctlParams[kernel.shmmax]": fcomp5}
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
