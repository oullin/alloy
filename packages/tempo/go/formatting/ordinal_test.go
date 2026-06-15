package formatting

import "testing"

func TestOrdinal(t *testing.T) {
	cases := map[int]string{
		1:   "1st",
		2:   "2nd",
		3:   "3rd",
		4:   "4th",
		11:  "11th",
		12:  "12th",
		13:  "13th",
		21:  "21st",
		22:  "22nd",
		23:  "23rd",
		111: "111th",
	}

	for value, want := range cases {
		if got := Ordinal(value); got != want {
			t.Fatalf("Ordinal(%d) = %q, want %q", value, got, want)
		}
	}
}
