/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"twr.dev/pdbl/pkg/kube"
	"twr.dev/pdbl/pkg/printers"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List the Pod Disruption Budget (PDB) resources",
	Long:  `List the Pod Disruption Budget (PDB) resources`,
	Run: func(cmd *cobra.Command, args []string) {

		client, err := kube.CreateKubeClient(kubeconfig, configContext, noWarnings)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		pdbList, err := kube.GetAllPDBs(client, namespace)
		if err != nil {
			fmt.Println("No PDB's were found")
			os.Exit(1)
		}

		w := printers.GetNewTabWriter(os.Stdout)

		defer w.Flush()

		fmt.Fprintln(w, "NAMESPACE\tNAME\tMIN AVAILABLE\tMAX UNAVAILABLE\tALLOWED DISRUPTIONS\tAGE\t")

		for _, pdb := range pdbList.Items {

			var (
				pdbMaxUnavailable string
				pdbMinAvailable   string
				pdbAge            string
			)

			// Check for "nil" returned value
			if pdb.Spec.MaxUnavailable == nil {
				pdbMaxUnavailable = "N/A"
			} else {
				pdbMaxUnavailable = pdb.Spec.MaxUnavailable.String()
			}

			// Check for "nil" returned value
			if pdb.Spec.MinAvailable == nil {
				pdbMinAvailable = "N/A"
			} else {
				pdbMinAvailable = pdb.Spec.MinAvailable.String()
			}

			currTime := time.Now()
			pdbAge = kube.GetPDBAge(currTime, pdb.GetCreationTimestamp().Time)

			fmt.Fprintf(w, "%v\t%v\t%v\t%v\t%v\t%v\t\n", pdb.Namespace, pdb.Name, pdbMinAvailable, pdbMaxUnavailable, pdb.Status.DisruptionsAllowed, pdbAge)

		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// listCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// listCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
