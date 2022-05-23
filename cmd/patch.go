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

	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	policyv1beta1 "k8s.io/api/policy/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/intstr"
	"twr.dev/pdbl/pkg/kube"
	"twr.dev/pdbl/pkg/printers"
	"twr.dev/pdbl/pkg/resources"
)

var (
	PDBLMaxUnavailableAnnotation = "pdbl.k8s.t-mobile.com/maxUnavailable-original"
	PDBLMinAvailableAnnotation   = "pdbl.k8s.t-mobile.com/minAvailable-original"
	noop                         bool
	patchStatus                  string
)

// printPDBAvailValue takes a string formatted value from a PDB's "minAvailable" or "maxUnavailable" field and returns the value or "N/A"
func printPDBAvailValue(value string) string {
	if value == "" {
		return "N/A"
	} else {
		return value
	}
}

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
			dryRunOptions  metav1.UpdateOptions
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
			fmt.Fprintln(w, "NAME\tNAMESPACE\tMAX UNAVAILABLE OLD\tMAX UNAVAILABLE NEW\tMIN AVAILABLE OLD\tMIN AVAILABLE NEW\tALLOWED DISRUPTIONS\tPATCH STATUS\t")
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
			pods, err = client.CoreV1().Pods(pdb.Namespace).List(context.TODO(), metav1.ListOptions{LabelSelector: pdblabels})
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

			/*
				Need to check PDB to see if minAvailable or maxUnavailable is used (they're mutually exclusive)
			*/

			// Assess if this is a no-op run
			if noop {
				dryRunOptions = metav1.UpdateOptions{DryRun: []string{metav1.DryRunAll}}
				patchStatus = "(dry-run only)"
			} else {
				dryRunOptions = metav1.UpdateOptions{}
				patchStatus = ("configured")
			}

			if pdb.Spec.MaxUnavailable != nil {
				currPDB.OldMaxUnavailable = pdb.Spec.MaxUnavailable.StrVal
				pdb.ObjectMeta.Annotations[PDBLMaxUnavailableAnnotation] = currPDB.OldMaxUnavailable
				// Detect an appropriate non-blocking value and update the value
				value := intstr.FromString("%90")
				currPDB.NewMaxUnavailable = value.StrVal
				pdb.Spec.MaxUnavailable = &value

				newPdb, err := client.PolicyV1beta1().PodDisruptionBudgets(pdb.Namespace).Update(context.TODO(), &pdb, dryRunOptions)
				if err != nil {
					fmt.Printf("ERROR: Updating PDB \"%v/%v\" failed: %v\n", pdb.Namespace, pdb.Name, err)
				}

				pdb = *newPdb

			} else if pdb.Spec.MinAvailable != nil {
				currPDB.OldMinAvailable = pdb.Spec.MinAvailable.StrVal
				pdb.ObjectMeta.Annotations[PDBLMinAvailableAnnotation] = currPDB.OldMinAvailable
				// Detect an appropriate non-blocking value and update the value
				value := intstr.FromString("1%")
				currPDB.NewMinAvailable = value.StrVal
				pdb.Spec.MinAvailable = &value

				newPdb, err := client.PolicyV1beta1().PodDisruptionBudgets(pdb.Namespace).Update(context.TODO(), &pdb, dryRunOptions)
				if err != nil {
					fmt.Printf("ERROR: Updating PDB \"%v/%v\" failed: %v\n", pdb.Namespace, pdb.Name, err)
				}

				pdb = *newPdb
			}

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
				fmt.Fprintf(w, "%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\t\n", pdb.Name, pdb.Namespace, printPDBAvailValue(pdb.OldMaxUnavailable), printPDBAvailValue(pdb.NewMaxUnavailable), printPDBAvailValue(pdb.OldMinAvailable), printPDBAvailValue(pdb.NewMinAvailable), pdb.DisruptionsAllowed, patchStatus)
			}
		}

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
	patchCmd.Flags().BoolP("no-blocking", "b", false, "Assess all PDB's, not just those that are blocking (Default: False)")
	patchCmd.Flags().Int16VarP(&blockingThreshold, "blocking-threshold", "t", 0, "Set the threshold for blocking PDB's. This number is the upper bound for \"Allowed Disruptions\" for a PDB (Default: 0)")
	patchCmd.Flags().Bool("no-headers", false, "Output without column headers (Default: False)")
	patchCmd.Flags().Bool("show-no-pods", false, "Output PDB's that don't match any pods (Default: False)")
	patchCmd.Flags().StringVarP(&outputFormatTmp, "output", "o", "", "Specify the output format. One of: json")
	patchCmd.Flags().BoolVarP(&noop, "no-op", "", false, "Run command in a no-op mode. Information will be simulated, but not executed (Default: False)")
}
