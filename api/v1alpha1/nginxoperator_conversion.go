package v1alpha1

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/conversion"

	"github.com/jessequinn/nginx-operator/api/v1alpha2"
)

// ConvertTo converts v1alpha1 to v1alpha2
func (no *NginxOperator) ConvertTo(dst conversion.Hub) error {
	objV1alpha2 := dst.(*v1alpha2.NginxOperator)
	objV1alpha2.ObjectMeta = no.ObjectMeta
	objV1alpha2.Status.Conditions = no.Status.Conditions

	if no.Spec.Replicas != nil {
		objV1alpha2.Spec.Replicas = no.Spec.Replicas
	}

	if len(no.Spec.ForceRedeploy) > 0 {
		objV1alpha2.Spec.ForceRedeploy = no.Spec.ForceRedeploy
	}

	if no.Spec.Port != nil {
		objV1alpha2.Spec.Ports = make([]v1.ContainerPort, 0, 1)
		objV1alpha2.Spec.Ports = append(objV1alpha2.Spec.Ports, v1.ContainerPort{ContainerPort: *no.Spec.Port})
	}

	return nil
}

// ConvertFrom converts v1alpha2 to v1alpha1
func (no *NginxOperator) ConvertFrom(src conversion.Hub) error {
	objV1alpha2 := src.(*v1alpha2.NginxOperator)
	no.ObjectMeta = objV1alpha2.ObjectMeta
	no.Status.Conditions = objV1alpha2.Status.Conditions

	if objV1alpha2.Spec.Replicas != nil {
		no.Spec.Replicas = objV1alpha2.Spec.Replicas
	}

	if len(objV1alpha2.Spec.ForceRedeploy) > 0 {
		no.Spec.ForceRedeploy = objV1alpha2.Spec.ForceRedeploy
	}

	if len(objV1alpha2.Spec.Ports) > 0 {
		no.Spec.Port = pointer.Int32(objV1alpha2.Spec.Ports[0].ContainerPort)
	}

	return nil
}
