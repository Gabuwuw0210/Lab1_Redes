# Lab1_Redes
este readme es para nosotros, para levantar el programa, primero asegura tener py3 y go instalados

0. Explicacion de los csv: `historial` guarda el historial de mensajes, `usuarios` son los perfiles registrados en el sistema, `sesiones` son la lista de tokens de usuario tanto activas como inactivas

   Pasos a seguir:
1. abre 2 powershell donde tengas la carpeta Lab1_Redes descargada
2. en la powershell 1, prende el servidor escribiendo `go mod init lab1_redes` para crear una wea parecida a un makefile y luego utiliza `go run ./servidor` para correrlo
3. ahora para la parte del cliente usamos la powershell 2:


- si quieres REGISTRAR un cliente en usuarios.csv, utiliza `cliente_http.py` escribiendo `python cliente_http.py`
- si quieres hacer LOGIN para acceder y luego mandar MSG, utiliza `cliente_tcp.py` escribiendo `python cliente_tcp.py`
- si quieres mandar un heartbeat al servidor, utliza `cliente_udp.py` escribiendo `python cliente_udp.py` (esto actualiza el valor de `timestamp_ultimo_heartbeat` en `sesiones.csv`)

falta hacer el componente 3 y 4
