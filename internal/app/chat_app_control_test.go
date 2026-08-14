package app

import "testing"

func TestPlanChatMessageAppControlDoesNotClaimAcquisitionProviderReview(t *testing.T) {
	if intent, ok := planChatMessageAppControl("review integration provider setup", nil); ok {
		t.Fatalf("acquisition planning must bypass app control; got capability=%q", intent.CapabilityID)
	}
}

func TestPlanChatMessageAppControlRetainsExplicitProviderBackedActions(t *testing.T) {
	for _, prompt := range []string{
		"generate listing content",
		"create catalog content",
		"analyze image",
		"process image",
	} {
		intent, ok := planChatMessageAppControl(prompt, nil)
		if !ok || intent.CapabilityID != "content_generate" || !intent.SetupNeeded {
			t.Fatalf("prompt %q must retain provider setup guidance, got ok=%v intent=%+v", prompt, ok, intent)
		}
	}
}
