/*
Copyright © 2022 TWR Engineering

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package kube

import (
	"context"
	"fmt"

	policyv1 "k8s.io/api/policy/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	// Import all auth client plugins
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/tools/clientcmd"
)

// CreateKubeClient creates a new kubernetes client interface
func CreateKubeClient(kubeconfig string, configContext string) (kubernetes.Interface, error) {

	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	loadingRules.DefaultClientConfig = &clientcmd.DefaultClientConfig
	loadingRules.ExplicitPath = kubeconfig
	configOverrides := &clientcmd.ConfigOverrides{ClusterDefaults: clientcmd.ClusterDefaults, CurrentContext: configContext}

	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
	config, err := kubeConfig.ClientConfig()
	clientset, _ := kubernetes.NewForConfig(config)

	return clientset, err

}

// GetPDBs gets a list of Kubernetes PDB's
func GetAllPDBs(client kubernetes.Interface, namespace string) (*policyv1.PodDisruptionBudgetList, error) {

	var pdbs *policyv1.PodDisruptionBudgetList

	pdbs, err := client.PolicyV1().PodDisruptionBudgets(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		fmt.Println(err)

		return pdbs, err
	}

	return pdbs, nil
}
