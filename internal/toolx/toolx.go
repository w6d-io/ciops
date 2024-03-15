/*
Copyright 2020 WILDCARD SA.

Licensed under the WILDCARD SA License, Version 1.0 (the "License");
WILDCARD SA is register in french corporation.
You may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.w6d.io/licenses/LICENSE-1.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is prohibited.
Created on 26/01/2024
*/

package toolx

import (
	"bytes"
	pipelinev1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/client-go/kubernetes/scheme"
	"runtime/debug"

	"github.com/go-logr/logr"

	"github.com/w6d-io/x/toolx"
)

func ShowVersion(log logr.Logger, version string) {
	var info []interface{}
	info = append(info, "version", version)
	bi, _ := debug.ReadBuildInfo()
	for _, s := range bi.Settings {
		if toolx.InArray(s.Key, []string{"vcs.time", "vcs.revision"}) {
			info = append(info, s.Key[4:], s.Value)
		}
	}
	log.Info("start service", info...)
}

func DeDuplicateParams(src []pipelinev1.ParamSpec) (n []pipelinev1.ParamSpec) {
	for _, i := range src {
		if !IsContainParams(n, i) {
			n = append(n, i)
		}
	}
	return
}

func IsContainParams(t []pipelinev1.ParamSpec, p pipelinev1.ParamSpec) bool {
	for _, c := range t {
		if c.Name == p.Name {
			return true
		}
	}
	return false
}

// GetObjectContain ...
func GetObjectContain(obj runtime.Object) string {
	s := json.NewSerializerWithOptions(json.DefaultMetaFactory, scheme.Scheme, scheme.Scheme, json.SerializerOptions{Yaml: true})
	buf := new(bytes.Buffer)
	if err := s.Encode(obj, buf); err != nil {
		return "<ERROR>\n"
	}
	return buf.String()
}
