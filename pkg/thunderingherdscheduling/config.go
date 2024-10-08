package thunderingherdscheduling

import (
	"encoding/json"
	"fmt"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
)

func ParseArguments(obj runtime.Object) (*ThunderingHerdSchedulingArgs, error) {
	conf := &ThunderingHerdSchedulingArgs{}
	if obj != nil {
		// workaround for wrong type conversion
		cfg, ok := obj.(*runtime.Unknown)
		if !ok {
			return nil, fmt.Errorf("conversion error, got %T", obj)
		}
		err := json.Unmarshal(cfg.Raw, conf)
		if err != nil {
			return nil, err
		}
	}

	//SetDefaultThunderingHerdArgs(conf)
	return conf, nil
}

func SetDefaultThunderingHerdArgs(args *ThunderingHerdSchedulingArgs) {

	if args.ParallelStartingPodsPerNode == nil {
		defaultParallelPods := 3
		args.ParallelStartingPodsPerNode = &defaultParallelPods
	}

	if args.TimeoutSeconds == nil {
		defaultTimeoutSeconds := 5
		args.TimeoutSeconds = &defaultTimeoutSeconds
	}

	if args.MaxRetries == nil {
		defaultRetries := 5
		args.MaxRetries = &defaultRetries
	}
}

type ThunderingHerdSchedulingArgs struct {
	meta_v1.TypeMeta

	ParallelStartingPodsPerNode *int `json:"parallelStartingPodsPerNode"`
	TimeoutSeconds              *int `json:"timeoutSeconds"`
	MaxRetries                  *int `json:"maxRetries"`
}

func (in *ThunderingHerdSchedulingArgs) PrintArgs() {
	klog.Info("Configuration")
	klog.Infof("ParallelStartingPodsPerNode=%d", *in.ParallelStartingPodsPerNode)
	klog.Infof("TimeoutSeconds=%d", *in.TimeoutSeconds)
	klog.Infof("MaxRetries=%d", *in.MaxRetries)
}

func (in *ThunderingHerdSchedulingArgs) DeepCopy() *ThunderingHerdSchedulingArgs {
	if in == nil {
		return nil
	}
	out := new(ThunderingHerdSchedulingArgs)
	in.DeepCopyInto(out)
	return out
}

func (in *ThunderingHerdSchedulingArgs) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}

	return nil
}

func (in *ThunderingHerdSchedulingArgs) DeepCopyInto(out *ThunderingHerdSchedulingArgs) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	out.MaxRetries = in.MaxRetries
	out.TimeoutSeconds = in.TimeoutSeconds
	out.ParallelStartingPodsPerNode = in.ParallelStartingPodsPerNode
	return
}
