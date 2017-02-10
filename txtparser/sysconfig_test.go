package txtparser

import (
	"fmt"
	"reflect"
	"testing"
)

var sysconfSampleText = `## Path:        Productivity/Other
## Description: Limits for system tuning profile "sap-netweaver".
## ServiceRestart: tuned

## Type:        integer
## Default:     8388608
#
# The lower tuning limit of the size of tmpfs mounted on /dev/shm in KiloBytes.
# It should not be smaller than 8388608 (8GB).
#
TMPFS_SIZE_MIN=8388608

## Type:        regexp(^@(sapsys|sdba|dba)[[:space:]]+(-|hard|soft)[[:space:]]+(nofile)[[:space:]]+[[:digit:]]+)
## Default:     ""
#
# Maximum number of open files for SAP application groups sapsys, sdba, and dba.
# Consult with manual page limits.conf(5) for the correct syntax.
#
LIMIT_1="@sapsys soft nofile 65536"
LIMIT_2="@sapsys hard nofile 65536"
BOOL_TEST_YES="yes"
BOOL_TEST_TRUE="true"
BOOL_TEST_EMPTY=""
BOOL_TEST_NO="no"
BOOL_TEST_FALSE="false"
STRARY_TEST=" foo bar abc "
INTARY_TEST=" 12 34 abc 56 "
`

var sysconfigMatchText = `## Path:        Productivity/Other
## Description: Limits for system tuning profile "sap-netweaver".
## ServiceRestart: tuned

## Type:        integer
## Default:     8388608
#
# The lower tuning limit of the size of tmpfs mounted on /dev/shm in KiloBytes.
# It should not be smaller than 8388608 (8GB).
#
TMPFS_SIZE_MIN="8388608"

## Type:        regexp(^@(sapsys|sdba|dba)[[:space:]]+(-|hard|soft)[[:space:]]+(nofile)[[:space:]]+[[:digit:]]+)
## Default:     ""
#
# Maximum number of open files for SAP application groups sapsys, sdba, and dba.
# Consult with manual page limits.conf(5) for the correct syntax.
#
LIMIT_1="new value"
LIMIT_2="@sapsys hard nofile 65536"
BOOL_TEST_YES="yes"
BOOL_TEST_TRUE="true"
BOOL_TEST_EMPTY=""
BOOL_TEST_NO="no"
BOOL_TEST_FALSE="false"
STRARY_TEST="foo bar"
INTARY_TEST="12 34"
newkey="orange"
STRARY_TEST2="foo bar"
`

func TestSysconfig(t *testing.T) {
	// Parse the sample text
	conf, err := ParseSysconfig(sysconfSampleText)
	if err != nil {
		t.Fatal(err)
	}
	// Read keys
	if val := conf.GetString("LIMIT_1", ""); val != "@sapsys soft nofile 65536" {
		t.Fatal(val)
	}
	if val := conf.GetString("TMPFS_SIZE_MIN", ""); val != "8388608" {
		t.Fatal(val)
	}
	if val := conf.GetInt("TMPFS_SIZE_MIN", 0); val != 8388608 {
		t.Fatal(val)
	}
	if val := conf.GetUint64("TMPFS_SIZE_MIN", 0); val != 8388608 {
		t.Fatal(val)
	}
	if val := conf.GetString("KEY_DOES_NOT_EXIST", "DEFAULT"); val != "DEFAULT" {
		t.Fatal(val)
	}
	if val := conf.GetInt("KEY_DOES_NOT_EXIST", 12); val != 12 {
		t.Fatal(val)
	}
	// Read array keys
	if val := conf.GetStringArray("STRARY_TEST", nil); !reflect.DeepEqual(val, []string{"foo", "bar", "abc"}) {
		t.Fatal(val)
	}
	if val := conf.GetIntArray("INTARY_TEST", nil); !reflect.DeepEqual(val, []int{12, 34, 56}) {
		t.Fatal(val)
	}
	// Read boolean keys
	if val := conf.GetBool("BOOL_TEST_YES", false); !val {
		t.Fatal(val)
	}
	if val := conf.GetBool("BOOL_TEST_TRUE", false); !val {
		t.Fatal(val)
	}
	if val := conf.GetBool("BOOL_TEST_EMPTY", true); !val {
		t.Fatal(val)
	}
	if val := conf.GetBool("BOOL_TEST_EMPTY", false); val {
		t.Fatal(val)
	}
	if val := conf.GetBool("BOOL_TEST_NO", true); val {
		t.Fatal(val)
	}
	if val := conf.GetBool("BOOL_TEST_FALSE", true); val {
		t.Fatal(val)
	}
	// Write keys
	conf.Set("LIMIT_1", "new value")
	conf.Set("newkey", "orange")
	if val := conf.GetString("LIMIT_1", ""); val != "new value" {
		t.Fatal(val)
	}
	if val := conf.GetInt("newkey", 12); val != 12 {
		t.Fatal(val)
	}
	if val := conf.GetString("newkey", ""); val != "orange" {
		t.Fatal(val)
	}
	// Write array keys
	conf.SetIntArray("INTARY_TEST", []int{12, 34})
	conf.SetStrArray("STRARY_TEST", []string{"foo", "bar"})
	conf.SetStrArray("STRARY_TEST2", []string{"foo", "bar"})
	// The converted back text should carry "new value" for LIMIT_1 and newkey.
	if txt := conf.ToText(); txt != sysconfigMatchText {
		fmt.Println("==================")
		fmt.Println(txt)
		fmt.Println("==================")
		t.Fatal("failed to convert back into text")
	}
}
