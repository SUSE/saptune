package system

import (
	"fmt"
	"testing"
)

var limitsSampleText = `# yadi yadi yada
# /etc/security/limits.conf
*               hard    nproc           8000
*               soft    nproc           4000
root            hard    nproc           16000
root            soft    nproc           8000

some random soft nofile 12345

*               hard    nofile          32000
*               soft    nofile          16000
root            hard    nofile          64000
root            soft    nofile          32000
@dba            hard    memlock         unlimited
`

var limitsMatchText = `# yadi yadi yada
# /etc/security/limits.conf
* hard nproc 1234
* soft nproc 4000
root hard nproc 16000
root soft nproc 8000

some random soft nofile 12345

* hard nofile 32000
* soft nofile 16000
root hard nofile 64000
root soft nofile 32000
@dba hard memlock unlimited
@sapsys soft nofile 65535
@dba soft memlock unlimited
`

func TestSecLimits(t *testing.T) {
	// Parse the sample text
	limits := ParseSecLimits(limitsSampleText)
	// Read keys
	if value, exists := limits.Get("*", "soft", "nproc"); !exists || value != "4000" {
		t.Fatal(value, exists)
	}
	if value, exists := limits.Get("@dba", "hard", "memlock"); !exists || value != "unlimited" {
		t.Fatal(value, exists)
	}
	if value, exists := limits.Get("does_not_exist", "soft", "nproc"); exists {
		t.Fatal(value, exists)
	}
	// Write keys
	limits.Set("*", "hard", "nproc", "1234")
	limits.Set("@sapsys", "soft", "nofile", "65535")
	limits.Set("@dba", "soft", "memlock", "unlimited")
	// The converted back text should carry new value for nproc and new entry
	if txt := limits.ToText(); txt != limitsMatchText {
		fmt.Println("==============")
		fmt.Println(txt)
		fmt.Println("==============")
		t.Fatal("failed to convert back into text")
	}
}
