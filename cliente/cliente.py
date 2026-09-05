import socket
import threading
import time
import urllib.request
import urllib.parse
import urllib.error

URL = "http://127.0.0.1:8080/register"
PUERTO_TCP = 9000
HOST = "127.0.0.1"

def hilo_heartbeat(token, puerto_udp, logout):
    cliente = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
    mensaje = f"HEARTBEAT {token}"
    try:
        while not logout.is_set():
            cliente.sendto(
                mensaje.encode("utf-8"),
                (HOST, puerto_udp)
            )
            logout.wait(3)
    except Exception as e:
        print(f"\nHilo UDP se ha detenido por error: {e}")
    cliente.close()



def registrar_usuario():
    print("\nREGISTRO")
    username = input("Username: ")
    password = input("Password: ")

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
        print("No se pudo conectar con el servidor:", error.reason)

def hilo_receptor(cliente):
    while True:
        try:
            respuesta=cliente.recv(1024).decode("utf-8")
            if not respuesta:
                print("\n[Desconectado del servidor]")
                break
            mensaje=respuesta.strip()
            if mensaje in ["ACK", "BYE"]:
                continue
            print(f"\n{mensaje}")
            print("> ", end="", flush=True)
                
        except Exception:
            break

def main():
    while True:
        print("\n¿Qué te gustaría hacer?")
        print("1. Registrar nuevo usuario")
        print("2. Iniciar sesión")
        print("3. Salir")
        opcion=input("Elige una opción del 1 al 3: ")
        if opcion=="1":
            registrar_usuario()
        elif opcion=="2":
            username = input("Username: ")
            password = input("Password: ")

            cliente = socket.socket(socket.AF_INET, socket.SOCK_STREAM)

            try:
                cliente.connect(("127.0.0.1", 9000))
            except Exception as e:
                print("Error TCP: ", e)
                continue

            mensaje = f"LOGIN {username} {password}\n"
            cliente.sendall(mensaje.encode("utf-8"))

            respuesta = cliente.recv(1024).decode("utf-8").strip()
            print("Respuesta del servidor:", respuesta)
            partes=respuesta.split(" ")

            if partes[0]=="OK":
                token=partes[1]
                puerto_udp=int(partes[2])

                logout=threading.Event()
                threading.Thread(target=hilo_heartbeat, args=(token, puerto_udp,logout), daemon=True).start()
                threading.Thread(target=hilo_receptor, args=(cliente,), daemon=True).start()
                
                print("\n¡Sesión iniciada! Puedes escribir mensajes. Escribe LOGOUT para salir.")


                while True:

                    mensaje = input("> ")
                    if mensaje == "LOGOUT":
                        cliente.sendall(b"LOGOUT\n")
                        logout.set()
                        break

                    comando = f"MSG {token} {mensaje}\n"
                    try:
                        cliente.sendall(comando.encode("utf-8"))
                    except Exception:
                        break
                cliente.close()
                print("Sesión cerrada.")
            else:
                cliente.close()

        elif opcion=="3":
            break
        else:
            print("Opción no reconocida. Por favor, ingrese opción del 1 al 3: ")



if __name__ == "__main__":
    main()