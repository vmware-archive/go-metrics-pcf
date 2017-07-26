// Copyright (C) 2017-Present Pivotal Software, Inc. All rights reserved.
//
// This program and the accompanying materials are made available under
// the terms of the under the Apache License, Version 2.0 (the "License”);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//
// See the License for the specific language governing permissions and
// limitations under the License.

package pcfmetrics

import (
	"encoding/json"
	"os"
	"fmt"
)

type credentials struct {
	AccessToken string `json:"access_key"`
	Url string `json:"endpoint"`
}

func getInstanceIndex() string {
	return os.Getenv("INSTANCE_INDEX")
}

func getInstanceGuid() string {
	return os.Getenv("INSTANCE_GUID")
}

func getAppGuid() (string, error) {
	var vcapApplication map[string]*json.RawMessage
	err := json.Unmarshal([]byte(os.Getenv("VCAP_APPLICATION")), &vcapApplication)
	if err != nil {
		return "", err
	}

	var appGuid string
	err = json.Unmarshal(*vcapApplication["application_id"], &appGuid)
	if err != nil {
		return "", err
	}

	return appGuid, nil
}

func getCredentials(serviceName string) (serviceCredentials *credentials, err error) {
	var allServices map[string]*json.RawMessage
	err = json.Unmarshal([]byte(os.Getenv("VCAP_SERVICES")), &allServices)
	if err != nil {
		return nil, err
	}

	forwarderService, err := getService(serviceName)
	if err != nil {
		return nil, err
	}

	var serviceValues []map[string]*json.RawMessage
	err = json.Unmarshal(*forwarderService, &serviceValues)
	if err != nil {
		return nil, err
	}

	var creds credentials
	err = json.Unmarshal(*serviceValues[0]["credentials"], &creds)
	return &creds, err
}

func getService(serviceName string) (service *json.RawMessage, err error) {
	var allServices map[string]*json.RawMessage
	err = json.Unmarshal([]byte(os.Getenv("VCAP_SERVICES")), &allServices)
	if err != nil {
		return nil, err
	}

	service, ok := allServices[serviceName]
	if !ok {
		return nil, fmt.Errorf("could not find service with name: %s", serviceName)
	}

	return service, nil
}
