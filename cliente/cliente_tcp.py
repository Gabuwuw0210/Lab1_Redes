import socket

username = input("Username: ")
password = input("Password: ")

cliente = socket.socket(socket.AF_INET, socket.SOCK_STREAM)

cliente.connect(("127.0.0.1", 9000))

mensaje = f"LOGIN {username} {password}\n"

cliente.sendall(mensaje.encode("utf-8"))

respuesta = cliente.recv(1024).decode("utf-8")

print("Respuesta del servidor:", respuesta.strip())

while True:

    mensaje = input("> ")

    cliente.sendall((mensaje + "\n").encode("utf-8"))

    respuesta = cliente.recv(1024).decode("utf-8")

    print("Servidor:", respuesta.strip())

    if mensaje == "LOGOUT":
        break

cliente.close()
