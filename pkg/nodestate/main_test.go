package nodestate

import "testing"

func TestShouldReturnValidValueIfNotYetInitialized(t *testing.T) {
	n := NodeState{
		nodeMap: map[string]NodeStateModel{},
	}

	unhealthy := n.UnhealthyPods("test-node")
	if unhealthy != 0 {
		t.Errorf("Got invalid number of unhealthy pods, expected 0 but got %d", unhealthy)
	}
}
