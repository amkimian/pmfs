package shell

import "testing"

func TestCommandParsing(m *testing.T) {
	cp := CommandParser{}
	executor := ShellExecutor{}
	cp.parse("cd /alan", executor)
	if cp.parameters[0] != "/alan" {
		m.Error("Could not parse cd")
	}
}
