package podcounter

import (
	"context"
	"encoding/json"
	v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	"strconv"
)

const Annotation = "ThunderingHerdScheduling/Count"

type PodCounterInterface interface {
	CurrentCounter(pod *v1.Pod) int
	IncrementCounter(pod *v1.Pod) (int, error)
	SetCounter(pod *v1.Pod, val int) error
}

type Counter struct {
	client kubernetes.Interface
}

func New(client kubernetes.Interface) PodCounterInterface {
	return Counter{
		client,
	}
}

func (c Counter) CurrentCounter(pod *v1.Pod) int {
	if strVal, exists := pod.Annotations[Annotation]; exists {
		val, err := strconv.Atoi(strVal)
		if err != nil {
			klog.Error("Failed to parse annotation %s", Annotation, err)
			return 0
		}

		return val
	}

	return 0
}

func (c Counter) IncrementCounter(pod *v1.Pod) (int, error) {
	counter := c.CurrentCounter(pod)
	counter += 1
	err := c.SetCounter(pod, counter)
	return counter, err
}

func (c Counter) SetCounter(pod *v1.Pod, val int) error {
	patch := struct {
		Metadata struct {
			Annotations map[string]string `json:"annotations"`
		} `json:"metadata"`
	}{}
	patch.Metadata.Annotations = map[string]string{}
	patch.Metadata.Annotations[Annotation] = strconv.Itoa(val)
	patchJson, _ := json.Marshal(patch)

	_, err := c.client.CoreV1().Pods(pod.Namespace).Patch(context.TODO(), pod.Name, types.MergePatchType, patchJson, meta_v1.PatchOptions{})
	return err
}
