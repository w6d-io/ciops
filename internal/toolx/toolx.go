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
