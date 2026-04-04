# NamespaceManager
Resource Allocation and Management in Multi-Cluster project

cmd/api	เป็นจุดเริ่มต้นของแอปพลิเคชัน ใช้สำหรับรัน API server และงานเบื้องหลังของระบบ
config	ใช้จัดการค่าคอนฟิกต่าง ๆ ของระบบ เช่น OAuth, session และการตั้งค่าที่เกี่ยวข้อง
docs	รวบรวมเอกสารของโปรเจกต์ เช่น Swagger, ER Diagram และ Sequence Diagram
internal/app	ทำหน้าที่ bootstrap ระบบ เช่น ผูก dependency, กำหนด route และเชื่อมต่อฐานข้อมูล
internal/middleware	เก็บ middleware ส่วนกลาง เช่น การตรวจสอบสิทธิ์ผู้ใช้และ role
internal/models	กำหนดโครงสร้างข้อมูลหลัก (entity/model) และการ mapping กับฐานข้อมูล
internal/auth	ดูแลระบบยืนยันตัวตนและการจัดการโทเค็น เช่น JWT, login/register และ blacklist
internal/admin	รวมฟีเจอร์สำหรับผู้ดูแลระบบ เช่น การจัดการสิทธิ์ super admin