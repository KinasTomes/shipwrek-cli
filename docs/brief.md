# ShipWrek CLI – Project Brief (v0.2)

> A polished, interactive TUI multiplayer Battleship game for local area networks written in Go.

---

## 1. Overview

**ShipWrek CLI** is a modern terminal-based multiplayer Battleship game built with **Go** and the **Bubble Tea** TUI framework ecosystem.

Rather than relying on basic command-line text prints (`fmt.Println`), ShipWrek CLI delivers a rich, full-screen, component-driven Terminal User Interface (TUI). Players interact with visual board grids, styled panels, animated indicators, and keyboard-driven targeting over a Local Area Network (LAN).

The architecture enforces a strict separation of concerns (Clean Architecture), separating core game logic, TCP/UDP network communication, and Bubble Tea presentation layers.

---

## 2. Goals & Non-Goals

### Goals
- **Rich TUI Experience**: Responsive full-screen interface using standard terminal alternate buffer (`tea.EnterAltScreen`).
- **Keyboard-First Controls**: Smooth navigation using arrow keys / WASD, hotkeys, and interactive selection.
- **LAN Multiplayer**: Low-latency peer connectivity via TCP with UDP Broadcast discovery for seamless local lobby discovery.
- **Clean Architecture**: Decoupled core game engine, transport protocols, and TUI component screens.
- **Cross-Platform**: Support Windows, Linux, and macOS terminal environments with proper Unicode/ASCII fallbacks.

### Non-Goals (v1.0)
- Internet matchmaking or cloud servers (LAN only).
- User authentication or persistent database accounts.
- AI bot opponents (planned for future releases).
- GUI / Web graphics engines (pure TUI emphasis).

---

## 3. Tech Stack

