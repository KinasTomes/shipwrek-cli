package game

// CanPlaceShip is the package-level form used by placement screens and tests.
func CanPlaceShip(board *Board, ship Ship) bool {
	return board != nil && board.CanPlaceShip(ship)
}

func canPlaceShip(board Board, ship Ship) bool {
	if ship.Length <= 0 || len(ship.Coordinates) != ship.Length {
		return false
	}
	if expected, ok := standardShipLengths[ship.Type]; !ok || expected != ship.Length {
		return false
	}
	if !validShipShape(ship.Coordinates) {
		return false
	}
	if _, exists := board.Fleet.Find(ship.Type); exists {
		return false
	}

	seen := make(map[Coordinate]struct{}, len(ship.Coordinates))
	for _, coordinate := range ship.Coordinates {
		if !coordinate.IsValid() {
			return false
		}
		if _, exists := seen[coordinate]; exists {
			return false
		}
		seen[coordinate] = struct{}{}
		if board.Cells[coordinate.Y][coordinate.X] != CellEmpty {
			return false
		}
	}
	return true
}

func validShipShape(coordinates []Coordinate) bool {
	if len(coordinates) < 2 {
		return false
	}

	first := coordinates[0]
	allSameX := true
	allSameY := true
	minX, maxX := first.X, first.X
	minY, maxY := first.Y, first.Y
	for _, coordinate := range coordinates[1:] {
		allSameX = allSameX && coordinate.X == first.X
		allSameY = allSameY && coordinate.Y == first.Y
		if coordinate.X < minX {
			minX = coordinate.X
		}
		if coordinate.X > maxX {
			maxX = coordinate.X
		}
		if coordinate.Y < minY {
			minY = coordinate.Y
		}
		if coordinate.Y > maxY {
			maxY = coordinate.Y
		}
	}

	return (allSameX && maxY-minY+1 == len(coordinates)) ||
		(allSameY && maxX-minX+1 == len(coordinates))
}
