package helpers

import (
	"fmt"
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

	// Need to also make the distinction when the scnerio is "good pdb + currently unhealthy pods" and disruptions aren't allowed

	if it's a percentage && type == "maxUnavailable"; then

		targetValue := "100%"

	if it's a percentage && type == "minAvailable"; then

		targetValue == "0%"

	if it's an integer && type == "maxUnavailable"; then

		targetValue == "1"

	if it's an integer && type == "minAvailable"; then

		targetValue == "0"








*/
