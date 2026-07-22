Biết. Và đúng là ý bạn đang muốn làm **TUI**, không phải CLI kiểu `fmt.Println()` đơn giản.

OpenCode hiện dùng **OpenTUI**, một terminal UI core viết bằng Zig với binding cho TypeScript; nó hỗ trợ layout theo component, input tương tác và render giao diện phức tạp trong terminal. ([opencode.ai][1])

Với **ShipWrek CLI viết bằng Go**, mình không cần cố bắt chước đúng công nghệ của OpenCode. Stack hợp lý nhất là:

```text
Go
├── Bubble Tea    // event loop + state management
├── Lip Gloss     // màu, border, layout, spacing
└── Bubbles       // text input, spinner, list, viewport
```

Bubble Tea dùng kiến trúc Elm: `Model → Update → View`, và được thiết kế cho cả TUI đơn giản lẫn full-screen phức tạp. ([GitHub][2])

## ShipWrek nên trông như thế này

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
│   4  · · ■ ■ ■ · · · · ·             4  · · · · · · · · · ·         │
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

Không cần bắt người chơi gõ `C7`. Có thể điều khiển bằng:

* `WASD` hoặc phím mũi tên để di chuyển cursor
* `Enter` để bắn hoặc đặt tàu
* `R` để xoay tàu
* `Tab` để đổi panel
* `Esc` để quay lại
* `Q` để thoát
* `?` mở help overlay

## Các màn hình

### Main menu

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

### LAN lobby

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

### Placement screen

Tàu được chọn ở sidebar, cursor nằm trên board, vùng đặt hợp lệ được highlight.

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

## Kiến trúc code nên sửa lại

```text
shipwrek-cli/
├── cmd/
│   └── shipwrek/
│       └── main.go
│
├── internal/
│   ├── app/
│   │   ├── app.go
│   │   ├── messages.go
│   │   └── keymap.go
│   │
│   ├── screen/
│   │   ├── menu/
│   │   │   └── model.go
│   │   ├── lobby/
│   │   │   └── model.go
│   │   ├── placement/
│   │   │   └── model.go
│   │   ├── battle/
│   │   │   └── model.go
│   │   └── result/
│   │       └── model.go
│   │
│   ├── component/
│   │   ├── board.go
│   │   ├── fleet.go
│   │   ├── statusbar.go
│   │   ├── modal.go
│   │   └── header.go
│   │
│   ├── game/
│   │   ├── board.go
│   │   ├── ship.go
│   │   ├── coordinate.go
│   │   ├── match.go
│   │   └── rules.go
│   │
│   ├── network/
│   │   ├── host.go
│   │   ├── client.go
│   │   ├── peer.go
│   │   ├── protocol.go
│   │   └── discovery.go
│   │
│   └── theme/
│       ├── colors.go
│       ├── styles.go
│       └── icons.go
│
├── docs/
│   ├── SPEC.md
│   └── PROTOCOL.md
│
├── go.mod
└── README.md
```

## Bubble Tea model

App có một model gốc quản lý màn hình hiện tại:

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
	width  int
	height int

	screen Screen

	menu      menu.Model
	lobby     lobby.Model
	placement placement.Model
	battle    battle.Model
	result    result.Model
}
```

`Update()` chuyển event tới màn hình tương ứng:

```go
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case ChangeScreenMsg:
		m.screen = msg.Screen

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	}

	switch m.screen {
	case ScreenMenu:
		updated, cmd := m.menu.Update(msg)
		m.menu = updated
		return m, cmd

	case ScreenLobby:
		updated, cmd := m.lobby.Update(msg)
		m.lobby = updated
		return m, cmd

	case ScreenPlacement:
		updated, cmd := m.placement.Update(msg)
		m.placement = updated
		return m, cmd

	case ScreenBattle:
		updated, cmd := m.battle.Update(msg)
		m.battle = updated
		return m, cmd
	}

	return m, nil
}
```

Điểm quan trọng là networking không được block TUI. TCP receive phải chạy thành Bubble Tea command hoặc goroutine rồi trả event về:

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

		return NetworkMessageMsg{
			Message: message,
		}
	}
}
```

Sau khi xử lý xong một message, lại gọi `WaitForNetworkMessage(peer)` để chờ message tiếp theo.

## Dependencies

```bash
go get github.com/charmbracelet/bubbletea
go get github.com/charmbracelet/lipgloss
go get github.com/charmbracelet/bubbles
```

Không cần giữ mục tiêu “standard library only” nữa. Nó sẽ giới hạn phần đẹp nhất của project.

## Chỉnh lại định hướng tài liệu

Tên vẫn là **ShipWrek CLI**, nhưng mô tả chính xác hơn là:

> A polished terminal multiplayer Battleship game for local networks.

Thay vì:

> lightweight CLI game

Nên chốt các yêu cầu UI:

* Full-screen alternate terminal buffer
* Responsive theo kích thước terminal
* Keyboard-first navigation
* Mouse không bắt buộc
* Unicode icons có fallback ASCII
* Màu có fallback cho terminal không hỗ trợ true color
* Không nháy màn hình khi render
* Modal help và confirmation
* Animated spinner trong lobby
* Toast khi kết nối, bắn trúng hoặc mất kết nối
* Board cursor và focused panel rõ ràng
* Terminal resize không làm hỏng layout

## Stack cuối cùng

```text
ShipWrek CLI

Language       Go
TUI framework  Bubble Tea
Styling        Lip Gloss
Components     Bubbles
Networking     TCP
Discovery      UDP broadcast
Protocol       JSON Lines
Architecture   Elm-style event loop
Distribution   Single executable
```

Đây sẽ không còn là project “in vài dòng ra terminal” nữa, mà là một **terminal application hoàn chỉnh**. Với AI hỗ trợ, chọn TUI tử tế ngay từ đầu hợp lý hơn nhiều so với làm bản thô rồi phải đập lại toàn bộ UI sau.

[1]: https://opencode.ai/?utm_source=chatgpt.com "OpenCode | The open source AI coding agent"
[2]: https://github.com/charmbracelet/bubbletea?utm_source=chatgpt.com "charmbracelet/bubbletea: A powerful little TUI framework"
