package game

import "fmt"

type ShipType string

const (
	Carrier    ShipType = "Carrier"
	Battleship ShipType = "Battleship"
	Cruiser    ShipType = "Cruiser"
	Submarine  ShipType = "Submarine"
	Destroyer  ShipType = "Destroyer"

	ShipCarrier    = Carrier
	ShipBattleship = Battleship
	ShipCruiser    = Cruiser
	ShipSubmarine  = Submarine
	ShipDestroyer  = Destroyer
)

var standardShipLengths = map[ShipType]int{
	Carrier:    5,
	Battleship: 4,
	Cruiser:    3,
	Submarine:  3,
	Destroyer:  2,
}

var standardShipOrder = []ShipType{Carrier, Battleship, Cruiser, Submarine, Destroyer}

// Ship tracks one vessel and the coordinates that have been hit.
type Ship struct {
	Type        ShipType
	Length      int
	Coordinates []Coordinate
	Hits        map[Coordinate]bool
}

func NewShip(shipType ShipType) (Ship, error) {
	length, ok := standardShipLengths[shipType]
	if !ok {
		return Ship{}, fmt.Errorf("unknown ship type %q", shipType)
	}
	return Ship{Type: shipType, Length: length, Hits: make(map[Coordinate]bool)}, nil
}

func (s Ship) IsHit(coordinate Coordinate) bool {
	return s.Hits != nil && s.Hits[coordinate]
}

func (s *Ship) RegisterHit(coordinate Coordinate) bool {
	if !s.contains(coordinate) || s.IsHit(coordinate) {
		return false
	}
	if s.Hits == nil {
		s.Hits = make(map[Coordinate]bool)
	}
	s.Hits[coordinate] = true
	return true
}

func (s Ship) IsSunk() bool {
	if len(s.Coordinates) == 0 {
		return false
	}
	for _, coordinate := range s.Coordinates {
		if !s.IsHit(coordinate) {
			return false
		}
	}
	return true
}

func (s Ship) RemainingHealth() int {
	remaining := len(s.Coordinates)
	for _, coordinate := range s.Coordinates {
		if s.IsHit(coordinate) {
			remaining--
		}
	}
	return remaining
}

func (s Ship) contains(coordinate Coordinate) bool {
	for _, placed := range s.Coordinates {
		if placed == coordinate {
			return true
		}
	}
	return false
}

// Fleet is the collection of ships on one player's board.
type Fleet []Ship

func StandardFleet() Fleet {
	fleet := make(Fleet, 0, len(standardShipOrder))
	for _, shipType := range standardShipOrder {
		ship, _ := NewShip(shipType)
		fleet = append(fleet, ship)
	}
	return fleet
}

func NewFleet() Fleet {
	return StandardFleet()
}

func (f Fleet) Find(shipType ShipType) (Ship, bool) {
	for _, ship := range f {
		if ship.Type == shipType {
			return ship, true
		}
	}
	return Ship{}, false
}

func (f Fleet) AllSunk() bool {
	if len(f) == 0 {
		return false
	}
	for _, ship := range f {
		if !ship.IsSunk() {
			return false
		}
	}
	return true
}

func (f Fleet) RemainingShips() int {
	remaining := 0
	for _, ship := range f {
		if !ship.IsSunk() {
			remaining++
		}
	}
	return remaining
}

func (f Fleet) IsComplete() bool {
	if len(f) != len(standardShipOrder) {
		return false
	}
	for _, shipType := range standardShipOrder {
		ship, ok := f.Find(shipType)
		if !ok || len(ship.Coordinates) != ship.Length || ship.Length != standardShipLengths[shipType] {
			return false
		}
	}
	return true
}
