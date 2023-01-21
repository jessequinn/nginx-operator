package assets

import (
	"embed"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

var (
	//go:embed manifests/*
	manifests embed.FS

	scheme = runtime.NewScheme()
	codec  = serializer.NewCodecFactory(scheme)
)

func init() {
	if err := appsv1.AddToScheme(scheme); err != nil {
		panic(err)
	}

	if err := v1.AddToScheme(scheme); err != nil {
		panic(err)
	}
}

// GetDeploymentFromFile converts Yaml to deployment object
func GetDeploymentFromFile(name string) (*appsv1.Deployment, error) {
	deploymentBytes, err := manifests.ReadFile(name)
	if err != nil {
		return nil, err
	}

	deploymentObject, err := runtime.Decode(codec.UniversalDecoder(appsv1.SchemeGroupVersion), deploymentBytes)
	if err != nil {
		return nil, err
	}

	return deploymentObject.(*appsv1.Deployment), nil
}

// GetServiceFromFile converts Yaml to service object
func GetServiceFromFile(name string) (*v1.Service, error) {
	serviceBytes, err := manifests.ReadFile(name)
	if err != nil {
		return nil, err
	}

	serviceObject, err := runtime.Decode(codec.UniversalDecoder(v1.SchemeGroupVersion), serviceBytes)
	if err != nil {
		return nil, err
	}

	return serviceObject.(*v1.Service), nil
}
