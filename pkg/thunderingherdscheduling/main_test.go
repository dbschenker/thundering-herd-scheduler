package thunderingherdscheduling

import (
	"context"
	"errors"
	v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/kubernetes/pkg/scheduler/framework"
	"sync"
	"testing"
	"time"
)

func TestShouldScheduleDirectlyAsThereAreNoUnreadyPods(t *testing.T) {
	scheduler := getTestingScheduler(0, 0)
	state := &framework.CycleState{}
	pod := getStartingPod("test-pod", "test-namespace", "uuid", true)

	resp, _ := scheduler.Permit(context.TODO(), state, &pod, "test-node")

	if resp.Code() != framework.Success {
		t.Errorf("Failed to schedule pod, expected response code Success, but got %s", resp.Code())
	}
}

func TestShouldScheduleDirectlyAsRetryCountExceeded(t *testing.T) {
	scheduler := getTestingScheduler(5, 6)
	state := &framework.CycleState{}
	pod := getStartingPod("test-pod", "test-namespace", "uuid", true)

	resp, _ := scheduler.Permit(context.TODO(), state, &pod, "test-node")
	if resp.Code() != framework.Success {
		t.Errorf("Failed to schedule pod, expected response code Success, but got %s", resp.Code())
	}
}

func TestShouldReturnWaitWhenTooManyNotReadyPodsAreInPlace(t *testing.T) {
	scheduler := getTestingScheduler(0, 6)
	state := &framework.CycleState{}
	pod := getStartingPod("test-pod", "test-namespace", "uuid", true)

	resp, waitTime := scheduler.Permit(context.TODO(), state, &pod, "test-node")
	if resp.Code() != framework.Wait {
		t.Errorf("Failed to schedule pod, expected response code Wait, but got %s", resp.Code())
	}

	if waitTime != 25*time.Second {
		t.Errorf("Scheduler returned wrong waitTime, expected 25 seconds, but got %f", waitTime.Seconds())
	}
}

func TestShouldReturnWaitBasedOnRetry(t *testing.T) {
	scheduler := getTestingScheduler(1, 6)
	state := &framework.CycleState{}
	pod := getStartingPod("test-pod", "test-namespace", "uuid", true)

	_, waitTime := scheduler.Permit(context.TODO(), state, &pod, "test-node")
	if waitTime != 50*time.Second {
		t.Errorf("Scheduler returned wrong waitTime, expected 50 seconds, but got %f", waitTime.Seconds())
	}
}

func TestShedulerShouldContinueIfCounterFails(t *testing.T) {
	scheduler := getTestingScheduler(1, 6)
	scheduler.counter = PodCounterTest{
		counter:   1,
		exception: errors.New("testerror occured"),
	}
	state := &framework.CycleState{}
	pod := getStartingPod("test-pod", "test-namespace", "uuid", true)

	resp, _ := scheduler.Permit(context.TODO(), state, &pod, "test-node")
	if resp.Code() != framework.Success {
		t.Errorf("Failed to schedule pod, expected response code Success, but got %s", resp.Code())
	}
}

func TestFulfillmentOfInterface(t *testing.T) {
	scheduler := getTestingScheduler(1, 6)

	if scheduler.Name() != "ThunderingHerdScheduling" {
		t.Errorf("Expected Scheduler to return ThunderingHerdScheduling as name, but got %s", scheduler.Name())
	}
}

func getTestingScheduler(retryCounter int, notReadyPods int) *ThunderingHerdScheduling {
	var m sync.Mutex
	counter := PodCounterTest{
		counter: retryCounter,
	}
	nodeState := NodeStateTest{
		notReadyPods: notReadyPods,
	}
	args := &ThunderingHerdSchedulingArgs{}
	SetDefaultThunderingHerdArgs(args)

	scheduler := &ThunderingHerdScheduling{
		counter:   counter,
		args:      args,
		nodestate: nodeState,
		mutex:     &m,
	}

	return scheduler
}

type NodeStateTest struct {
	notReadyPods int
}

func (n NodeStateTest) NotReadyPods(_ string) int {
	return n.notReadyPods
}

func (n NodeStateTest) AddSchedulingPod(_ *v1.Pod, _ string) {
	n.notReadyPods = n.notReadyPods + 1
}

type PodCounterTest struct {
	counter   int
	exception error
}

func (p PodCounterTest) CurrentCounter(_ *v1.Pod) int {
	return p.counter
}

func (p PodCounterTest) IncrementCounter(_ *v1.Pod) (int, error) {
	p.counter = p.counter + 1
	return p.counter, p.exception
}

func (p PodCounterTest) SetCounter(_ *v1.Pod, val int) error {
	p.counter = val
	return nil
}

func getStartingPod(name string, namespace string, uuid string, container bool) v1.Pod {
	objMeta := meta_v1.ObjectMeta{
		Name:      name,
		Namespace: namespace,
		UID:       types.UID(uuid),
	}
	p := v1.Pod{
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
