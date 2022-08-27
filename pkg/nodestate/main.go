package nodestate

import (
	v1 "k8s.io/api/core/v1"
)

type NodeStateInterface interface {
	NotReadyPods(nodeName string) int
	AddSchedulingPod(pod *v1.Pod, nodeName string)
}
