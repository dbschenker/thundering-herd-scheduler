package thunderingherdscheduling

import (
	"context"
	"github.com/dbschenker/thundering-herd-scheduler/pkg/nodestate"
	"github.com/dbschenker/thundering-herd-scheduler/pkg/podcounter"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/scheduler/framework"
	"math"
	"sync"
	"time"
)

const (
	Name = "ThunderingHerdScheduling"
)

type ThunderingHerdScheduling struct {
	counter   podcounter.PodCounterInterface
	nodestate nodestate.NodeStateInterface
	args      *ThunderingHerdSchedulingArgs
	mutex     *sync.Mutex
}

var _ framework.PermitPlugin = &ThunderingHerdScheduling{}

func (t *ThunderingHerdScheduling) Permit(_ context.Context, _ *framework.CycleState, p *v1.Pod, nodeName string) (*framework.Status, time.Duration) {
	t.mutex.Lock()

	status, duration := t.PermitInternal(p, nodeName)
	if status.Code() == framework.Success {
		t.nodestate.AddSchedulingPod(p, nodeName)
	}

	t.mutex.Unlock()
	return status, duration
}

func (t *ThunderingHerdScheduling) PermitInternal(p *v1.Pod, nodeName string) (*framework.Status, time.Duration) {
	notReadyPods := t.nodestate.NotReadyPods(nodeName)

	maxAllowedStartingPods, err := t.nodestate.NotReadyPodsAllowedInParallel(t.args.ParallelStartingPodsPerNode, t.args.ParallelStartingPodsPerCore, nodeName)

	klog.Infof("Node %s is allowed to start %d pods in parallel", nodeName, maxAllowedStartingPods)

	if err != nil {
		return framework.NewStatus(framework.Error, err.Error()), 0
	}

	if notReadyPods >= maxAllowedStartingPods {
		counter, err := t.counter.IncrementCounter(p)
		if err != nil {
			// to prevent any kind of issue with the scheduler
			klog.ErrorS(err, "Failed to increase counter with error", "pod", klog.KObj(p))
			return framework.NewStatus(framework.Success), 0
		}

		if counter > *t.args.MaxRetries {
			klog.Warning("Pod had to wait for > max retries, scheduling it", "pod", klog.KObj(p))
			return framework.NewStatus(framework.Success), 0
		}

		// we need to wait
		timeoutSeconds := t.args.TimeoutSeconds
		waitTime := powInt(*timeoutSeconds, 2) * counter

		klog.Info("Pod has to wait as there are already more pods not ready then allowed to start parallel on node",
			"pod", klog.KObj(p),
			"maxAllowedStartingPods", maxAllowedStartingPods,
			"notReadyPods", notReadyPods,
			"nodeName", nodeName,
			"waitTime", waitTime)

		return framework.NewStatus(framework.Wait), time.Duration(waitTime) * time.Second
	} else {
		return framework.NewStatus(framework.Success), 0
	}
}

func (t *ThunderingHerdScheduling) Name() string {
	return Name
}

func New(_ context.Context, obj runtime.Object, handle framework.Handle) (framework.Plugin, error) {
	args, err := ParseArguments(obj)
	if err != nil {
		return nil, err
	}

	var m sync.Mutex
	c := &ThunderingHerdScheduling{
		counter:   podcounter.New(handle.ClientSet()),
		args:      args,
		nodestate: nodestate.NewNodeStateV2(handle.ClientSet()),
		mutex:     &m,
	}

	klog.Info("Registering Thundering Herd Scheduler")
	args.PrintArgs()

	return c, nil
}

func powInt(x, y int) int {
	return int(math.Pow(float64(x), float64(y)))
}
