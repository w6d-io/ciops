/*
Copyright 2020 WILDCARD SA.

Licensed under the WILDCARD SA License, Version 1.0 (the "License");
WILDCARD SA is register in french corporation.
You may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.w6d.io/licenses/LICENSE-1.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is prohibited.
Created on 23/09/2022
*/

package pipelineruns

import (
    "bytes"
    "fmt"
    "gitlab.w6d.io/w6d/ciops/internal/config"
    "k8s.io/apimachinery/pkg/runtime"
    "k8s.io/apimachinery/pkg/runtime/serializer/json"
    "k8s.io/client-go/kubernetes/scheme"
)

func getPipelinerunName(id int64) string {
    return fmt.Sprintf("%s-%d", config.GetPipelinerunPrefix(), id)
}

func getObjectContain(obj runtime.Object) string {
    s := json.NewSerializerWithOptions(json.DefaultMetaFactory, scheme.Scheme, scheme.Scheme, json.SerializerOptions{Yaml: true})
    buf := new(bytes.Buffer)
    if err := s.Encode(obj, buf); err != nil {
        return "<ERROR>\n"
    }
    return buf.String()
}
