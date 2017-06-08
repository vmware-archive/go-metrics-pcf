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
	"errors"
	"os"
)

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

func getToken(serviceName string) (string, error) {
	token, _, err := getCredentials(serviceName)
	return token, err
}

func getUrl(serviceName string) (string, error) {
	_, url, err := getCredentials(serviceName)
	return url, err
}

func getCredentials(serviceName string) (accessToken, url string, err error) {
	var allServices map[string]*json.RawMessage
	err = json.Unmarshal([]byte(os.Getenv("VCAP_SERVICES")), &allServices)
	if err != nil {
		return "", "", err
	}

	for k, v := range allServices {
		if k == serviceName {
			var serviceValues []map[string]*json.RawMessage
			err = json.Unmarshal(*v, &serviceValues)
			if err != nil {
				return "", "", err
			}

			var creds map[string]string
			err = json.Unmarshal(*serviceValues[0]["credentials"], &creds)
			if err != nil {
				return "", "", err

			}

			return creds["access_key"], creds["hostname"], nil
		}
	}

	return "", "", errors.New("custom metrics service not found")
}
