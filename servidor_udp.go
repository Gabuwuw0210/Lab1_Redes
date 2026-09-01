package main

import (
	"encoding/csv"
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)

const archivoSesiones = "sesiones.csv"

func actualizarHeartbeat(token string) bool {

	archivo, err := os.Open(archivoSesiones)

	if err != nil {
		fmt.Println("Error al abrir sesiones.csv:", err)
		return false
	}

	defer archivo.Close()

	lector := csv.NewReader(archivo)

	registros, err := lector.ReadAll()

	if err != nil {
		fmt.Println("Error al leer sesiones.csv:", err)
		return false
	}

	encontrado := false

	for i, registro := range registros {

		if i == 0 {
			continue
		}

		if len(registro) < 5 {
			continue
		}

		if registro[0] == token {

			if registro[4] != "ACTIVO" {
				return false
			}

			registro[3] = time.Now().Format("2006-01-02 15:04:05")

			encontrado = true
			break
		}
	}

	if !encontrado {
		return false
	}

	archivoNuevo, err := os.Create(archivoSesiones)

	if err != nil {
		fmt.Println("Error al actualizar sesiones.csv:", err)
		return false
	}

	defer archivoNuevo.Close()

	escritor := csv.NewWriter(archivoNuevo)

	err = escritor.WriteAll(registros)

	if err != nil {
		fmt.Println("Error al escribir sesiones.csv:", err)
		return false
	}

	return true
}

func main() {

	direccion := net.UDPAddr{
		IP:   net.ParseIP("127.0.0.1"),
		Port: 9001,
	}

	servidor, err := net.ListenUDP("udp", &direccion)

	if err != nil {
		fmt.Println("Error al iniciar servidor UDP:", err)
		return
	}

	defer servidor.Close()

	fmt.Println("Servidor UDP escuchando en 127.0.0.1:9001")

	buffer := make([]byte, 1024)

	for {

		n, direccionCliente, err := servidor.ReadFromUDP(buffer)

		if err != nil {
			fmt.Println("Error al recibir UDP:", err)
			continue
		}

		mensaje := strings.TrimSpace(string(buffer[:n]))

		fmt.Println(
			"Mensaje UDP recibido desde",
			direccionCliente,
			":",
			mensaje,
		)

		partes := strings.SplitN(mensaje, " ", 2)

		if len(partes) != 2 || partes[0] != "HEARTBEAT" {
			fmt.Println("Heartbeat invalido")
			continue
		}

		token := partes[1]

		if actualizarHeartbeat(token) {
			fmt.Println("Heartbeat valido. Sesion actualizada:", token)
		} else {
			fmt.Println("Heartbeat rechazado. Token invalido o sesion inactiva:", token)
		}
	}
}
