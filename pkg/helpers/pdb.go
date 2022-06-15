package helpers

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	policyv1beta1 "k8s.io/api/policy/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"twr.dev/pdbl/pkg/resources"
)

// GetTargetPDBValue returns an appropriate value to set a PDB's maxUnavailable/MinAvailable field to be non-blocking for node operations
func GetTargetPDBValue(forced bool, availabilityType string, pdb resources.PDB) intstr.IntOrString {

	var (
		output             intstr.IntOrString
		numMatchedPods     int
		forcedPercentValue intstr.IntOrString
		forcedIntValue     intstr.IntOrString
	)

	numMatchedPods = len(pdb.Pods)

	switch availabilityType {
	case "minAvailable":
		forcedIntValue = intstr.FromInt(1)
		forcedPercentValue = intstr.FromString("0%")
	case "maxUnavailable":
		forcedIntValue = intstr.FromInt(numMatchedPods)
		forcedPercentValue = intstr.FromString("100%")
	default:
		fmt.Printf("ERROR: Invalid availability type specified: %v\n", availabilityType)
		os.Exit(1)
	}

	// Handle situations where there's only 1 replica
	if numMatchedPods == 1 {
		fmt.Println("Only patches 1 available pod")
		output = intstr.FromString("100%")
	} else if isPercent(availabilityType, pdb) {
		fmt.Println("Is a percent")
		output = forcedPercentValue
	} else {
		fmt.Println("Is an int")
		output = forcedIntValue
	}

	return output

	/*

		TODO: Need to finish this up


		if it matches only on one pod; then

			// Need to also make the distinction when the scenario is "good pdb + currently unhealthy pods" and disruptions aren't allowed

			if it's a percentage && type == "maxUnavailable"; then

				targetValue := "100%"

			if it's a percentage && type == "minAvailable"; then

				targetValue == "0%"

			if it's an integer && type == "maxUnavailable"; then

				targetValue == "1"

			if it's an integer && type == "minAvailable"; then

				targetValue == "0"
	*/

}

func GetSimplePDB(client kubernetes.Interface, pdb policyv1beta1.PodDisruptionBudget, showNoPods bool) resources.PDB {

	var currPDB resources.PDB

	// Set LabelSelectors for pods to Selectors value from PDB
	pdbSelectors := pdb.Spec.Selector.MatchLabels
	pdblabels := labels.Set(pdbSelectors).String()
	pdbAnnotations := pdb.Annotations

	currPDB.Name = pdb.Name
	currPDB.Namespace = pdb.Namespace
	currPDB.Selectors = pdblabels
	currPDB.DisruptionsAllowed = int(pdb.Status.DisruptionsAllowed)

	// Check if a patch annotation exists
	for annotationKey, _ := range pdbAnnotations {
		if strings.HasPrefix(annotationKey, "pdbl") {
			currPDB.PatchStatus = true
		}
	}

	// Get a list of Pods that match the Selectors from the PDB
	pods, err := client.CoreV1().Pods(pdb.Namespace).List(context.TODO(), metav1.ListOptions{LabelSelector: pdblabels})
	if err != nil {
		fmt.Printf("ERROR: Unable to lookup Pods: \n%v\n", err)
		os.Exit(1)
	}

	// Loop through Pod list and print information for output
	for _, pod := range pods.Items {
		currPDB.Pods = append(currPDB.Pods, pod.Name)
	}

	return currPDB
}

// GetPDBAge gets the age of a Pod Disruption Budget in a friendly format similar to "kubectl"
func GetPDBAge(currTime time.Time, pdbTime time.Time) string {

	var output string
	output = "Error"

	numDays := currTime.Sub(pdbTime).Hours() / 24
	numHours := currTime.Sub(pdbTime).Hours()
	numMins := currTime.Sub(pdbTime).Minutes()

	if numDays > 1 {
		output = strconv.Itoa(int(currTime.Sub(pdbTime).Hours()/24)) + "d"
	} else if numHours > 2 {
		output = strconv.Itoa(int(currTime.Sub(pdbTime).Hours())) + "h"
	} else if numMins > 2 && numMins < 10 {
		output = strconv.Itoa(int(currTime.Sub(pdbTime).Minutes())) + "m" + strconv.Itoa(int(currTime.Sub(pdbTime).Seconds())%60) + "s"
	} else if numMins > 2 && numMins >= 10 {
		output = strconv.Itoa(int(currTime.Sub(pdbTime).Minutes())) + "m"
	} else {
		output = strconv.Itoa(int(currTime.Sub(pdbTime).Seconds())) + "s"
	}

	return output
}

// isPercent checks if the availability value of a PDB is a percentage or not
func isPercent(availabilityType string, pdb resources.PDB) bool {

	var availabilityValue string
	var lengthOfAvailabilityValue int

	output := false

	switch availabilityType {
	case "minAvailable":
		availabilityValue = pdb.OldMinAvailable
		lengthOfAvailabilityValue = len(availabilityValue)
	case "maxUnavailable":
		availabilityValue = pdb.OldMaxUnavailable
		lengthOfAvailabilityValue = len(availabilityValue)
	}

	fmt.Printf("Availability Value: %v\n", availabilityValue)
	fmt.Printf("Last character: %v\n", string(availabilityValue[lengthOfAvailabilityValue-1]))
	if lengthOfAvailabilityValue > 1 && string(availabilityValue[lengthOfAvailabilityValue-1]) == "%" {
		fmt.Println("Matched percent logic")
		output = true
	}

	return output

}
