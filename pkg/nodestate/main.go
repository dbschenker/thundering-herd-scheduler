package nodestate

import (
	"fmt"
	v1 "k8s.io/api/core/v1"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
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
			n.updateMetrics()
			//n.printAllPodsInAllStates()
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			pod := newObj.(*v1.Pod)
			n.addPod(pod)
			n.updateMetrics()
			//n.printAllPodsInAllStates()
		},
		AddFunc: func(obj interface{}) {
			pod := obj.(*v1.Pod)
			n.addPod(pod)
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

func (n *NodeState) UnhealthyPods(nodeName string) int {
	if val, ok := n.nodeMap[nodeName]; ok {
		return len(val.unhealthyPods) + len(val.startingPods)
	}

	return 0
}

func (n *NodeState) AddSchedulingPod(pod *v1.Pod, nodeName string) {
	assignedPod := pod.DeepCopy()
	assignedPod.Spec.NodeName = nodeName
	n.addStartingPod(assignedPod)
}

func (n *NodeState) deletePod(pod *v1.Pod) {
	n.deletePodFromStarting(pod)
	n.deletePodFromRunning(pod)
	n.deletePodFromUnhealthy(pod)
}

func (n *NodeState) deletePodFromStarting(pod *v1.Pod) {
	if pod.Spec.NodeName != "" {
		if val, ok := n.nodeMap[pod.Spec.NodeName]; ok {
			key := podStoringKey(pod)
			val.startingPods = remove(val.startingPods, key)
			n.nodeMap[pod.Spec.NodeName] = val
		}
	}
}

func (n *NodeState) deletePodFromRunning(pod *v1.Pod) {
	if pod.Spec.NodeName != "" {
		if val, ok := n.nodeMap[pod.Spec.NodeName]; ok {
			key := podStoringKey(pod)
			val.runningPods = remove(val.runningPods, key)
			n.nodeMap[pod.Spec.NodeName] = val
		}
	}
}

func (n *NodeState) deletePodFromUnhealthy(pod *v1.Pod) {
	if pod.Spec.NodeName != "" {
		if val, ok := n.nodeMap[pod.Spec.NodeName]; ok {
			key := podStoringKey(pod)
			val.unhealthyPods = remove(val.unhealthyPods, key)
			n.nodeMap[pod.Spec.NodeName] = val
		}
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
		startCount := len(model.startingPods)
		startingPodsMetric.WithLabelValues(node).Set(float64(startCount))

		runningCount := len(model.runningPods)
		runningPodsMetric.WithLabelValues(node).Set(float64(runningCount))

		unhealthyCount := len(model.unhealthyPods)
		unhealthyPodsMetric.WithLabelValues(node).Set(float64(unhealthyCount))
	}
}

func (n *NodeState) initializeNodeMap(nodeName string) {
	if _, ok := n.nodeMap[nodeName]; !ok {
		n.nodeMap[nodeName] = NodeStateModel{
			startingPods:  []string{},
			runningPods:   []string{},
			unhealthyPods: []string{},
		}
	}
}

func (n *NodeState) addStartingPod(pod *v1.Pod) {
	key := podStoringKey(pod)
	if pod.Spec.NodeName == "" {
		return
	}

	n.initializeNodeMap(pod.Spec.NodeName)

	node := n.nodeMap[pod.Spec.NodeName]
	if _, found := find(node.startingPods, key); !found {
		node.startingPods = append(node.startingPods, key)
	}
	n.nodeMap[pod.Spec.NodeName] = node

	n.deletePodFromRunning(pod)
	n.deletePodFromUnhealthy(pod)
}

func (n *NodeState) addRunningPod(pod *v1.Pod) {
	key := podStoringKey(pod)
	if pod.Spec.NodeName == "" {
		return
	}

	n.initializeNodeMap(pod.Spec.NodeName)

	node := n.nodeMap[pod.Spec.NodeName]
	if _, found := find(node.runningPods, key); !found {
		node.runningPods = append(node.runningPods, key)
	}
	n.nodeMap[pod.Spec.NodeName] = node

	n.deletePodFromStarting(pod)
	n.deletePodFromUnhealthy(pod)
}

func (n *NodeState) addUnhealthyPod(pod *v1.Pod) {
	key := podStoringKey(pod)
	if pod.Spec.NodeName == "" {
		return
	}

	n.initializeNodeMap(pod.Spec.NodeName)

	node := n.nodeMap[pod.Spec.NodeName]
	if _, found := find(node.unhealthyPods, key); !found {
		node.unhealthyPods = append(node.unhealthyPods, key)
	}
	n.nodeMap[pod.Spec.NodeName] = node

	n.deletePodFromStarting(pod)
	n.deletePodFromRunning(pod)
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

func remove(s []string, r string) []string {
	for i, v := range s {
		if v == r {
			return append(s[:i], s[i+1:]...)
		}
	}
	return s
}

func find(slice []string, val string) (int, bool) {
	for i, item := range slice {
		if item == val {
			return i, true
		}
	}
	return -1, false
}

func podStoringKey(pod *v1.Pod) string {
	return fmt.Sprintf("%s-%s-%s", pod.Name, pod.Namespace, pod.UID)
}
