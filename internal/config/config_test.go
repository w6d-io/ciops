/*
Copyright 2020 WILDCARD SA.

Licensed under the WILDCARD SA License, Version 1.0 (the "License");
WILDCARD SA is register in french corporation.
You may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.w6d.io/licenses/LICENSE-1.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is prohibited.
Created on 21/09/2022
*/

package config_test

import (
	. "github.com/onsi/ginkgo"
	"github.com/spf13/viper"

	"github.com/w6d-io/ciops/internal/config"
)

var _ = Describe("Config", func() {
	Context("Manage issues", func() {
		BeforeEach(func() {
			config.SkipValidation = true
			viper.Reset()
		})
		It("File does not exist", func() {
			config.CfgFile = "testdata/no-file.yaml"
			config.Init()
		})
		It("File does not exist", func() {
			config.CfgFile = "testdata/bad-content.yaml"
			config.Init()
		})
	})
	Context("Validate config", func() {
		BeforeEach(func() {
			config.SkipValidation = true
			viper.Reset()
		})
		It("File does not exist", func() {
			config.CfgFile = "testdata/file1.yaml"
			config.Init()
		})

	})
})
