import socket
import threading
import urllib.request
import urllib.parse
import urllib.error

URL = "http://127.0.0.1:8080/register"
HOST = "127.0.0.1"
PUERTO_TCP = 9000

#HTTP
def registrar_usuario():
    print("\nREGISTRO")

    username = input("Username: ")
    password = input("Password: ")

    if " " in password:
        print("La contraseña no puede contener espacios.")
        return

    datos = {
        "username": username,
        "password": password
    }

    datos_codificados = urllib.parse.urlencode(datos).encode("utf-8")

    solicitud = urllib.request.Request(
        URL,
        data=datos_codificados,
        method="POST"
    )

    solicitud.add_header(
        "Content-Type",
        "application/x-www-form-urlencoded"
    )

    try:
        respuesta = urllib.request.urlopen(solicitud)

        print("Registro exitoso")
        print("Codigo HTTP:", respuesta.status)
        print("Respuesta:", respuesta.read().decode("utf-8"))

    except urllib.error.HTTPError as error:
        print("Error en registro")
        print("Codigo HTTP:", error.code)
        print("Respuesta:", error.read().decode("utf-8"))

    except urllib.error.URLError as error:
        print(
            "No se pudo conectar con el servidor:",
            error.reason
        )

#UDP
def hilo_heartbeat(token, puerto_udp, detener):
    cliente_udp = socket.socket(
        socket.AF_INET,
        socket.SOCK_DGRAM
    )

    mensaje = f"HEARTBEAT {token}".encode("utf-8")

    try:
        while not detener.is_set():
            cliente_udp.sendto(
                mensaje,
                (HOST, puerto_udp)
            )

            detener.wait(3)

    finally:
        cliente_udp.close()

#TCP
def hilo_receptor(cliente, detener):
    lector = cliente.makefile(
        "r",
        encoding="utf-8"
    )

    try:
        while not detener.is_set():
            linea = lector.readline()

            if not linea:
                print("\n[Servidor desconectado]")
                detener.set()
                break

            linea = linea.rstrip("\n")

            if linea == "ACK":
                continue

            if linea == "BYE":
                detener.set()
                break

            print(f"\n{linea}")
            print("> ", end="", flush=True)

    except OSError:
        pass

    finally:
        lector.close()

def iniciar_sesion():
    username = input("Username: ")
    password = input("Password: ")

    cliente = socket.socket(
        socket.AF_INET,
        socket.SOCK_STREAM
    )

    try:
        cliente.connect(
            (HOST, PUERTO_TCP)
        )
    except OSError as error:
        print("Error TCP:", error)
        return

    # El protocolo permite espacios en username,
    # pero la contraseña no debe contener espacios.
    if " " in password:
        print(
            "La contraseña no puede contener espacios "
            "con el protocolo actual."
        )
        cliente.close()
        return

    mensaje = f"LOGIN {username} {password}\n"

    try:
        cliente.sendall(
            mensaje.encode("utf-8")
        )

        respuesta = cliente.recv(1024).decode("utf-8").strip()

    except OSError as error:
        print("Error de comunicación:", error)
        cliente.close()
        return

    print("Respuesta del servidor:", respuesta)

    partes = respuesta.split()

    if len(partes) != 3 or partes[0] != "OK":
        cliente.close()
        return

    token = partes[1]
    puerto_udp = int(partes[2])

    detener = threading.Event()

    hilo_udp = threading.Thread(
        target=hilo_heartbeat,
        args=(token, puerto_udp, detener),
        daemon=True
    )

    hilo_tcp = threading.Thread(
        target=hilo_receptor,
        args=(cliente, detener),
        daemon=True
    )

    hilo_udp.start()
    hilo_tcp.start()

    print(
        "\n¡Sesión iniciada! "
        "Puedes escribir mensajes."
    )
    print("Escribe LOGOUT para salir.")

    try:
        while not detener.is_set():
            mensaje = input("> ")

            if detener.is_set():
                break

            if mensaje == "":
                continue

            if mensaje == "LOGOUT":
                cliente.sendall(
                    b"LOGOUT\n"
                )
                break

            comando = f"MSG {token} {mensaje}\n"

            try:
                cliente.sendall(
                    comando.encode("utf-8")
                )
            except OSError:
                break

    except (EOFError, KeyboardInterrupt):
        try:
            cliente.sendall(b"LOGOUT\n")
        except OSError:
            pass

    finally:
        detener.set()
        cliente.close()

    print("Sesión cerrada.")

def main():
    while True:
        print("\n¿Qué te gustaría hacer?")
        print("1. Registrar nuevo usuario")
        print("2. Iniciar sesión")
        print("3. Salir")

        opcion = input(
            "Elige una opción del 1 al 3: "
        )

        if opcion == "1":
            registrar_usuario()

        elif opcion == "2":
            iniciar_sesion()

        elif opcion == "3":
            break

        else:
            print(
                "Opción no reconocida."
            )

if __name__ == "__main__":
    main()
