package nodestate

import (
	"context"
	"github.com/benbjohnson/clock"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	testclient "k8s.io/client-go/kubernetes/fake"
	"testing"
	"time"
)

func TestShouldCountNotReadyPodsWithZero(t *testing.T) {
	client := testclient.NewSimpleClientset()

	pods := []v1.Pod{
		mockRunningPod("test-pod", "ns-1", "11b666eb-a361-4b4e-8953-f88224462564", "node-1"),
		mockRunningPod("test-pod-2", "ns-2", "d14b61cd-4a3a-477e-ac0a-2b2c50f301ee", "node-1"),
		mockRunningPod("test-pod-3", "ns-3", "36847994-2dae-46e3-8ee5-af6afc2a5d63", "node-1"),
	}

	for _, pod := range pods {
		client.CoreV1().Pods(pod.Namespace).Create(context.TODO(), &pod, meta_v1.CreateOptions{})
	}

	stateV2 := NewNodeStateV2(client)

	notReadyPods := stateV2.NotReadyPods("node-1")
	if notReadyPods != 0 {
		t.Errorf("Expected 0 unhealthy pods but got %d", notReadyPods)
	}
}

func TestShouldCountNotReadyPodsAndFindSome(t *testing.T) {
	client := testclient.NewSimpleClientset()

	pods := []v1.Pod{
		mockRunningPod("test-pod", "ns-1", "11b666eb-a361-4b4e-8953-f88224462564", "node-1"),
		mockUnhealthyPod("test-pod-2", "ns-2", "a8c0c923-2d28-4e18-85c0-3023ad460d8e", "node-1"),
		mockUnhealthyPod("test-pod-3", "ns-3", "8fc4799d-8181-426a-8247-0371f9f6fbeb", "node-1"),
	}

	for _, pod := range pods {
		client.CoreV1().Pods(pod.Namespace).Create(context.TODO(), &pod, meta_v1.CreateOptions{})
	}

	stateV2 := NewNodeStateV2(client)

	notReadyPods := stateV2.NotReadyPods("node-1")
	if notReadyPods != 2 {
		t.Errorf("Expected 2 unhealthy pods but got %d", notReadyPods)
	}
}

func TestShouldAddPodInSchedulingPhaseToInternalList(t *testing.T) {
	client := testclient.NewSimpleClientset()
	stateV2 := NewNodeStateV2(client)

	pod := mockRunningPod("qwe", "asd", "33d30e5a-548d-4c89-9821-f18bc1f9df2c", "node-1")
	stateV2.AddSchedulingPod(&pod, "node-1")

	notReadyPods := stateV2.NotReadyPods("node-1")
	if notReadyPods != 1 {
		t.Errorf("Expected 1 unhealthy pods but got %d", notReadyPods)
	}
}

func TestShouldAddMultiplePodsInSchedulingPhaseToInternalList(t *testing.T) {
	client := testclient.NewSimpleClientset()
	stateV2 := NewNodeStateV2(client)

	pod1 := mockRunningPod("pod-1", "ns-1", "33d30e5a-548d-4c89-9821-f18bc1f9df2c", "node-1")
	pod2 := mockRunningPod("pod-2", "ns-1", "bb0acc1a-46a0-446b-86e4-30dfae9ad450", "node-1")
	stateV2.AddSchedulingPod(&pod1, "node-1")
	stateV2.AddSchedulingPod(&pod2, "node-1")

	notReadyPods := stateV2.NotReadyPods("node-1")
	if notReadyPods != 2 {
		t.Errorf("Expected 2 unhealthy pods but got %d", notReadyPods)
	}
}

func TestShouldNotAppendSchedulingPodMultipleTimes(t *testing.T) {
	client := testclient.NewSimpleClientset()
	stateV2 := NewNodeStateV2(client)

	pod1 := mockRunningPod("pod-1", "ns-1", "33d30e5a-548d-4c89-9821-f18bc1f9df2c", "node-1")

	stateV2.AddSchedulingPod(&pod1, "node-1")
	stateV2.AddSchedulingPod(&pod1, "node-1")
	stateV2.AddSchedulingPod(&pod1, "node-1")

	notReadyPods := stateV2.NotReadyPods("node-1")
	if notReadyPods != 1 {
		t.Errorf("Expected 1 unhealthy pods but got %d", notReadyPods)
	}
}

func TestShouldRemoveFromInSchedulingPodList(t *testing.T) {
	c := clock.NewMock()

	client := testclient.NewSimpleClientset()
	stateV2 := internalNewNodeStateV2(client, c)

	pod1 := mockRunningPod("pod-1", "ns-1", "33d30e5a-548d-4c89-9821-f18bc1f9df2c", "node-1")
	pod2 := mockRunningPod("pod-12", "ns-1", "532ee84e-ad8f-4a5b-99e3-b52ef909226b", "node-1")
	stateV2.AddSchedulingPod(&pod1, "node-1")
	stateV2.AddSchedulingPod(&pod2, "node-1")

	c.Add(6 * time.Second)

	notReadyPods := stateV2.NotReadyPods("node-1")
	if notReadyPods != 0 {
		t.Errorf("Expected 0 unhealthy pods but got %d", notReadyPods)
	}
}

func mockRunningPod(name string, namespace string, uuid string, nodeName string) v1.Pod {
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
			Conditions: []v1.PodCondition{
				{
					Status: v1.ConditionTrue,
					Type:   v1.PodReady,
				},
			},
		},
	}
	p.ObjectMeta = objMeta

	return p
}

func mockUnhealthyPod(name string, namespace string, uuid string, nodeName string) v1.Pod {
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
			Conditions: []v1.PodCondition{},
		},
	}
	p.ObjectMeta = objMeta

	return p
}

func TestCalculateParallelStartingPodsPerCore(t *testing.T) {
	testcases := []struct {
		name        string
		podsPerCore float64
		cpu         string
		expected    int
	}{
		{
			name:        "1 pod per core, 2 cores",
			podsPerCore: 1,
			cpu:         "2",
			expected:    2,
		},
		{
			name:        "0.5 pod per core, 2 cores",
			podsPerCore: 0.5,
			cpu:         "2",
			expected:    1,
		},
		{
			name:        "0.4 pod per core, 2 cores",
			podsPerCore: 0.4,
			cpu:         "2",
			expected:    1,
		},
		{
			name:        "0.4 pod per core, 1.6 cores",
			podsPerCore: 0.4,
			cpu:         "1600m",
			expected:    1,
		},
		{
			name:        "0.3 pod per core, 1.6 cores",
			podsPerCore: 0.3,
			cpu:         "1600m",
			expected:    1,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			q := resource.MustParse(tc.cpu)
			result := calculateParallelStartingPodsPerCore(tc.podsPerCore, &q)
			assert.Equal(t, tc.expected, result)
		})
	}
}
