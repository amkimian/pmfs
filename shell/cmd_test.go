package shell

import "testing"

func TestCommandParsing(m *testing.T) {
	cp := CommandParser{}
	cp.parse("cd /alan")
	if cp.parameters[0] != "/alan" {
		m.Error("Could not parse cd")
	}
}
