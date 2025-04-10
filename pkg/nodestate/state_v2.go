package nodestate

import (
	"context"
	"fmt"
	"github.com/benbjohnson/clock"
	"k8s.io/apimachinery/pkg/api/resource"
	"math"

	v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	"sync"
	"time"
)

type NodeStateV2 struct {
	scheduledPods map[string][]string
	client        kubernetes.Interface
	lock          *sync.RWMutex
	clock         clock.Clock
}

func NewNodeStateV2(client kubernetes.Interface) NodeStateInterface {
	return internalNewNodeStateV2(client, clock.New())
}

func internalNewNodeStateV2(client kubernetes.Interface, c clock.Clock) NodeStateInterface {
	var lock = sync.RWMutex{}
	return &NodeStateV2{
		client:        client,
		scheduledPods: make(map[string][]string),
		lock:          &lock,
		clock:         c,
	}
}

func (n *NodeStateV2) NotReadyPodsAllowedInParallel(parallelStartingPodsPerNode *int, parallelStartingPodsPerCore *float64, nodeName string) (int, error) {
	if parallelStartingPodsPerNode != nil {
		return *parallelStartingPodsPerNode, nil
	}

	node, err := n.client.CoreV1().Nodes().Get(context.TODO(), nodeName, meta_v1.GetOptions{})
	if err != nil {
		return -1, fmt.Errorf("node %s can't be queried from api server: %v", nodeName, err)
	}

	allocatableCpu := node.Status.Allocatable.Cpu()
	ret := calculateParallelStartingPodsPerCore(*parallelStartingPodsPerCore, allocatableCpu)

	return ret, nil
}

func (n *NodeStateV2) NotReadyPods(nodeName string) int {
	// copied from https://github.com/kubernetes/kubernetes/blob/4f2d7b93da2464a3147e0a7e71d896dd2bade9ad/pkg/printers/internalversion/describe.go#L2451
	fieldSelector, err := fields.ParseSelector("spec.nodeName=" + nodeName + ",status.phase!=" + string(v1.PodSucceeded) + ",status.phase!=" + string(v1.PodFailed))
	if err != nil {
		klog.Errorf("Failed to create field selector for running pod calculation with error %v", err)
		return -1
	}

	nodeNonTerminatedPodsList, err := n.client.CoreV1().Pods("").List(context.TODO(), meta_v1.ListOptions{FieldSelector: fieldSelector.String()})
	if err != nil {
		klog.Errorf("Failed to list pods on node %s with error %v", nodeName, err)
		return -1
	}

	notReadyPods := 0
	for _, pod := range nodeNonTerminatedPodsList.Items {
		if !isPodReady(pod) {
			notReadyPods++
		}
	}

	return notReadyPods + n.scheduledPodsOnNode(nodeName)
}

func (n *NodeStateV2) AddSchedulingPod(pod *v1.Pod, nodeName string) {
	podKey := podStoringKey(pod)

	n.lock.Lock()
	defer n.lock.Unlock()

	if _, ok := n.scheduledPods[nodeName]; !ok {
		n.scheduledPods[nodeName] = []string{}
	}

	if !contains(n.scheduledPods[nodeName], podKey) {
		n.scheduledPods[nodeName] = append(n.scheduledPods[nodeName], podKey)
	}

	n.clock.AfterFunc(5*time.Second, func() {
		n.lock.Lock()
		defer n.lock.Unlock()

		n.scheduledPods[nodeName] = remove(n.scheduledPods[nodeName], podKey)
	})
}

func (n *NodeStateV2) scheduledPodsOnNode(nodeName string) int {
	n.lock.RLock()
	defer n.lock.RUnlock()

	if _, ok := n.scheduledPods[nodeName]; ok {
		return len(n.scheduledPods[nodeName])
	}

	return 0
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func remove(s []string, r string) []string {
	for i, v := range s {
		if v == r {
			return append(s[:i], s[i+1:]...)
		}
	}
	return s
}

func podStoringKey(pod *v1.Pod) string {
	return fmt.Sprintf("%s-%s-%s", pod.Name, pod.Namespace, pod.UID)
}

// copied from https://github.com/helm/helm/blob/d7b4c38c42cb0b77f1bcebf9bb4ae7695a10da0b/pkg/kube/ready.go#L215
func isPodReady(pod v1.Pod) bool {
	for _, c := range pod.Status.Conditions {
		if c.Type == v1.PodReady && c.Status == v1.ConditionTrue {
			return true
		}
	}
	return false
}

// regardless of number of cores in order to avoid starvation, at least one node can be scheduled
func calculateParallelStartingPodsPerCore(podsPerCore float64, cpu *resource.Quantity) int {
	val := cpu.AsApproximateFloat64() * podsPerCore
	if val < 1 {
		return 1
	}
	return int(math.Floor(val))
}
