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
func GetTargetPDBValue(forced bool, availabilityType string, pdb resources.PDB) {

	var (
		output             intstr.IntOrString
		availabilityValue  string
		numMatchedPods     int
		forcedPercentValue intstr.IntOrString
		forcedIntValue     intstr.IntOrString
	)

	numMatchedPods = len(pdb.Pods)

	if availabilityType == "minAvailable" {
		availabilityValue = pdb.OldMinAvailable
		forcedIntValue = intstr.FromInt(1)
		forcedPercentValue = intstr.FromString("0%")
	} else if availabilityType == "maxUnavailable" {
		availabilityValue = pdb.OldMaxUnavailable
		forcedIntValue = intstr.FromInt(numMatchedPods)
		forcedPercentValue = intstr.FromString("100%")
	} else {
		fmt.Printf("ERROR: Invalid availability type specified: %v\n", availabilityType)
		os.Exit(1)
	}

	lengthOfAvailabilityValue := len(availabilityValue)

	// Handle situations where there's only 1 replica
	if numMatchedPods == 1 {
		output = intstr.FromString("100%")
		fmt.Printf("getTargetPDBValue - Only matches one pod: %v; New Value: %v\n", numMatchedPods, output.StrVal)
	}

	// Handle situations where there's a percentage
	if lengthOfAvailabilityValue > 1 && string(availabilityValue[len(availabilityValue)-1]) == "%" {

		//newValue := (1 / 100) * numMatchedPods

		if forced {
			output = forcedPercentValue
		}

		// Set to new target value
		//TODO: Need to detect availability type in order to calculate appropriate value
		//output = intstr.FromString(strconv.Itoa(newValue) + "%")
		//fmt.Println("getTargetPDBValue - Percent Value detected")

		// Handle situations where there's an integer
	} else {

		if forced {
			output = forcedIntValue
		}

		// Set to new target value
		//TODO: Need to detect availability type in order to calculate appropriate value
		//output = intstr.FromString(strconv.Itoa(newValue))

		//fmt.Println("getTargetPDBValue - Integer Value detected")
	}

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
