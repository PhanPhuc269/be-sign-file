# Tổng kết xử lý chức năng ký tài liệu

1. **Luồng ký tài liệu**
   - Khi người dùng yêu cầu ký tài liệu, backend kiểm tra sự tồn tại của tài liệu và người ký.
   - Sinh khóa RSA (private key) tạm thời cho mỗi lần ký (demo), thực tế nên lưu trữ an toàn.
   - Tính digest (băm SHA256) của nội dung file gốc.
   - Ký digest bằng private key, lưu chữ ký (base64) và thông tin thuật toán vào DB.
   - Lưu private key (base64) vào DB (chỉ dùng cho demo, không dùng cho production).
   - Tạo file đã ký: nối nội dung file gốc với marker `---BEGIN SIGNATURE---` và chữ ký, đảm bảo phần trước marker giống 100% file gốc.

2. **Xác minh chữ ký tài liệu**
   - Khi upload file đã ký để xác minh, backend tách phần nội dung gốc và phần chữ ký dựa vào marker.
   - Tính lại digest của phần nội dung gốc, so sánh với digest đã lưu trong DB.
   - Nếu digest khớp, giải mã chữ ký và xác minh bằng public key (lấy từ private key demo).
   - Trả về kết quả xác minh: hợp lệ hoặc không hợp lệ, kèm thông báo chi tiết.
