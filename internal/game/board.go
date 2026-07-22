package game

import (
	"errors"
	"fmt"
)

type CellState uint8

const (
	CellEmpty CellState = iota
	CellShip
	CellHit
	CellMiss
)

var (
	ErrInvalidCoordinate = errors.New("coordinate is outside the board")
	ErrInvalidShip       = errors.New("ship is invalid")
	ErrOutOfBounds       = errors.New("ship would extend beyond the board")
	ErrOverlap           = errors.New("ship overlaps another ship")
	ErrDuplicateShip     = errors.New("ship type is already placed")
	ErrInvalidDirection  = errors.New("ship direction is invalid")
	ErrAlreadyShot       = errors.New("coordinate has already been targeted")
)

// Board is a fixed 10x10 game board. Cells are indexed as Cells[Y][X].
type Board struct {
	Cells FleetCells
	Fleet Fleet
}

// FleetCells is named to make the fixed board shape explicit to callers.
type FleetCells [BoardSize][BoardSize]CellState

func NewBoard() Board {
	return Board{Cells: FleetCells{}, Fleet: make(Fleet, 0, len(standardShipOrder))}
}

func (b Board) CellAt(coordinate Coordinate) (CellState, error) {
	if !coordinate.IsValid() {
		return CellEmpty, ErrInvalidCoordinate
	}
	return b.Cells[coordinate.Y][coordinate.X], nil
}

func (b Board) IsReady() bool {
	return b.Fleet.IsComplete()
}

func (b Board) RemainingShips() int {
	return b.Fleet.RemainingShips()
}

func (b Board) IsDefeated() bool {
	return b.IsReady() && b.Fleet.AllSunk()
}

func (b Board) CanPlaceShip(ship Ship) bool {
	return canPlaceShip(b, ship)
}

// PlaceShip adds a fully described ship to the board after validating it.
func (b *Board) PlaceShip(ship Ship) error {
	if b == nil {
		return ErrInvalidShip
	}
	if ship.Length <= 0 || len(ship.Coordinates) != ship.Length {
		return ErrInvalidShip
	}
	if expected, ok := standardShipLengths[ship.Type]; !ok || expected != ship.Length {
		return ErrInvalidShip
	}
	if !validShipShape(ship.Coordinates) {
		return ErrInvalidShip
	}
	if _, exists := b.Fleet.Find(ship.Type); exists {
		return ErrDuplicateShip
	}

	seen := make(map[Coordinate]struct{}, len(ship.Coordinates))
	for _, coordinate := range ship.Coordinates {
		if !coordinate.IsValid() {
			return ErrOutOfBounds
		}
		if _, exists := seen[coordinate]; exists {
			return ErrInvalidShip
		}
		seen[coordinate] = struct{}{}
		if b.Cells[coordinate.Y][coordinate.X] != CellEmpty {
			return ErrOverlap
		}
	}

	placed := ship
	placed.Coordinates = append([]Coordinate(nil), ship.Coordinates...)
	placed.Hits = make(map[Coordinate]bool)
	b.Fleet = append(b.Fleet, placed)
	for _, coordinate := range placed.Coordinates {
		b.Cells[coordinate.Y][coordinate.X] = CellShip
	}
	return nil
}

// Place creates and places a standard ship from a starting coordinate and direction.
func (b *Board) Place(shipType ShipType, start Coordinate, direction Direction) error {
	if direction != Horizontal && direction != Vertical {
		return ErrInvalidDirection
	}
	ship, err := NewShip(shipType)
	if err != nil {
		return err
	}
	ship.Coordinates = coordinatesFor(start, ship.Length, direction)
	return b.PlaceShip(ship)
}

type ShotOutcome uint8

const (
	ShotMiss ShotOutcome = iota
	ShotHit
	ShotSunk
	ShotVictory
)

type ShotResult struct {
	Coordinate Coordinate
	Outcome    ShotOutcome
	ShipType   ShipType
}

func (r ShotResult) Hit() bool {
	return r.Outcome == ShotHit || r.Outcome == ShotSunk || r.Outcome == ShotVictory
}

func (r ShotResult) Sunk() bool {
	return r.Outcome == ShotSunk || r.Outcome == ShotVictory
}

func (r ShotResult) Victory() bool {
	return r.Outcome == ShotVictory
}

// ReceiveShot applies one shot to the board and reports its outcome.
func (b *Board) ReceiveShot(coordinate Coordinate) (ShotResult, error) {
	if b == nil || !coordinate.IsValid() {
		return ShotResult{Coordinate: coordinate}, ErrInvalidCoordinate
	}
	cell := b.Cells[coordinate.Y][coordinate.X]
	if cell == CellHit || cell == CellMiss {
		return ShotResult{Coordinate: coordinate}, ErrAlreadyShot
	}
	if cell == CellEmpty {
		b.Cells[coordinate.Y][coordinate.X] = CellMiss
		return ShotResult{Coordinate: coordinate, Outcome: ShotMiss}, nil
	}

	for index := range b.Fleet {
		if !b.Fleet[index].contains(coordinate) {
			continue
		}
		b.Fleet[index].RegisterHit(coordinate)
		b.Cells[coordinate.Y][coordinate.X] = CellHit
		result := ShotResult{Coordinate: coordinate, Outcome: ShotHit, ShipType: b.Fleet[index].Type}
		if b.Fleet[index].IsSunk() {
			result.Outcome = ShotSunk
			if b.Fleet.AllSunk() {
				result.Outcome = ShotVictory
			}
		}
		return result, nil
	}

	return ShotResult{Coordinate: coordinate}, fmt.Errorf("board invariant violated at %s", coordinate)
}

func (b Board) ShipAt(coordinate Coordinate) (Ship, bool) {
	if !coordinate.IsValid() {
		return Ship{}, false
	}
	for _, ship := range b.Fleet {
		if ship.contains(coordinate) {
			return ship, true
		}
	}
	return Ship{}, false
}