| Layer | Technology | Purpose |
| :--- | :--- | :--- |
| **Language** | Go 1.25+ | Runtime & standard libraries |
| **TUI Core** | [`charmbracelet/bubbletea`](https://github.com/charmbracelet/bubbletea) | Elm-architecture event loop & state management |
| **Styling & Layout** | [`charmbracelet/lipgloss`](https://github.com/charmbracelet/lipgloss) | Colors, borders, padding, alignment & themes |
| **UI Components** | [`charmbracelet/bubbles`](https://github.com/charmbracelet/bubbles) | Text inputs, spinners, lists, viewports |
| **Networking** | Go `net` (TCP & UDP) | Sockets for game state sync & UDP broadcast discovery |
| **Protocol** | JSON / JSON Lines | Structured network message serialization |

---

## 4. Controls & Keyboard Navigation

Navigation is entirely keyboard-driven:

| Key | Action |
| :--- | :--- |
| `↑ ↓ ← →` / `W A S D` | Move board cursor or menu selection |
| `Enter` | Confirm menu action / Place ship / Fire shot |
| `R` | Rotate ship during placement phase (Horizontal / Vertical) |
| `Tab` | Switch focus between UI panels |
| `Esc` | Back to previous screen / Cancel selection |
| `?` | Toggle Help & Keymap Overlay |
| `Q` / `Ctrl+C` | Quit game safely |

---

## 5. UI Screen Mockups

### 5.1 Main Menu
```text
╭──────────────────────────────╮
│         ⚓ SHIPWREK          │
│                              │
│      ▸ Host LAN Game         │
│        Join LAN Game         │
│        How to Play           │
│        Settings              │
│        Quit                  │
│                              │
│  ↑/↓ navigate • enter select │
╰──────────────────────────────╯
```

### 5.2 LAN Lobby
```text
╭──────────────────────────────────────────────────────╮
│ LAN LOBBY                                            │
├──────────────────────────────────────────────────────┤
│ Room       rosemary-fleet                            │
│ Host       192.168.1.24:7777                         │
│ Status     Waiting for opponent...                   │
│                                                      │
│ Players                                              │
│ ● Rosemary                              READY         │
│ ◌ Waiting for player...                              │
│                                                      │
│                       ◜ spinning indicator ◞          │
├──────────────────────────────────────────────────────┤
│ esc back                                             │
╰──────────────────────────────────────────────────────╯
```

### 5.3 Fleet Placement Screen
```text
╭────────────────────────────────────────────────────────────╮
│ PLACE YOUR FLEET                            Ships left: 3   │
├────────────────────────────────────────────────────────────┤
│                                                            │
│    A B C D E F G H I J       Selected                      │
│ 1  · · · · · · · · · ·       Destroyer                     │
│ 2  · · ▒ ▒ ▒ · · · · ·       Length: 3                     │
│ 3  · · · · · · · · · ·       Direction: Horizontal         │
│                                                            │
│                              Fleet                         │
│                              ✓ Carrier                     │
│                              ✓ Battleship                  │
│                              ▸ Cruiser                     │
│                              ○ Submarine                   │
│                              ○ Destroyer                   │
├────────────────────────────────────────────────────────────┤
│ arrows move • r rotate • enter place • backspace remove   │
╰────────────────────────────────────────────────────────────╯
```

### 5.4 Battle Screen
```text
╭──────────────────────────────────────────────────────────────────────╮
│  ⚓ SHIPWREK                                      LAN • 12 ms • P2   │
├──────────────────────────────────────────────────────────────────────┤
│                                                                      │
│  YOUR FLEET                         ENEMY WATERS                      │
│                                                                      │
│      A B C D E F G H I J                A B C D E F G H I J          │
│   1  · · · · · · · · · ·             1  · · · · · · · · · ·         │
│   2  · ■ ■ ■ ■ · · · · ·             2  · · · ○ · · · · · ·         │
│   3  · · · · · · · · · ·             3  · · · ╳ · · · · · ·         │
│   4  · ■ ■ ■ · · · · ·             4  · · · · · · · · · ·         │
│   5  · · · · · · · · · ·             5  · · · · ◉ · · · · ·         │
│                                                                      │
│  Fleet                                                              │
│  Carrier      █████   5/5                                            │
│  Battleship   ████    3/4                                            │
│  Cruiser      ███     3/3                                            │
│                                                                      │
├──────────────────────────────────────────────────────────────────────┤
│  YOUR TURN                                                           │
│  Select a target with ↑ ↓ ← →, then press Enter                      │
├──────────────────────────────────────────────────────────────────────┤
│  q quit   r rotate   enter confirm   tab switch board   ? help       │
╰──────────────────────────────────────────────────────────────────────╯
```

---

## 6. Project Directory Structure

```text
shipwrek-cli/
├── cmd/
│   └── shipwrek/
│       └── main.go
├── internal/
│   ├── app/
│   │   ├── app.go             # Root Bubble Tea model & router
│   │   ├── messages.go        # Shared Bubble Tea custom messages (tea.Msg)
│   │   └── keymap.go          # Keybindings configuration
│   ├── screen/
│   │   ├── menu/              # Main menu screen model & view
│   │   │   └── model.go
│   │   ├── lobby/             # Host/Join lobby screen
│   │   │   └── model.go
│   │   ├── placement/         # Interactive ship placement screen
│   │   │   └── model.go
│   │   ├── battle/            # Main combat screen & board grids
│   │   │   └── model.go
│   │   └── result/            # Game over victory/defeat summary
│   │       └── model.go
│   ├── component/
│   │   ├── board.go           # Renderable grid component
│   │   ├── fleet.go           # Fleet status list component
│   │   ├── statusbar.go       # Bottom status & shortcut bar
│   │   ├── modal.go           # Help & confirmation popups
│   │   └── header.go          # Top title & connection indicator
│   ├── game/
│   │   ├── board.go           # Grid data model & hit checks
│   │   ├── ship.go            # Ship definitions & health tracking
│   │   ├── coordinate.go      # X/Y coordinate utilities
│   │   ├── match.go           # Game state machine (turns, state)
│   │   └── rules.go           # Game placement & hit validation rules
│   ├── network/
│   │   ├── host.go            # TCP server listener implementation
│   │   ├── client.go          # TCP client connection logic
│   │   ├── peer.go            # Connection reader/writer wrapper
│   │   ├── protocol.go        # JSON message schemas & encoder/decoder
│   │   └── discovery.go       # UDP broadcast beacon & listener
│   └── theme/
│       ├── colors.go          # Color palette definitions (Lip Gloss)
│       ├── styles.go          # Reusable component styles
│       └── icons.go           # UTF-8 icons & ASCII fallback symbols
├── docs/
│   ├── brief.md               # Project brief & architectural plan
│   ├── SPEC.md                # Detailed technical spec
│   └── PROTOCOL.md             # Network packet specification
├── go.mod
└── README.md
```

---

## 7. Architecture & Bubble Tea Integration

### 7.1 Elm Architecture Pattern
The app uses a unified root model (`app.Model`) that delegates updates and rendering to active screen sub-models:

```go
type Screen int

const (
    ScreenMenu Screen = iota
    ScreenLobby
    ScreenPlacement
    ScreenBattle
    ScreenResult
)

type Model struct {
    width     int
    height    int
    screen    Screen
    menu      menu.Model
    lobby     lobby.Model
    placement placement.Model
    battle    battle.Model
    result    result.Model
}
```

### 7.2 Non-Blocking Asynchronous Networking
Network sockets must **never** block the UI thread. TCP read loops run in asynchronous background goroutines that produce Bubble Tea commands (`tea.Cmd`) returning messages (`tea.Msg`):

```go
type NetworkMessageMsg struct {
    Message network.Message
}

func WaitForNetworkMessage(peer *network.Peer) tea.Cmd {
    return func() tea.Msg {
        message, err := peer.Receive()
        if err != nil {
            return NetworkErrorMsg{Err: err}
        }
        return NetworkMessageMsg{Message: message}
    }
}
```

---

## 8. Game Rules & Ships

### Fleet Composition

| Ship Class | Length | Grid Symbol |
| :--- | :---: | :---: |
| **Carrier** | 5 | `█████` |
| **Battleship** | 4 | `████` |
| **Cruiser** | 3 | `███` |
| **Submarine** | 3 | `███` |
| **Destroyer** | 2 | `██` |

### Board Symbols

| State | Character | Description |
| :--- | :---: | :--- |
| Empty / Water | `·` | Unexplored coordinate |
| Ship (Own) | `■` | Intact ship segment on own grid |
| Hit | `╳` | Damaged ship segment |
| Miss | `○` | Shot landed in open water |
| Active Target | `◉` | Current user cursor position |

---

## 9. Development Roadmap

### Milestone 1: Pure Engine & Domain Model
- Board, Ship, Coordinate models in `internal/game`.
- Placement validation rules and hit/miss mechanics.
- Unit testing core game rules without UI or Network.

### Milestone 2: TUI Layout & Theme Foundations
- Bubble Tea root application loop setup.
- Lip Gloss styles, colors, and layout wrappers.
- Main Menu and Fleet Placement screen UI components.

### Milestone 3: Networking & Protocol Layer
- TCP server/client connection handling.
- JSON Lines protocol definition for game events (`place`, `shoot`, `turn`, `result`, `sync`).
- LAN UDP broadcast beacon for auto-discovery in Lobby.

### Milestone 4: Async Network + TUI Integration
- Connect network events to Bubble Tea `tea.Cmd` event loop.
- Full game loop between Host and Joiner over LAN.
- Interactive turn handling, real-time board updates, and winner resolution.

### Milestone 5: Polish & UX Enhancements
- Animated lobby spinner (`bubbles/spinner`).
- Toast notifications for game events (Hits, Misses, Opponent Ready, Connection Status).
- Terminal resizing handlers (`tea.WindowSizeMsg`).
- Keyboard shortcut overlays and help modals.

---

## 10. License

MIT
