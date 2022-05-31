package nodestate

import (
	v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"testing"
	"time"
)

func TestShouldAddSchedulingPodWithNodeAssigned(t *testing.T) {
	p := v1.Pod{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "test-namespace",
			UID:       types.UID("uuid"),
		},
		Status: v1.PodStatus{
			ContainerStatuses: []v1.ContainerStatus{},
		},
	}

	n := NodeState{
		nodeMap: map[string]NodeStateModel{},
	}
	n.AddSchedulingPod(&p, "test-node")

	nodeState := n.nodeMap["test-node"]
	podsInPhaseComperator(t, nodeState, 0, 0, 0, 1)
}

func TestShouldCleanupExpiredPods(t *testing.T) {
	p1 := v1.Pod{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "test-namespace",
			UID:       types.UID("uuid"),
		},
		Status: v1.PodStatus{
			ContainerStatuses: []v1.ContainerStatus{},
		},
	}
	p2 := v1.Pod{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      "test-pod-2",
			Namespace: "test-namespace",
			UID:       types.UID("uuid-2"),
		},
		Status: v1.PodStatus{
			ContainerStatuses: []v1.ContainerStatus{},
		},
	}
	p3 := v1.Pod{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      "test-pod-3",
			Namespace: "test-namespace",
			UID:       types.UID("uuid-3"),
		},
		Status: v1.PodStatus{
			ContainerStatuses: []v1.ContainerStatus{},
		},
	}

	n := NodeState{
		nodeMap: map[string]NodeStateModel{},
	}
	n.AddSchedulingPod(&p1, "test-node")
	n.AddSchedulingPod(&p2, "test-node")
	n.AddSchedulingPod(&p3, "test-node-1")
	past := time.Now().Add(-5 * time.Minute)
	n.nodeMap["test-node"].schedulingPods[1].expiry = &past

	n.cleanupExpirablePods()

	podsInPhaseComperator(t, n.nodeMap["test-node"], 0, 0, 0, 1)
	podsInPhaseComperator(t, n.nodeMap["test-node-1"], 0, 0, 0, 1)
}

func TestShouldNotCountExpiredPods(t *testing.T) {
	p := v1.Pod{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "test-namespace",
			UID:       types.UID("uuid"),
		},
		Status: v1.PodStatus{
			ContainerStatuses: []v1.ContainerStatus{},
		},
	}

	n := NodeState{
		nodeMap: map[string]NodeStateModel{},
	}
	n.AddSchedulingPod(&p, "test-node")
	beforeCount := n.UnhealthyPods("test-node")
	if beforeCount != 1 {
		t.Errorf("Expected number of scheduling pods on test-node to be 1, but got %d", beforeCount)
	}

	past := time.Now().Add(-5 * time.Minute)
	n.nodeMap["test-node"].schedulingPods[0].expiry = &past

	afterCount := n.UnhealthyPods("test-node")
	if afterCount != 0 {
		t.Errorf("Expected number of scheduling pods on test-node to be 0, but got %d", afterCount)
	}
}
