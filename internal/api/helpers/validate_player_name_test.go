package helpers

import (
	"log"
	"testing"
)

// Test successfuly validate func

func TestFetchCharacterData(t *testing.T) {
	gotVal, gotErr := FetchCharacterData("Niru", "na")

	if gotErr != nil {
		t.Errorf("got error " + gotErr.Error())
		return
	}
	if gotVal == nil {
		t.Error("validate failed for expected known working character name")
		return
	}
	log.Println(gotVal.CharacterName)
}
