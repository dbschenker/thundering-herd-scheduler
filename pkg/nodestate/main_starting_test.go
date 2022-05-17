package nodestate

import (
	v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"testing"
)

func TestAddStartingPodWithoutContainerStatus(t *testing.T) {
	p := getStartingPod("pod", "default", "uuid", "test-node", false)

	n := NodeState{
		nodeMap: map[string]NodeStateModel{},
	}
	n.addPod(&p)

	if _, ok := n.nodeMap["test-node"]; !ok {
		t.Errorf("No value added to nodeMap for node: test-node")
	}
	nodeState := n.nodeMap["test-node"]
	podsInPhaseComperator(t, nodeState, 1, 0, 0, 0)
}

func TestAddStartingPodWithStatus(t *testing.T) {
	p := getStartingPod("pod", "default", "uuid", "test-node", true)

	n := NodeState{
		nodeMap: map[string]NodeStateModel{},
	}
	n.addPod(&p)

	if _, ok := n.nodeMap["test-node"]; !ok {
		t.Errorf("No value added to nodeMap for node: test-node")
	}
	nodeState := n.nodeMap["test-node"]
	podsInPhaseComperator(t, nodeState, 1, 0, 0, 0)
}

func TestShouldDoNothingWhileAddingPodTwice(t *testing.T) {
	p := getStartingPod("pod", "default", "uuid", "test-node", true)

	n := NodeState{
		nodeMap: map[string]NodeStateModel{},
	}
	n.addPod(&p)

	n.addPod(&p)

	if _, ok := n.nodeMap["test-node"]; !ok {
		t.Errorf("No value added to nodeMap for node: test-node")
	}
	nodeState := n.nodeMap["test-node"]
	podsInPhaseComperator(t, nodeState, 1, 0, 0, 0)
}

func TestShouldRemoveStartingPodDueToDelete(t *testing.T) {
	p := getStartingPod("pod", "default", "uuid", "test-node", true)

	n := NodeState{
		nodeMap: map[string]NodeStateModel{},
	}
	n.addPod(&p)

	nodeState := n.nodeMap["test-node"]
	podsInPhaseComperator(t, nodeState, 1, 0, 0, 0)

	n.deletePod(&p)
	newNodeState := n.nodeMap["test-node"]
	podsInPhaseComperator(t, newNodeState, 0, 0, 0, 0)
}

func TestShouldDoNothingIfNodeNameIsEmpty(t *testing.T) {
	p := getStartingPod("pod", "default", "uuid", "", true)

	n := NodeState{
		nodeMap: map[string]NodeStateModel{},
	}
	n.addPod(&p)

	if len(n.nodeMap) != 0 {
		t.Errorf("pod was added to nodemap, but it was not scheduled yet")
	}
}

func TestValidNumberOfUnhealthyPodsInStartingPhase(t *testing.T) {
	pod1 := getStartingPod("pod", "default", "uuid", "test-node", true)
	pod2 := getStartingPod("pod1", "default", "uuid", "test-node", true)
	pod3 := getStartingPod("pod2", "default", "uuid", "test-node-1", true)
	pod4 := getStartingPod("pod3", "default", "uuid", "test-node", true)

	n := NodeState{
		nodeMap: map[string]NodeStateModel{},
	}
	n.addPod(&pod1)
	n.addPod(&pod2)
	n.addPod(&pod3)
	n.AddSchedulingPod(&pod4, "test-node")

	unhealthyPodsOnTestNode := n.UnhealthyPods("test-node")
	if unhealthyPodsOnTestNode != 3 {
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

func podsInPhaseComperator(t *testing.T, n NodeStateModel, expectedStartingPods int, expectedRunningPods int, expectedUnhealthyPods int, expectedSchedulingPods int) {
	if len(n.startingPods) != expectedStartingPods {
		t.Errorf("Wrong number of starting pods calculated, expected %d, got %d", expectedStartingPods, len(n.startingPods))
	}

	if len(n.runningPods) != expectedRunningPods {
		t.Errorf("Wrong number of running pods calculated, expected %d, got %d", expectedRunningPods, len(n.runningPods))
	}

	if len(n.unhealthyPods) != expectedUnhealthyPods {
		t.Errorf("Wrong number of unhealthy pods calculated, expected %d, got %d", expectedUnhealthyPods, len(n.unhealthyPods))
	}

	if len(n.schedulingPods) != expectedSchedulingPods {
		t.Errorf("Wrong number of scheduling pods calculated, expected %d, got %d", expectedSchedulingPods, len(n.schedulingPods))
	}
}

func getStartingPod(name string, namespace string, uuid string, nodeName string, container bool) v1.Pod {
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

	if container {
		status := v1.ContainerStatus{
			Name:         "test-container",
			Ready:        false,
			RestartCount: 0,
		}
		p.Status.ContainerStatuses = append(p.Status.ContainerStatuses, status)
	}

	return p
}
