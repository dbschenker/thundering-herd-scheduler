package nodestate

import (
	v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"testing"
)

func TestShouldRemovePodIfInSuccessPhase(t *testing.T) {
	runningPod := getRunningPod("test-pod", "test-namespace", "test-uuid", "test-node")
	n := NodeState{
		nodeMap: map[string]NodeStateModel{},
	}
	n.addPod(&runningPod)

	nodeState := n.nodeMap["test-node"]
	podsInPhaseComperator(t, nodeState, 0, 1, 0)

	completedPod := getPodInPhase("test-pod", "test-namespace", "test-uuid", "test-node", v1.PodSucceeded)
	n.addPod(&completedPod)

	newNodeState := n.nodeMap["test-node"]
	podsInPhaseComperator(t, newNodeState, 0, 0, 0)
}

func TestShouldRemovePodInFailedPhase(t *testing.T) {
	runningPod := getRunningPod("test-pod", "test-namespace", "test-uuid", "test-node")
	n := NodeState{
		nodeMap: map[string]NodeStateModel{},
	}
	n.addPod(&runningPod)

	nodeState := n.nodeMap["test-node"]
	podsInPhaseComperator(t, nodeState, 0, 1, 0)

	completedPod := getPodInPhase("test-pod", "test-namespace", "test-uuid", "test-node", v1.PodFailed)
	n.addPod(&completedPod)

	newNodeState := n.nodeMap["test-node"]
	podsInPhaseComperator(t, newNodeState, 0, 0, 0)
}

func getPodInPhase(name string, namespace string, uuid string, nodeName string, phase v1.PodPhase) v1.Pod {
	return v1.Pod{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			UID:       types.UID(uuid),
		},
		Spec: v1.PodSpec{
			NodeName: nodeName,
		},
		Status: v1.PodStatus{
			Phase: phase,
		},
	}
}
