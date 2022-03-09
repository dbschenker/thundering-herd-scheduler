package podcounter

import (
	"context"
	v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	"testing"
)

func TestShouldReturnNullIfNoCounterSet(t *testing.T) {
	var clientset kubernetes.Interface
	clientset = fake.NewSimpleClientset()

	c := Counter{
		clientset,
	}

	p := getCounterTestPod(false, "")
	currentValue := c.CurrentCounter(&p)

	if currentValue != 0 {
		t.Errorf("Initial counter expected to be 0, but was %d", currentValue)
	}
}

func TestShouldReturnConcreteValueIfCounterSet(t *testing.T) {
	var clientset kubernetes.Interface
	clientset = fake.NewSimpleClientset()

	c := New(clientset)

	p := getCounterTestPod(true, "2")
	currentValue := c.CurrentCounter(&p)

	if currentValue != 2 {
		t.Errorf("Counter expected to be 2, but was %d", currentValue)
	}
}

func TestShouldReturnNullIfNotNumberSet(t *testing.T) {
	var clientset kubernetes.Interface
	clientset = fake.NewSimpleClientset()

	c := Counter{
		clientset,
	}

	p := getCounterTestPod(true, "InvalidNumber")
	currentValue := c.CurrentCounter(&p)

	if currentValue != 0 {
		t.Errorf("Counter expected to be 0, but was %d", currentValue)
	}
}

func TestShouldSetCounterToValue(t *testing.T) {
	var clientset kubernetes.Interface
	p := getCounterTestPod(true, "2")
	clientset = fake.NewSimpleClientset(&p)

	c := Counter{
		clientset,
	}

	err := c.SetCounter(&p, 4)
	if err != nil {
		t.Errorf("Failed to set pod counter to 4 with error %v", err)
	}

	resp, _ := clientset.CoreV1().Pods("test-namespace").Get(context.TODO(), "test-pod", meta_v1.GetOptions{})

	if resp.Annotations[Annotation] != "4" {
		t.Errorf("Failed to patch pod counter, expected 4 but got %s", resp.Annotations[Annotation])
	}
}

func TestShouldIncrementCounterAndReturnValue(t *testing.T) {
	var clientset kubernetes.Interface
	p := getCounterTestPod(true, "5")
	clientset = fake.NewSimpleClientset(&p)

	c := Counter{
		clientset,
	}

	val, err := c.IncrementCounter(&p)
	if err != nil {
		t.Errorf("Failed to increase pod counter to 6 with error %v", err)
	}

	if val != 6 {
		t.Errorf("Failed to increase counter, expected 6 but got %d", val)
	}

	resp, _ := clientset.CoreV1().Pods("test-namespace").Get(context.TODO(), "test-pod", meta_v1.GetOptions{})
	if resp.Annotations[Annotation] != "6" {
		t.Errorf("Failed to validate patched pod, expected 6 but got %s", resp.Annotations[Annotation])
	}
}

func getCounterTestPod(counterEnabled bool, value string) v1.Pod {
	p := v1.Pod{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:        "test-pod",
			Namespace:   "test-namespace",
			UID:         types.UID("uuid"),
			Annotations: map[string]string{},
		},
	}

	if counterEnabled {
		p.Annotations[Annotation] = value
	}

	return p
}
