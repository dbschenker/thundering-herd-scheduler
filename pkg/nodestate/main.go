package nodestate

import (
	"fmt"
	v1 "k8s.io/api/core/v1"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"time"
)

type NodeStateInterface interface {
	UnhealthyPods(nodeName string) int
	AddSchedulingPod(pod *v1.Pod, nodeName string)
}

type NodeState struct {
	nodeMap map[string]NodeStateModel
}

func NewNodeState(client kubernetes.Interface) NodeStateInterface {
	n := NodeState{
		nodeMap: map[string]NodeStateModel{},
	}

	informerFactory := kubeinformers.NewSharedInformerFactory(client, 0)
	podInformer := informerFactory.Core().V1().Pods().Informer()
	podInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		DeleteFunc: func(obj interface{}) {
			pod := obj.(*v1.Pod)
			n.deletePod(pod)
			n.cleanupExpirablePods()
			n.updateMetrics()
			//n.printAllPodsInAllStates()
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			pod := newObj.(*v1.Pod)
			n.addPod(pod)
			n.cleanupExpirablePods()
			n.updateMetrics()
			//n.printAllPodsInAllStates()
		},
		AddFunc: func(obj interface{}) {
			pod := obj.(*v1.Pod)
			n.addPod(pod)
			n.cleanupExpirablePods()
			n.updateMetrics()
			//n.printAllPodsInAllStates()
		},
	})
	go func() {
		podInformer.Run(make(chan struct{}))
	}()

	return &n
}

//func (n *NodeState) printAllPodsInAllStates() {
//	for node, model := range n.nodeMap {
//		klog.Infof("State for pods on node %s starting=%s; running=%s; unhealthy=%s", node, strings.Join(model.startingPods, ","), strings.Join(model.runningPods, ","), strings.Join(model.unhealthyPods, ","))
//	}
//}

func (n *NodeState) cleanupExpirablePods() {
	for node, nodeState := range n.nodeMap {
		nodeState.schedulingPods = cleanupExpired(nodeState.schedulingPods)
		nodeState.unhealthyPods = cleanupExpired(nodeState.unhealthyPods)
		nodeState.startingPods = cleanupExpired(nodeState.startingPods)
		nodeState.runningPods = cleanupExpired(nodeState.runningPods)
		n.nodeMap[node] = nodeState
	}
}

func (n *NodeState) UnhealthyPods(nodeName string) int {
	if val, ok := n.nodeMap[nodeName]; ok {
		return countNonExpiredByType(val.unhealthyPods) + countNonExpiredByType(val.startingPods) + countNonExpiredByType(val.schedulingPods)
	}

	return 0
}

func (n *NodeState) AddSchedulingPod(pod *v1.Pod, nodeName string) {
	assignedPod := pod.DeepCopy()
	assignedPod.Spec.NodeName = nodeName

	key := podStoringKey(assignedPod)

	n.initializeNodeMap(assignedPod.Spec.NodeName)
	n.removeAlreadyExistingPodVersionsFromAllNodes(assignedPod)

	node := n.nodeMap[assignedPod.Spec.NodeName]
	if _, found := find(node.schedulingPods, key); !found {
		expiry := time.Now().Add(1 * time.Minute)
		entry := PodNodeStateModel{
			key:    key,
			expiry: &expiry,
		}
		node.schedulingPods = append(node.schedulingPods, entry)
	}
	n.nodeMap[assignedPod.Spec.NodeName] = node

	n.updateMetrics()
}

func (n *NodeState) deletePod(pod *v1.Pod) {
	n.removeAlreadyExistingPodVersionsFromAllNodes(pod)
}

func (n *NodeState) removeAlreadyExistingPodVersionsFromAllNodes(pod *v1.Pod) {
	key := podStoringKey(pod)

	for node, model := range n.nodeMap {
		model.runningPods = remove(model.runningPods, key)
		model.startingPods = remove(model.startingPods, key)
		model.unhealthyPods = remove(model.unhealthyPods, key)
		model.schedulingPods = remove(model.schedulingPods, key)
		n.nodeMap[node] = model
	}
}

func (n *NodeState) addPod(pod *v1.Pod) {
	if pod.Spec.NodeName == "" {
		return
	}

	n.initializeNodeMap(pod.Spec.NodeName)

	ready := allContainerReady(pod)
	restartingContainer := hasRestartingContainer(pod)

	// jobs which are executed successful or failed have a pod phase of failed or succeeded
	if pod.Status.Phase == v1.PodSucceeded || pod.Status.Phase == v1.PodFailed {
		n.deletePod(pod)
		return
	}

	// pod is starting
	if len(pod.Status.ContainerStatuses) == 0 {
		n.addStartingPod(pod)
		return
	}

	// containers not yet ready, but no restarts; that's most probably a starting pod
	if !ready && !restartingContainer {
		n.addStartingPod(pod)
		return
	}

	if ready { // all containers are ready, we are in running phase
		n.addRunningPod(pod)
		return
	} else { // not all pods are healthy
		n.addUnhealthyPod(pod)
		return
	}
}

