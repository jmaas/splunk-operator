// Copyright (c) 2018-2020 Splunk Inc. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// 	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package controller

import (
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/source"

	splctrl "github.com/splunk/splunk-operator/pkg/splunk/controller"
)

// SplunkControllersToAdd is a list of Splunk controllers to add to the Manager
var SplunkControllersToAdd []splctrl.SplunkController

// AddToManager adds all Splunk Controllers to the Manager
func AddToManager(mgr manager.Manager) error {
	// Use a new go client to work-around issues with the operator sdk design.
	// If WATCH_NAMESPACE is empty for monitoring cluster-wide custom Splunk resources,
	// the default caching client will attempt to list all resources in all namespaces for
	// any get requests, even if the request is namespace-scoped.
	c, err := client.New(mgr.GetConfig(), client.Options{})
	if err != nil {
		return err
	}

	// call AddToManager for each of the registered controllers
	for _, ctrl := range SplunkControllersToAdd {
		if err = addToManager(mgr, ctrl, c); err != nil {
			return err
		}
	}

	return nil
}

// addToManager adds a specific Splunk Controller to the Manager.
// The Manager will set fields on the Controller and Start it when the Manager is Started.
func addToManager(mgr manager.Manager, sctl splctrl.SplunkController, c client.Client) error {

	// Create a new controller
	ctrl, err := splctrl.NewController(mgr, sctl, c)
	if err != nil {
		return err
	}

	// Watch for changes to primary custom resource
	err = ctrl.Watch(&source.Kind{Type: sctl.GetInstance()}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resources
	for _, t := range sctl.GetWatchTypes() {
		err = ctrl.Watch(&source.Kind{Type: t}, &handler.EnqueueRequestForOwner{
			IsController: true,
			OwnerType:    sctl.GetInstance(),
		})
		if err != nil {
			return err
		}
	}

	return err
}
