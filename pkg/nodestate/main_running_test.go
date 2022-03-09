package nodestate

import (
	v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"testing"
)

func TestRunningPod(t *testing.T) {
	p := getRunningPod("test-pod", "default", "uuid", "test-node")

	n := NodeState{
		nodeMap: map[string]NodeStateModel{},
	}
	n.addPod(&p)

	if _, ok := n.nodeMap["test-node"]; !ok {
		t.Errorf("No value added to nodeMap for node: test-node")
	}
	nodeState := n.nodeMap["test-node"]
	podsInPhaseComperator(t, nodeState, 0, 1, 0)
}

func TestRunningPodIdempotence(t *testing.T) {
	p := getRunningPod("test-pod", "default", "uuid", "test-node")

	n := NodeState{
		nodeMap: map[string]NodeStateModel{},
	}
	n.addPod(&p)

	nodeState := n.nodeMap["test-node"]
	podsInPhaseComperator(t, nodeState, 0, 1, 0)

	newNodeState := n.nodeMap["test-node"]
	podsInPhaseComperator(t, newNodeState, 0, 1, 0)
}

func TestPodMovedFromStartingToRunningState(t *testing.T) {
	startingPod := getStartingPod("test-pod", "default", "uuid", "test-node", true)
	runningPod := getRunningPod("test-pod", "default", "uuid", "test-node")

	n := NodeState{
		nodeMap: map[string]NodeStateModel{},
	}
	n.addPod(&startingPod)
	nodeState := n.nodeMap["test-node"]
	podsInPhaseComperator(t, nodeState, 1, 0, 0)

	n.addPod(&runningPod)
	newNodeState := n.nodeMap["test-node"]
	podsInPhaseComperator(t, newNodeState, 0, 1, 0)
}

func TestPodMovedFromUnhealthyToRunningState(t *testing.T) {
	unhealthyPod := getUnhealthyPod("test-pod", "default", "uuid", "test-node")
	runningPod := getRunningPod("test-pod", "default", "uuid", "test-node")

	n := NodeState{
		nodeMap: map[string]NodeStateModel{},
	}
	n.addPod(&unhealthyPod)
	nodeState := n.nodeMap["test-node"]
	podsInPhaseComperator(t, nodeState, 0, 0, 1)

	n.addPod(&runningPod)
	newNodeState := n.nodeMap["test-node"]
	podsInPhaseComperator(t, newNodeState, 0, 1, 0)
}

func TestRemovePod(t *testing.T) {
	runningPod := getRunningPod("test-pod", "default", "uuid", "test-node")

	n := NodeState{
		nodeMap: map[string]NodeStateModel{},
	}
	n.addPod(&runningPod)
	nodeState := n.nodeMap["test-node"]
	podsInPhaseComperator(t, nodeState, 0, 1, 0)

	n.deletePod(&runningPod)
	newNodeState := n.nodeMap["test-node"]
	podsInPhaseComperator(t, newNodeState, 0, 0, 0)
}

func TestValidNumberOfUnhealthyPods(t *testing.T) {
	runningPod1 := getRunningPod("test-pod", "default", "uuid", "test-node")
	runningPod2 := getRunningPod("test-pod-2", "default", "uuid", "test-node")

	n := NodeState{
		nodeMap: map[string]NodeStateModel{},
	}
	n.addPod(&runningPod1)
	n.addPod(&runningPod2)

	unhealthy := n.UnhealthyPods("test-node")
	if unhealthy != 0 {
		t.Errorf("Got invalid number of unhealthy pods, expected 0 but got %d", unhealthy)
	}
}

func getRunningPod(name string, namespace string, uuid string, nodeName string) v1.Pod {
	objMeta := meta_v1.ObjectMeta{
		Name:      name,
		Namespace: namespace,
		UID:       types.UID(uuid),
	}
	p := v1.Pod{
		Spec: v1.PodSpec{
			NodeName: nodeName,
		},
		Status: v1.PodStatus{
			ContainerStatuses: []v1.ContainerStatus{},
		},
	}
	p.ObjectMeta = objMeta

	status := v1.ContainerStatus{
		Name:         "test-container",
		Ready:        true,
		RestartCount: 0,
	}
	p.Status.ContainerStatuses = append(p.Status.ContainerStatuses, status)

	return p
}
