package game

import "testing"

func TestCoordinateStringAndParse(t *testing.T) {
	coordinate := Coordinate{X: 9, Y: 9}
	if got := coordinate.String(); got != "J10" {
		t.Fatalf("String() = %q, want J10", got)
	}
	parsed, err := ParseCoordinate(" j10 ")
	if err != nil {
		t.Fatalf("ParseCoordinate() error = %v", err)
	}
	if parsed != coordinate {
		t.Fatalf("ParseCoordinate() = %+v, want %+v", parsed, coordinate)
	}
	for _, value := range []string{"A0", "K1", "A11", "", "A"} {
		if _, err := ParseCoordinate(value); err == nil {
			t.Errorf("ParseCoordinate(%q) unexpectedly succeeded", value)
		}
	}
}
