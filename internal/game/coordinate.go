package game

import (
	"fmt"
	"strconv"
	"strings"
)

const BoardSize = 10

// Coordinate identifies a cell on the board. X is the column (A-J) and Y is
// the row (1-10). Both values are zero-based internally.
type Coordinate struct {
	X int
	Y int
}

func (c Coordinate) IsValid() bool {
	return c.X >= 0 && c.X < BoardSize && c.Y >= 0 && c.Y < BoardSize
}

// String returns the user-facing form of a coordinate, for example A1 or J10.
func (c Coordinate) String() string {
	if !c.IsValid() {
		return "<invalid>"
	}
	return fmt.Sprintf("%c%d", 'A'+c.X, c.Y+1)
}

// ParseCoordinate parses the user-facing form of a coordinate, such as C7.
func ParseCoordinate(value string) (Coordinate, error) {
	value = strings.TrimSpace(strings.ToUpper(value))
	if len(value) < 2 {
		return Coordinate{}, fmt.Errorf("invalid coordinate %q: expected a letter and row", value)
	}

	column := value[0]
	if column < 'A' || column >= 'A'+BoardSize {
		return Coordinate{}, fmt.Errorf("invalid coordinate %q: column must be A-J", value)
	}

	row, err := strconv.Atoi(value[1:])
	if err != nil || row < 1 || row > BoardSize {
		return Coordinate{}, fmt.Errorf("invalid coordinate %q: row must be 1-10", value)
	}

	return Coordinate{X: int(column - 'A'), Y: row - 1}, nil
}

type Direction uint8

const (
	Horizontal Direction = iota
	Vertical
)

func (d Direction) String() string {
	if d == Vertical {
		return "Vertical"
	}
	return "Horizontal"
}

func coordinatesFor(start Coordinate, length int, direction Direction) []Coordinate {
	if length <= 0 || !start.IsValid() {
		return nil
	}

	coordinates := make([]Coordinate, length)
	for i := range coordinates {
		coordinates[i] = start
		if direction == Vertical {
			coordinates[i].Y += i
		} else {
			coordinates[i].X += i
		}
	}
	return coordinates
}
