import socket

cliente = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)

mensaje = "HEARTBEAT kgOAmTRnLDVn3f9d_CyaKXyT"

cliente.sendto(
    mensaje.encode("utf-8"),
    ("127.0.0.1", 9001)
)

print("Heartbeat enviado")

cliente.close()
