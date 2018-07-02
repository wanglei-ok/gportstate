package portstate

import "testing"

func TestGetTcpPortState(t *testing.T) {
	s, ok := GetTcpPortState(8333)
	t.Logf("GetTcpPortState: %v,%v", s, ok )
}