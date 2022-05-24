/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	policyv1beta1 "k8s.io/api/policy/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"

	"github.com/spf13/cobra"
	"twr.dev/pdbl/pkg/kube"
	"twr.dev/pdbl/pkg/printers"
	"twr.dev/pdbl/pkg/resources"
)

var (
	blockingThreshold int16
	outputFormatTmp   string
)

// lookupCmd represents the lookup command
var lookupCmd = &cobra.Command{
	Use:   "lookup",
	Short: "Lookup the pods associated with a target Pod Disruption Budget (PDB) resource",
	Long:  `Lookup the pods associated with a target Pod Disruption Budget (PDB) resource`,
	Run: func(cmd *cobra.Command, args []string) {

		var (
			pdbList        *policyv1beta1.PodDisruptionBudgetList
			pdbListOptions metav1.ListOptions
			pdbOutput      resources.PDBOutput
			outputFormat   string
		)

		// Setup the Kubernetes Client
		client, err := kube.CreateKubeClient(kubeconfig, configContext, noWarnings)
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

			// Check if No-Blocking Filter is specified. If it is, output all PDB's whether they're blocking or not
			noBlockingFilter, err := cmd.Flags().GetBool("no-blocking")
			if err != nil {
				fmt.Printf("ERROR: Unable to read argument passed to \"no-blocking\" flag: %v", err)
				os.Exit(1)
			}
			if !noBlockingFilter {
				if pdb.Status.DisruptionsAllowed > int32(blockingThreshold) {
					continue
				}
			}

			// Set LabelSelectors for pods to Selectors value from PDB
			pdbSelectors := pdb.Spec.Selector.MatchLabels
			pdblabels := labels.Set(pdbSelectors).String()

			// Get a list of Pods that match the Selectors from the PDB
			pods, err := client.CoreV1().Pods(pdb.Namespace).List(context.TODO(), metav1.ListOptions{LabelSelector: pdblabels})
			if err != nil {
				fmt.Printf("ERROR: Unable to lookup Pods: \n%v\n", err)
				os.Exit(1)
			}

			// Check if any pods match the Selectors from the PDB, skip iteration if not
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

		if outputFormat == "json" {

			output, err := json.MarshalIndent(pdbOutput, "", "    ")
			if err != nil {
				fmt.Printf("ERROR: Problems marshaling output to JSON: %v\n", err)
			}
			fmt.Printf("%s\n", output)
		} else {
			for _, pdb := range pdbOutput.PDBs {
				fmt.Fprintf(w, "%v\t%v\t%v\t%v\t%v\t\n", pdb.Name, pdb.Namespace, len(pdb.Pods), pdb.DisruptionsAllowed, pdb.Selectors)
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
	lookupCmd.Flags().BoolP("no-blocking", "b", false, "Assess all PDB's, not just those that are blocking (Default: False)")
	lookupCmd.Flags().Int16VarP(&blockingThreshold, "blocking-threshold", "t", 0, "Set the threshold for blocking PDB's. This number is the upper bound for \"Allowed Disruptions\" for a PDB (Default: 0)")
	lookupCmd.Flags().Bool("no-headers", false, "Output without column headers (Default: False)")
	lookupCmd.Flags().Bool("show-no-pods", false, "Output PDB's that don't match any pods (Default: False)")
	lookupCmd.Flags().StringVarP(&outputFormatTmp, "output", "o", "", "Specify the output format. One of: json")

}
