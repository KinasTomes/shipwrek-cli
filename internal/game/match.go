package game

import "errors"

type MatchState uint8

const (
	Setup MatchState = iota
	PlayerTurn
	EnemyTurn
	GameOver
)

var (
	ErrMatchNotReady = errors.New("match requires both complete fleets")
	ErrNotYourTurn   = errors.New("it is not this player's turn")
	ErrMatchOver     = errors.New("match is already over")
)

// Match coordinates the two boards and enforces turn transitions. Board
// placement remains available until SetReady is called.
type Match struct {
	OwnBoard   Board
	EnemyBoard Board
	State      MatchState
	Won        bool
}

func NewMatch() *Match {
	return &Match{OwnBoard: NewBoard(), EnemyBoard: NewBoard(), State: Setup}
}

func NewMatchWithBoards(ownBoard, enemyBoard Board) *Match {
	return &Match{OwnBoard: ownBoard, EnemyBoard: enemyBoard, State: Setup}
}

// SetReady validates both fleets and starts the match on the player's turn.
func (m *Match) SetReady() error {
	if m == nil || !m.OwnBoard.IsReady() || !m.EnemyBoard.IsReady() {
		return ErrMatchNotReady
	}
	if m.State == GameOver {
		return ErrMatchOver
	}
	m.State = PlayerTurn
	return nil
}

func (m *Match) Start() error {
	return m.SetReady()
}

// Fire applies the player's shot to the enemy board and advances the turn
// unless the shot wins the match.
func (m *Match) Fire(coordinate Coordinate) (ShotResult, error) {
	if m == nil {
		return ShotResult{Coordinate: coordinate}, ErrMatchNotReady
	}
	if m.State == GameOver {
		return ShotResult{Coordinate: coordinate}, ErrMatchOver
	}
	if m.State != PlayerTurn {
		return ShotResult{Coordinate: coordinate}, ErrNotYourTurn
	}

	result, err := m.EnemyBoard.ReceiveShot(coordinate)
	if err != nil {
		return result, err
	}
	if result.Victory() {
		m.State = GameOver
		m.Won = true
	} else {
		m.State = EnemyTurn
	}
	return result, nil
}

// ReceiveEnemyShot applies the opponent's shot and returns the turn to the
// player unless the player's fleet has been defeated.
func (m *Match) ReceiveEnemyShot(coordinate Coordinate) (ShotResult, error) {
	if m == nil {
		return ShotResult{Coordinate: coordinate}, ErrMatchNotReady
	}
	if m.State == GameOver {
		return ShotResult{Coordinate: coordinate}, ErrMatchOver
	}
	if m.State != EnemyTurn {
		return ShotResult{Coordinate: coordinate}, ErrNotYourTurn
	}

	result, err := m.OwnBoard.ReceiveShot(coordinate)
	if err != nil {
		return result, err
	}
	if m.OwnBoard.IsDefeated() {
		m.State = GameOver
		m.Won = false
	} else {
		m.State = PlayerTurn
	}
	return result, nil
}
