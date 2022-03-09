package nodestate

import (
	v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"testing"
)

func TestUnhealthyPod(t *testing.T) {
	unhealthyPod := getUnhealthyPod("pod", "default", "uid", "test-node")
	n := NodeState{
		nodeMap: map[string]NodeStateModel{},
	}
	n.addPod(&unhealthyPod)

	if _, ok := n.nodeMap["test-node"]; !ok {
		t.Errorf("No value added to nodeMap for node: test-node")
	}
	nodeState := n.nodeMap["test-node"]
	podsInPhaseComperator(t, nodeState, 0, 0, 1)
}

func TestUnhealthyPodIdempotence(t *testing.T) {
	unhealthyPod := getUnhealthyPod("pod", "default", "uid", "test-node")
	n := NodeState{
		nodeMap: map[string]NodeStateModel{},
	}
	n.addPod(&unhealthyPod)

	nodeState := n.nodeMap["test-node"]
	podsInPhaseComperator(t, nodeState, 0, 0, 1)

	n.addPod(&unhealthyPod)
	newNodeState := n.nodeMap["test-node"]
	podsInPhaseComperator(t, newNodeState, 0, 0, 1)
}

func TestShouldRemoveStartingPodDueToUnhealthy(t *testing.T) {
	startingPod := getStartingPod("pod", "default", "uuid", "test-node", true)
	unhealthyPod := getUnhealthyPod("pod", "default", "uuid", "test-node")

	n := NodeState{
		nodeMap: map[string]NodeStateModel{},
	}
	n.addPod(&startingPod)

	nodeState := n.nodeMap["test-node"]
	podsInPhaseComperator(t, nodeState, 1, 0, 0)

	n.addPod(&unhealthyPod)
	newNodeState := n.nodeMap["test-node"]
	podsInPhaseComperator(t, newNodeState, 0, 0, 1)
}

func TestShouldRemoveRunningPodDueToUnhealthy(t *testing.T) {
	runningPod := getRunningPod("pod", "default", "uuid", "test-node")
	unhealthyPod := getUnhealthyPod("pod", "default", "uuid", "test-node")

	n := NodeState{
		nodeMap: map[string]NodeStateModel{},
	}
	n.addPod(&runningPod)

	nodeState := n.nodeMap["test-node"]
	podsInPhaseComperator(t, nodeState, 0, 1, 0)

	n.addPod(&unhealthyPod)
	newNodeState := n.nodeMap["test-node"]
	podsInPhaseComperator(t, newNodeState, 0, 0, 1)
}

func TestShouldShowValidNumberOfUnhealthyPodsInUnhealthyPhase(t *testing.T) {
	pod1 := getUnhealthyPod("pod-1", "default", "uuid", "test-node")
	pod2 := getUnhealthyPod("pod-2", "default", "uuid", "test-node")
	pod3 := getUnhealthyPod("pod-3", "default", "uuid", "test-node-1")

	n := NodeState{
		nodeMap: map[string]NodeStateModel{},
	}
	n.addPod(&pod1)
	n.addPod(&pod2)
	n.addPod(&pod3)

	unhealthyPodsOnTestNode := n.UnhealthyPods("test-node")
	if unhealthyPodsOnTestNode != 2 {
		t.Errorf("Expected number of unhealthy pods on test-node to be 2, but got %d", unhealthyPodsOnTestNode)
	}

	unhealthyPodsOnTestNode1 := n.UnhealthyPods("test-node-1")
	if unhealthyPodsOnTestNode1 != 1 {
		t.Errorf("Expected number of unhealthy pods on test-node-1 to be 1, but got %d", unhealthyPodsOnTestNode1)
	}

	unhealthyPodsOnTestNode2 := n.UnhealthyPods("test-node-2")
	if unhealthyPodsOnTestNode2 != 0 {
		t.Errorf("Expected number of unhealthy pods on test-node-2 to be 0, but got %d", unhealthyPodsOnTestNode2)
	}
}

func getUnhealthyPod(name string, namespace string, uuid string, nodeName string) v1.Pod {
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
		Ready:        false,
		RestartCount: 1,
	}
	p.Status.ContainerStatuses = append(p.Status.ContainerStatuses, status)

	return p
}
