import urllib.request
import urllib.parse

url = "http://127.0.0.1:8080/register"

username = input("Username: ")
password = input("Password: ")

datos = {
    "username": username,
    "password": password
}

datos_codificados = urllib.parse.urlencode(datos).encode("utf-8")

solicitud = urllib.request.Request(
    url,
    data=datos_codificados,
    method="POST"
)

solicitud.add_header(
    "Content-Type",
    "application/x-www-form-urlencoded"
)

try:
    respuesta = urllib.request.urlopen(solicitud)

    print("Codigo HTTP:", respuesta.status)
    print("Respuesta:", respuesta.read().decode("utf-8"))

except urllib.error.HTTPError as error:

    print("Codigo HTTP:", error.code)
    print("Respuesta:", error.read().decode("utf-8"))

except urllib.error.URLError as error:

    print("No se pudo conectar con el servidor:", error.reason)