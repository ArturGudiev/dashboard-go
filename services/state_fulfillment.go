package services

import (
	"arturgudiev/dashboard/ent"
	"time"
)

// ComputeRequirementIsFulfilled returns whether a requirement is fulfilled based on its latest check.
// Returns nil when status is unclear (no check, stale check, or missing once_in_days).
func ComputeRequirementIsFulfilled(requirement *ent.StateRequirement, latestCheck *ent.StateRequirementCheck, now time.Time) *bool {
	if requirement.OnceInDays == nil {
		return nil
	}
	if latestCheck == nil {
		return nil
	}

	validUntil := latestCheck.DateTime.AddDate(0, 0, *requirement.OnceInDays)
	if !validUntil.After(now) {
		return nil
	}

	fulfilled := latestCheck.IsFulfilled
	return &fulfilled
}

// ComputeStateIsFulfilled aggregates requirement fulfillment into a state-level status.
// Returns false if any requirement is not fulfilled, true if all are fulfilled,
// and nil otherwise (no requirements, unclear requirements, or mixed status).
func ComputeStateIsFulfilled(requirementStatuses []*bool) *bool {
	if len(requirementStatuses) == 0 {
		return nil
	}

	allFulfilled := true
	for _, status := range requirementStatuses {
		if status == nil {
			allFulfilled = false
			continue
		}
		if !*status {
			falseVal := false
			return &falseVal
		}
	}

	if allFulfilled {
		trueVal := true
		return &trueVal
	}

	return nil
}
