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

package v1alpha1

import "fmt"

type Action struct {
    ID           string                 `json:"id,omitempty"`
    Name         string                 `json:"name,omitempty"`
    ComponentID  string                 `json:"componentId,omitempty"`
    Ref          string                 `json:"ref,omitempty"`
    Data         map[string]string      `json:"data,omitempty"`
    Params       map[string]interface{} `json:"params,omitempty"`
    Environments map[string]string      `json:"environments,omitempty"`
    Status       string                 `json:"status,omitempty"`
    StartTime    int64                  `json:"startTime,omitempty"`
    EndTime      int64                  `json:"endTime,omitempty"`
}

type ConditionType string

type Condition struct {
    Id   string        `json:"id,omitempty"`
    Ref  string        `json:"ref,omitempty"`
    Type ConditionType `json:"type,omitempty"`
    When string        `json:"when,omitempty"`
}

type Conditions [][]Condition

type Task struct {
    ID            string     `json:"id,omitempty"`
    Name          string     `json:"name,omitempty"`
    SkipOnFailure bool       `json:"skipOnFailure,omitempty"`
    Conditions    Conditions `json:"conditions,omitempty"`
    Actions       []Action   `json:"actions,omitempty"`
    StartTime     int64      `json:"startTime,omitempty"`
    EndTime       int64      `json:"endTime,omitempty"`
    Status        string     `json:"status,omitempty"`
}

type Stage struct {
    ID        string `json:"id,omitempty"`
    Name      string `json:"name,omitempty"`
    Tasks     []Task `json:"tasks,omitempty"`
    Status    string `json:"status,omitempty"`
    EndTime   int64  `json:"endTime,omitempty"`
    StartTime int64  `json:"startTime,omitempty"`
}

type ProjectID int64

type Pipeline struct {
    ID               string    `json:"id,omitempty"`
    Type             string    `json:"type,omitempty"`
    PipelineIDNumber string    `json:"pipelineIdNumber,omitempty"`
    ProjectID        ProjectID `json:"projectId,omitempty"`
    Name             string    `json:"name,omitempty"`
    Stages           []Stage   `json:"stages,omitempty"`
    Status           string    `json:"status,omitempty"`
    StartTime        int64     `json:"startTime,omitempty"`
    EndTime          int64     `json:"endTime,omitempty"`
    LogUri           string    `json:"logUri,omitempty"`
    Complete         bool      `json:"complete,omitempty"`
    Force            bool      `json:"force,omitempty"`
    Artifacts        bool      `json:"artifacts,omitempty"`
    TriggerId        string    `json:"triggerId,omitempty"`
    Commit           Commit    `json:"commit,omitempty"`
    EventID          string    `json:"eventId,omitempty"`
}

type Commit struct {
    ID      string `json:"id,omitempty"`
    Ref     string `json:"ref,omitempty"`
    Message string `json:"message,omitempty"`
}

func (in ProjectID) String() string {
    return fmt.Sprintf("%d", in)
}
