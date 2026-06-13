package tempo

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

type fixture struct {
	Metadata struct {
		CarbonVersion string `json:"carbonVersion"`
		Source        string `json:"source"`
		Timezone      string `json:"timezone"`
	} `json:"metadata"`
	Cases []struct {
		Name               string `json:"name"`
		Input              string `json:"input"`
		ExpectedISO        string `json:"expectedIso"`
		ExpectedDate       string `json:"expectedDate"`
		AddDays            int    `json:"addDays"`
		ExpectedAddDaysISO string `json:"expectedAddDaysIso"`
	} `json:"cases"`
}

func TestSharedCarbonFixtures(t *testing.T) {
	contents, err := os.ReadFile(filepath.Join("..", "spec", "fixtures", "core.json"))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	var data fixture
	if err := json.Unmarshal(contents, &data); err != nil {
		t.Fatalf("decode fixture: %v", err)
	}

	if data.Metadata.Source != "carbon" || data.Metadata.CarbonVersion != "3.11.4" {
		t.Fatalf("unexpected fixture metadata: %+v", data.Metadata)
	}

	for _, item := range data.Cases {
		t.Run(item.Name, func(t *testing.T) {
			parsed, err := Parse(item.Input)
			if err != nil {
				t.Fatalf("parse fixture input: %v", err)
			}

			if got := parsed.ISOString(); got != item.ExpectedISO {
				t.Fatalf("ISOString() = %q, want %q", got, item.ExpectedISO)
			}

			if got := parsed.DateString(); got != item.ExpectedDate {
				t.Fatalf("DateString() = %q, want %q", got, item.ExpectedDate)
			}

			if item.ExpectedAddDaysISO != "" {
				if got := parsed.AddDays(item.AddDays).ISOString(); got != item.ExpectedAddDaysISO {
					t.Fatalf("AddDays(%d).ISOString() = %q, want %q", item.AddDays, got, item.ExpectedAddDaysISO)
				}
			}
		})
	}
}
