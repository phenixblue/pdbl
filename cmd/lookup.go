/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"context"
	"fmt"
	"os"

	v1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"

	"github.com/spf13/cobra"
	"twr.dev/pdbl/pkg/kube"
	"twr.dev/pdbl/pkg/printers"
)

// lookupCmd represents the lookup command
var lookupCmd = &cobra.Command{
	Use:   "lookup",
	Short: "Lookup the pods assocaited with a target Pod Disruption Bidget (PDB) resource",
	Long:  `Lookup the pods assocaited with a target Pod Disruption Bidget (PDB) resource`,
	Run: func(cmd *cobra.Command, args []string) {

		var (
			//targetPDB string
			pdbList        *policyv1.PodDisruptionBudgetList
			pods           *v1.PodList
			pdbListOptions metav1.ListOptions
		)

		// Setup the Kubernetes Client
		client, err := kube.CreateKubeClient(kubeconfig, configContext)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Check Command Args passed and update the PDB ListOptions FieldSelector to scope resources listed
		cmdArgs := args
		if len(cmdArgs) > 0 && cmdArgs[0] != "" {
			targetPDB := cmdArgs[0]

			pdbListOptions.FieldSelector = "metadata.name=" + targetPDB

		}

		// Get a list of PDB resources
		pdbList, err = client.PolicyV1().PodDisruptionBudgets(namespace).List(context.TODO(), pdbListOptions)
		if err != nil {
			fmt.Println("ERROR: Unable to lookup Pod Disruption Budget (PDB) %q")
			os.Exit(1)
		}

		// Start Printing of output with Headers
		w := printers.GetNewTabWriter(os.Stdout)
		defer w.Flush()
		fmt.Fprintln(w, "NAMESPACE\tNAME\tPDB\tALLOWED DISRUPTIONS\tSELECTORS\t")

		for _, pdb := range pdbList.Items {

			// Set LabelSelectors for pods to Selectors value from PDB
			pdbSelectors := pdb.Spec.Selector.MatchLabels
			labels := labels.Set(pdbSelectors).String()

			// Get a list of Pods that match the Selectors from the PDB
			pods, err = client.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{LabelSelector: labels})

			// Check if any pods matches the Selectors from the PDB, skip iteration if not
			if len(pods.Items) < 1 {
				continue
			}

			// Loop through Pod list and print information for output
			for _, pod := range pods.Items {
				fmt.Fprintf(w, "%v\t%v\t%v\t%v\t%v\t\n", pod.Namespace, pod.Name, pdb.Name, pdb.Status.DisruptionsAllowed, labels)
			}

		}

	},
}

func init() {
	rootCmd.AddCommand(lookupCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// lookupCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// lookupCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