func (n *NodeState) updateMetrics() {
	for node, model := range n.nodeMap {
		startCount := countNonExpiredByType(model.startingPods)
		startingPodsMetric.WithLabelValues(node).Set(float64(startCount))

		runningCount := countNonExpiredByType(model.runningPods)
		runningPodsMetric.WithLabelValues(node).Set(float64(runningCount))

		unhealthyCount := countNonExpiredByType(model.unhealthyPods)
		unhealthyPodsMetric.WithLabelValues(node).Set(float64(unhealthyCount))

		schedulingCount := countNonExpiredByType(model.schedulingPods)
		schedulingPodsMetric.WithLabelValues(node).Set(float64(schedulingCount))
	}
}

func (n *NodeState) initializeNodeMap(nodeName string) {
	if _, ok := n.nodeMap[nodeName]; !ok {
		n.nodeMap[nodeName] = NodeStateModel{
			startingPods:   []PodNodeStateModel{},
			runningPods:    []PodNodeStateModel{},
			unhealthyPods:  []PodNodeStateModel{},
			schedulingPods: []PodNodeStateModel{},
		}
	}
}

func (n *NodeState) addStartingPod(pod *v1.Pod) {
	key := podStoringKey(pod)
	if pod.Spec.NodeName == "" {
		return
	}

	n.initializeNodeMap(pod.Spec.NodeName)
	n.removeAlreadyExistingPodVersionsFromAllNodes(pod)

	node := n.nodeMap[pod.Spec.NodeName]
	if _, found := find(node.startingPods, key); !found {
		entry := PodNodeStateModel{
			key: key,
		}
		node.startingPods = append(node.startingPods, entry)
	}
	n.nodeMap[pod.Spec.NodeName] = node
}

func (n *NodeState) addRunningPod(pod *v1.Pod) {
	key := podStoringKey(pod)
	if pod.Spec.NodeName == "" {
		return
	}

	n.initializeNodeMap(pod.Spec.NodeName)
	n.removeAlreadyExistingPodVersionsFromAllNodes(pod)

	node := n.nodeMap[pod.Spec.NodeName]
	if _, found := find(node.runningPods, key); !found {
		entry := PodNodeStateModel{
			key: key,
		}
		node.runningPods = append(node.runningPods, entry)
	}
	n.nodeMap[pod.Spec.NodeName] = node
}

func (n *NodeState) addUnhealthyPod(pod *v1.Pod) {
	key := podStoringKey(pod)
	if pod.Spec.NodeName == "" {
		return
	}

	n.initializeNodeMap(pod.Spec.NodeName)
	n.removeAlreadyExistingPodVersionsFromAllNodes(pod)

	node := n.nodeMap[pod.Spec.NodeName]
	if _, found := find(node.unhealthyPods, key); !found {
		entry := PodNodeStateModel{
			key: key,
		}
		node.unhealthyPods = append(node.unhealthyPods, entry)
	}
	n.nodeMap[pod.Spec.NodeName] = node
}

func allContainerReady(pod *v1.Pod) bool {
	if len(pod.Status.ContainerStatuses) == 0 {
		return true
	}

	for _, containerStatus := range pod.Status.ContainerStatuses {
		if containerStatus.Ready != true {
			return false
		}
	}

	return true
}

func hasRestartingContainer(pod *v1.Pod) bool {
	if len(pod.Status.ContainerStatuses) == 0 {
		return false
	}

	for _, containerStatus := range pod.Status.ContainerStatuses {
		if containerStatus.Ready == false && containerStatus.RestartCount > 0 {
			return true
		}
	}

	return false
}

func remove(s []PodNodeStateModel, r string) []PodNodeStateModel {
	for i, v := range s {
		if v.key == r {
			return append(s[:i], s[i+1:]...)
		}
	}
	return s
}

func find(slice []PodNodeStateModel, val string) (int, bool) {
	for i, item := range slice {
		if item.key == val {
			return i, true
		}
	}
	return -1, false
}

func podStoringKey(pod *v1.Pod) string {
	return fmt.Sprintf("%s-%s-%s", pod.Name, pod.Namespace, pod.UID)
}

func cleanupExpired(nodeStateEntry []PodNodeStateModel) []PodNodeStateModel {
	now := time.Now()
	for _, entry := range nodeStateEntry {
		if entry.expiry != nil && entry.expiry.Before(now) {
			nodeStateEntry = remove(nodeStateEntry, entry.key)
		}
	}
	return nodeStateEntry
}

func countNonExpiredByType(podsInStateModel []PodNodeStateModel) int {
	count := 0
	for _, entry := range podsInStateModel {
		if entry.expiry == nil {
			count++
		} else {
			if entry.expiry.After(time.Now()) {
				count++
			}
		}
	}
	return count
}
