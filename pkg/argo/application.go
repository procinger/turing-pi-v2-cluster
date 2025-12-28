package argo

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Application struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata" protobuf:"bytes,1,opt,name=metadata"`
	Spec              ApplicationSpec `json:"spec" protobuf:"bytes,2,opt,name=spec"`
}

type ApplicationSpec struct {
	Source      *ApplicationSource     `json:"source,omitempty" protobuf:"bytes,1,opt,name=source"`
	Destination ApplicationDestination `json:"destination" protobuf:"bytes,2,name=destination"`
	Sources     ApplicationSources     `json:"sources,omitempty" protobuf:"bytes,8,opt,name=sources"`
}

type ApplicationSource struct {
	RepoURL        string                 `json:"repoURL" protobuf:"bytes,1,opt,name=repoURL"`
	Path           string                 `json:"path,omitempty" protobuf:"bytes,2,opt,name=path"`
	TargetRevision string                 `json:"targetRevision,omitempty" protobuf:"bytes,4,opt,name=targetRevision"`
	Helm           *ApplicationSourceHelm `json:"helm,omitempty" protobuf:"bytes,7,opt,name=helm"`
	Chart          string                 `json:"chart,omitempty" protobuf:"bytes,12,opt,name=chart"`
}

type ApplicationSourceHelm struct {
	Values string `json:"values,omitempty" patchStrategy:"replace" protobuf:"bytes,4,opt,name=values"`
}

type ApplicationDestination struct {
	Server    string `json:"server,omitempty" protobuf:"bytes,1,opt,name=server"`
	Namespace string `json:"namespace,omitempty" protobuf:"bytes,2,opt,name=namespace"`
	Name      string `json:"name,omitempty" protobuf:"bytes,3,opt,name=name"`
}

type ApplicationSources []ApplicationSource
