package thunderingherdscheduling

import (
	"k8s.io/apimachinery/pkg/conversion"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	schedschemev1beta2 "k8s.io/kube-scheduler/config/v1beta2"
)

const GroupName = "kubescheduler.config.k8s.io"

var SchemaVersionV1Beta2 = schema.GroupVersion{Group: GroupName, Version: "v1beta2"}
var SchemeGroupVersionInternal = schema.GroupVersion{Group: GroupName, Version: runtime.APIVersionInternal}

var (
	localSchemeBuilder = &schedschemev1beta2.SchemeBuilder
)

func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(SchemaVersionV1Beta2, &ThunderingHerdSchedulingArgs{})
	scheme.AddKnownTypes(SchemeGroupVersionInternal, &ThunderingHerdSchedulingArgs{})
	scheme.AddTypeDefaultingFunc(&ThunderingHerdSchedulingArgs{}, func(obj interface{}) {
		SetDefaultThunderingHerdArgs(obj.(*ThunderingHerdSchedulingArgs))
	})
	err := scheme.AddGeneratedConversionFunc((*ThunderingHerdSchedulingArgs)(nil), (*ThunderingHerdSchedulingArgs)(nil), func(a, b interface{}, scope conversion.Scope) error {
		a = b
		return nil
	})
	return err
}

func init() {
	localSchemeBuilder.Register(addKnownTypes)
}
