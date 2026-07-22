# ShipWrek CLI – Software Requirements Specification (SRS)

> Tài liệu tả các yêu cầu chức năng (Functional Requirements) và phi chức năng (Non-Functional Requirements) cho ứng dụng TUI game Battleship chơi qua LAN bằng Go.

---

## 1. Giới thiệu (Introduction)

### 1.1 Mục đích (Purpose)
Tài liệu này định nghĩa chi tiết các yêu cầu kỹ thuật, chức năng và giao diện cho dự án **ShipWrek CLI**. Đây là căn cứ để thực thi việc chia nhỏ công việc (Task Breakdown), thiết kế kiến trúc và nghiệm thu sản phẩm.

### 1.2 Phạm vi sản phẩm (Product Scope)
- **Tên ứng dụng**: ShipWrek CLI
- **Loại ứng dụng**: Terminal User Interface (TUI) Multiplayer Game
- **Ngôn ngữ & Framework**: Go 1.25+, Charm Bubble Tea, Lip Gloss, Bubbles
- **Môi trường hoạt động**: Terminal chạy trên Windows, Linux, macOS kết nối mạng nội bộ (LAN).

---

## 2. Yêu cầu chức năng (Functional Requirements - FR)

### FR-01: Core Game Engine (Luật chơi & Bàn chơi)
- **FR-01.1**: Khởi tạo bàn chơi kích thước cố định $10 \times 10$ với tọa độ trục ngang A-J và trục dọc 1-10.
- **FR-01.2**: Quản lý Hạm đội (Fleet) gồm 5 loại tàu tiêu chuẩn:
  - Carrier (Độ dài: 5)
  - Battleship (Độ dài: 4)
  - Cruiser (Độ dài: 3)
  - Submarine (Độ dài: 3)
  - Destroyer (Độ dài: 2)
- **FR-01.3**: Validate việc đặt tàu: Tàu không được vượt quá biên bàn chơi, không được đè lên hoặc va chạm với tàu khác. Hỗ trợ xoay tàu (Ngang / Dọc).
- **FR-01.4**: Xử lý logic bắn (Targeting & Shot validation): Kiếm tra tọa độ bắn hợp lệ, ghi nhận kết quả **Hit** (Trúng), **Miss** (Trượt), đánh dấu tàu chìm (Sunk) khi tất cả các ô của tàu đó bị bắn trúng.
- **FR-01.5**: Xác định điều kiện chiến thắng: Người chơi thắng khi bắn hạ toàn bộ 5 tàu của đối phương.

### FR-02: Terminal User Interface (Màn hình & Linh kiện UI)
- **FR-02.1**: **Main Menu Screen**: Hiển thị danh sách tùy chọn (Host Game, Join Game, How to Play, Settings, Quit) hỗ trợ di chuyển bằng phím mũi tên / WASD.
- **FR-02.2**: **Lobby Screen**:
  - Khi Host: Hiển thị IP address + Port lắng nghe, trạng thái chờ đối phương (spinner indicator).
  - Khi Join: Cho phép nhập IP Host hoặc tự chọn phòng tìm thấy từ LAN Discovery.
- **FR-02.3**: **Placement Screen**:
  - Cho phép người chơi di chuyển cursor trên bàn cờ để đặt từng tàu.
  - Phím `R` để xoay hướng tàu.
  - Highlight vùng đặt tàu (Màu xanh nếu hợp lệ, Màu đỏ nếu vi phạm quy tắc).
  - Danh sách Fleet bên cạnh cập nhật trạng thái các tàu đã/chưa đặt.
- **FR-02.4**: **Battle Screen**:
  - Giao diện 2 bàn cờ song song: **Your Fleet** (Bàn mình) & **Enemy Waters** (Bàn địch).
  - Bảng trạng thái lượt chơi (YOUR TURN / ENEMY TURN).
  - Thống kê chi tiết hạm đội và thanh hướng dẫn phím bấm ở footer.
- **FR-02.5**: **Result Screen**: Hiển thị màn hình tổng kết khi kết thúc trận đấu (Victory / Defeat) cùng nút Replay / Exit.
- **FR-02.6**: **Help Overlay / Modal**: Bấm `?` hiển thị bảng tra cứu phím tắt và luật chơi dạng popup overlay.

### FR-03: Networking & Protocol (Kết nối & Giao thức)
- **FR-03.1**: Khởi tạo TCP Server khi Host và TCP Client khi Join.
- **FR-03.2**: Mã hóa và giải mã thông điệp theo định dạng **JSON Lines** (mỗi packet kết thúc bằng ký tự newline `\n`).
- **FR-03.3**: Quản lý gói tin mạng:
  - `HELLO`: Trao đổi thông tin người chơi (Tên, Version).
  - `READY`: Xác nhận đã đặt xong hạm đội.
  - `SHOOT`: Gửi tọa độ bắn `(x, y)`.
  - `SHOT_RESULT`: Phản hồi kết quả của phát bắn (`Hit`, `Miss`, `Sunk`, `Victorious`).
  - `REMATCH`: Đề nghị/Đồng ý đấu lại.
  - `DISCONNECT`: Thông báo thoát game.

### FR-04: LAN Auto-Discovery (Tự động tìm phòng)
- **FR-04.1**: Host gửi gói tin UDP Broadcast theo chu kỳ (Beacon signal) chứa tên Room, Host IP và Port.
- **FR-04.2**: Client ở màn hình Join nghe trên cổng UDP để quét danh sách các Host đang mở trong cùng mạng LAN.

---

## 3. Yêu cầu phi chức năng (Non-Functional Requirements - NFR)

### NFR-01: Performance & Responsiveness
- **NFR-01.1**: Tốc độ phản hồi UI < 16ms (60 FPS smooth rendering trong Terminal).
- **NFR-01.2**: Độ trễ xử lý gói tin mạng qua TCP LAN < 50ms.
- **NFR-01.3**: Không gây nghẽn hoặc giật lag UI khi nhận dữ liệu mạng (Network Read Loop chạy Async độc lập với Bubble Tea Loop).

### NFR-02: Usability & UX
- **NFR-02.1**: Giao diện trực quan, phối màu hài hòa bằng Lip Gloss (Dark mode aesthetic).
- **NFR-02.2**: Hỗ trợ Unicode icons đầy đủ và tự động fallback sang ASCII tượng hình nếu terminal không hỗ trợ Unicode.
- **NFR-02.3**: Xử lý sự kiện co giãn kích thước terminal (`tea.WindowSizeMsg`) linh hoạt không đè vỡ layout.

### NFR-03: Compatibility & Portability
- **NFR-03.1**: Đóng gói thành 1 file binary duy nhất (Single executable file).
- **NFR-03.2**: Chạy tương thích trên Windows PowerShell / CMD / Windows Terminal, Linux Bash, macOS zsh.

### NFR-04: Reliability & Error Handling
- **NFR-04.1**: Xử lý mất kết nối mạng đột ngột (Socket Timeout/Disconnect) và hiển thị thông báo Toast / Modal lỗi cho người chơi.
- **NFR-04.2**: Validate dữ liệu đầu vào nghiêm ngặt, chống crash app khi nhận gói tin sai cấu trúc.

---

## 4. Giới hạn & Giả định (Constraints & Assumptions)

1. **Mạng nội bộ**: Trận đấu diễn ra trong phạm vi cùng subnet LAN hoặc VPN nội bộ (như Tailscale/ZeroTier). Không NAT traversal.
2. **Kích thước Terminal tối thiểu**: Người chơi cần mở terminal kích thước tối thiểu $80 \times 24$ ký tự để hiển thị trọn vẹn Battle Screen.
