/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	policyv1beta1 "k8s.io/api/policy/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"twr.dev/pdbl/pkg/kube"
	"twr.dev/pdbl/pkg/printers"
	"twr.dev/pdbl/pkg/resources"
)

// patchCmd represents the patch command
var patchCmd = &cobra.Command{
	Use:   "patch",
	Short: "Patch target Pod Disruption Budget (PDB) resources to prevent the blockage of cluster maintenance activities",
	Long:  `Patch target Pod Disruption Budget (PDB) resources to prevent the blockage of cluster maintenance activities`,
	Run: func(cmd *cobra.Command, args []string) {
		var (
			pdbList        *policyv1beta1.PodDisruptionBudgetList
			pods           *corev1.PodList
			pdbListOptions metav1.ListOptions
			pdbOutput      resources.PDBOutput
			outputFormat   string
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
		pdbList, err = client.PolicyV1beta1().PodDisruptionBudgets(namespace).List(context.TODO(), pdbListOptions)
		if err != nil {
			fmt.Printf("ERROR: Unable to lookup Pod Disruption Budgets (PDB's): \n%v\n", err)
			os.Exit(1)
		}

		// Set output format
		outputFormat = strings.ToLower(outputFormatTmp)

		// Set show-no-pods flag
		showNoPods, _ := cmd.Flags().GetBool("show-no-pods")

		// Skip printing headers if flag is set
		noHeaders, _ := cmd.Flags().GetBool("no-headers")

		// Start Printing of output with Headers
		w := printers.GetNewTabWriter(os.Stdout)
		defer w.Flush()

		if !noHeaders && outputFormat != "json" {
			fmt.Fprintln(w, "NAME\tNAMESPACE\tMATCHING PODS\tALLOWED DISRUPTIONS\tSELECTORS\t")
		}

		// Loop through PDB's
		for _, pdb := range pdbList.Items {

			var currPDB resources.PDB

			// Check if Blocking Filter is specified. If it is, only output PDB's that meet the Blocking Threshold
			blockingFilter, _ := cmd.Flags().GetBool("blocking")
			if blockingFilter {
				if pdb.Status.DisruptionsAllowed > int32(blockingThreshold) {
					continue
				}
			}

			// Set LabelSelectors for pods to Selectors value from PDB
			pdbSelectors := pdb.Spec.Selector.MatchLabels
			pdblabels := labels.Set(pdbSelectors).String()

			// Get a list of Pods that match the Selectors from the PDB
			pods, err = client.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{LabelSelector: pdblabels})
			if err != nil {
				fmt.Printf("ERROR: Unable to lookup Pods: \n%v\n", err)
				os.Exit(1)
			}

			// Check if any pods matches the Selectors from the PDB, skip iteration if not
			if len(pods.Items) < 1 {
				if !showNoPods {
					continue
				}
			}

			currPDB.Name = pdb.Name
			currPDB.Namespace = pdb.Namespace
			currPDB.Selectors = pdblabels
			currPDB.DisruptionsAllowed = int(pdb.Status.DisruptionsAllowed)

			// Loop through Pod list and print information for output
			for _, pod := range pods.Items {
				currPDB.Pods = append(currPDB.Pods, pod.Name)
			}

			pdbOutput.PDBs = append(pdbOutput.PDBs, currPDB)

		}

		fmt.Println("Patching PDB's now")
	},
}

func init() {
	rootCmd.AddCommand(patchCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// patchCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// patchCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	patchCmd.Flags().BoolP("blocking", "b", false, "Filter for blocking PDB's only (Default: False)")
	patchCmd.Flags().Int16VarP(&blockingThreshold, "blocking-threshold", "t", 0, "Set the threshold for blocking PDB's. This number is the upper bound for \"Allowed Disruptions\" for a PDB (Default: 0)")
	patchCmd.Flags().BoolP("no-op", "", false, "Run command in a no-op mode. Information will be simulated, but not executed (Default: False)")
}
