package helpers

import (
	"fmt"
	"strconv"
	"time"
)

// GetTargetPDBValue returns an appropriate value to set a PDB's maxUnavailable/MinAvailable field to be non-blocking for node operations
func GetTargetPDBValue(allowedDisruptions int, availabilityType string, availabilityValue string, numMatchedPods int) {

	lengthOfAvailabilityValue := len(availabilityValue)

	fmt.Printf("Availability Value: %v\n", availabilityValue)

	// Handle situations where there's only 1 replica
	if numMatchedPods == 1 {
		fmt.Printf("getTargetPDBValue - Only matches one pod: %v\n", numMatchedPods)
	}

	// Handle situations where there's a percentage
	if lengthOfAvailabilityValue > 1 && string(availabilityValue[len(availabilityValue)-1]) == "%" {
		fmt.Println("getTargetPDBValue - Percent Value detected")

		// Handle situations where there's an integer
	} else {

		fmt.Println("getTargetPDBValue - Integer Value detected")
	}

}

/*
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
