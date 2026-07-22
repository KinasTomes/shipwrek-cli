package game

import "testing"

func TestStandardFleet(t *testing.T) {
	fleet := StandardFleet()
	if len(fleet) != 5 {
		t.Fatalf("fleet has %d ships, want 5", len(fleet))
	}
	want := map[ShipType]int{Carrier: 5, Battleship: 4, Cruiser: 3, Submarine: 3, Destroyer: 2}
	for _, ship := range fleet {
		if ship.Length != want[ship.Type] {
			t.Errorf("%s length = %d, want %d", ship.Type, ship.Length, want[ship.Type])
		}
	}
}

func TestBoardPlacementRules(t *testing.T) {
	board := NewBoard()
	if err := board.Place(Carrier, Coordinate{X: 0, Y: 0}, Horizontal); err != nil {
		t.Fatalf("placing carrier: %v", err)
	}
	if board.Cells[0][0] != CellShip || board.Cells[0][4] != CellShip {
		t.Fatal("placed cells were not marked as ships")
	}

	if err := board.Place(Destroyer, Coordinate{X: 0, Y: 0}, Vertical); err == nil {
		t.Fatal("overlapping ship placement unexpectedly succeeded")
	}
	if err := board.Place(Battleship, Coordinate{X: 8, Y: 0}, Horizontal); err != ErrOutOfBounds {
		t.Fatalf("out-of-bounds placement error = %v, want %v", err, ErrOutOfBounds)
	}
	if err := board.Place(Carrier, Coordinate{X: 0, Y: 2}, Horizontal); err != ErrDuplicateShip {
		t.Fatalf("duplicate placement error = %v, want %v", err, ErrDuplicateShip)
	}
}

func TestReceiveShotTracksMissHitSunkAndVictory(t *testing.T) {
	board := NewBoard()
	if err := board.Place(Destroyer, Coordinate{X: 0, Y: 0}, Horizontal); err != nil {
		t.Fatal(err)
	}
	if err := board.Place(Submarine, Coordinate{X: 0, Y: 2}, Horizontal); err != nil {
		t.Fatal(err)
	}

	miss, err := board.ReceiveShot(Coordinate{X: 9, Y: 9})
	if err != nil || miss.Outcome != ShotMiss || board.Cells[9][9] != CellMiss {
		t.Fatalf("miss = %+v, err = %v", miss, err)
	}
	if _, err := board.ReceiveShot(Coordinate{X: 9, Y: 9}); err != ErrAlreadyShot {
		t.Fatalf("duplicate miss error = %v, want %v", err, ErrAlreadyShot)
	}

	hit, err := board.ReceiveShot(Coordinate{X: 0, Y: 0})
	if err != nil || hit.Outcome != ShotHit || hit.ShipType != Destroyer {
		t.Fatalf("first hit = %+v, err = %v", hit, err)
	}
	sunk, err := board.ReceiveShot(Coordinate{X: 1, Y: 0})
	if err != nil || sunk.Outcome != ShotSunk || !sunk.Sunk() || board.IsDefeated() {
		t.Fatalf("final hit = %+v, err = %v", sunk, err)
	}
}

func TestBoardWinsOnlyAfterEveryShipIsSunk(t *testing.T) {
	board := completeBoard()
	var final ShotResult
	for _, ship := range board.Fleet {
		for _, coordinate := range ship.Coordinates {
			result, err := board.ReceiveShot(coordinate)
			if err != nil {
				t.Fatalf("shooting %s: %v", coordinate, err)
			}
			final = result
		}
	}
	if final.Outcome != ShotVictory || !board.IsDefeated() || board.RemainingShips() != 0 {
		t.Fatalf("final = %+v, defeated = %v, remaining = %d", final, board.IsDefeated(), board.RemainingShips())
	}
}

func completeBoard() Board {
	board := NewBoard()
	placements := []struct {
		ship      ShipType
		start     Coordinate
		direction Direction
	}{
		{Carrier, Coordinate{X: 0, Y: 0}, Horizontal},
		{Battleship, Coordinate{X: 0, Y: 2}, Horizontal},
		{Cruiser, Coordinate{X: 0, Y: 4}, Horizontal},
		{Submarine, Coordinate{X: 0, Y: 6}, Horizontal},
		{Destroyer, Coordinate{X: 0, Y: 8}, Horizontal},
	}
	for _, placement := range placements {
		if err := board.Place(placement.ship, placement.start, placement.direction); err != nil {
			panic(err)
		}
	}
	return board
}

func TestMatchStateAndTurns(t *testing.T) {
	match := NewMatchWithBoards(completeBoard(), completeBoard())
	if err := match.SetReady(); err != nil {
		t.Fatalf("SetReady() error = %v", err)
	}
	if match.State != PlayerTurn {
		t.Fatalf("state = %v, want PlayerTurn", match.State)
	}
	if _, err := match.ReceiveEnemyShot(Coordinate{X: 9, Y: 9}); err != ErrNotYourTurn {
		t.Fatalf("out-of-turn shot error = %v, want %v", err, ErrNotYourTurn)
	}
	if _, err := match.Fire(Coordinate{X: 9, Y: 9}); err != nil || match.State != EnemyTurn {
		t.Fatalf("Fire() err = %v, state = %v", err, match.State)
	}
	if _, err := match.ReceiveEnemyShot(Coordinate{X: 9, Y: 9}); err != nil || match.State != PlayerTurn {
		t.Fatalf("ReceiveEnemyShot() err = %v, state = %v", err, match.State)
	}
}
