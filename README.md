## 1. Integrantes del Grupo
Catalina Díaz - 202473507-1

## 2. Instrucciones de Ejecución y Compilación

### Puertos por defecto
*   **Servicio HTTP (Registro):** 8080
*   **Servicio TCP (Autenticación y Chat):** 9000
*   **Servicio UDP (Heartbeat):** 9001

### Levantar el Servidor (Go)
Para compilar y ejecutar el servidor, abra una terminal en el directorio `servidor/` y ejecute:
1. `go mod init servidor` (Solo la primera vez para inicializar el módulo).
2. `go run .`

### Ejecutar el Cliente (Python)
Abra una nueva terminal en el directorio `cliente/` y ejecute:
1. `python cliente.py` o `python3 cliente.py`.
2. Siga el menú interactivo para registrar un usuario (HTTP) o iniciar sesión (TCP/UDP).

## 3. Documentación de Protocolos y Comandos

### Protocolo HTTP (Registro)
*   **Endpoint:** `POST /register`
*   **Descripción:** Recibe credenciales codificadas (`username` y `password`) y almacena el registro en `usuarios.csv`.

### Protocolo TCP (Autenticación y Mensajería)
El canal TCP utiliza cadenas de texto delimitadas por un salto de línea (`\n`).

*   **Comando de Login (Cliente -> Servidor):** 
    `LOGIN <username> <password>\n`
*   **Respuesta de Login Exitoso (Servidor -> Cliente):** 
    `OK <token> <puerto_udp>\n`
*   **Respuesta de Error de Login (Servidor -> Cliente):** 
    `ERROR INVALID_CREDENTIALS\n`
*   **Comando de Envío de Mensaje (Cliente -> Servidor):** 
    `MSG <token> <contenido_del_mensaje>\n`
*   **Acuse de Recibo (Servidor -> Cliente remitente):** 
    `ACK\n`
*   **Retransmisión de Mensaje / Broadcast (Servidor -> Todos los clientes):** 
    `INCOMING <usuario_emisor> <contenido_del_mensaje>\n`

### Protocolo UDP (Presencia / Heartbeat)
*   **Comando de Latido (Cliente -> Servidor):** 
    `HEARTBEAT <token>`
*   **Descripción:** Se envía un datagrama cada 3 segundos. Si el servidor no recibe este comando en 60 segundos (o en los primeros 30 segundos), la sesión se revoca y se cierra el socket TCP del usuario.
