package helpers

import "testing"

// Exercise the live endpoint so changes to Nexon's rankings API fail CI.
func TestFetchCharacterData(t *testing.T) {
	got, err := FetchCharacterData("niru", "na")
	if err != nil {
		t.Fatalf("live Nexon NA rankings lookup failed: %v", err)
	}
	if got == nil || got.CharacterName != "Niru" {
		t.Fatalf("live Nexon NA rankings lookup returned %+v, want Niru", got)
	}
	if got.Level <= 0 || got.JobName == "" || got.CharacterImgURL == "" {
		t.Error("live Nexon NA rankings response is missing the character's level, job, or image URL")
	}
}
