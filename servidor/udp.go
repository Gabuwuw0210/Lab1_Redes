package main

import (
	"encoding/csv"
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)

const (
	formatoFecha = "2006-01-02 15:04:05"
)

func actualizarHeartbeat(token string) bool {
	mutexArchivos.Lock()
	defer mutexArchivos.Unlock()

	archivo, err := os.Open(archivoSesiones)
	if err != nil {
		return false
	}

	lector := csv.NewReader(archivo)
	registros, err := lector.ReadAll()
	archivo.Close()

	if err != nil {
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

		if registro[0] != token {
			continue
		}

		if registro[4] != "ACTIVO" {
			return false
		}

		registro[3] = time.Now().Format(formatoFecha)
		encontrado = true
		break
	}

	if !encontrado {
		return false
	}

	archivoNuevo, err := os.Create(archivoSesiones)
	if err != nil {
		return false
	}

	defer archivoNuevo.Close()
	escritor := csv.NewWriter(archivoNuevo)

	if err := escritor.WriteAll(registros); err != nil {
		return false
	}

	escritor.Flush()
	return escritor.Error() == nil
}

func marcarSesionInactiva(token string) bool {
	mutexArchivos.Lock()
	defer mutexArchivos.Unlock()

	archivo, err := os.Open(archivoSesiones)
	if err != nil {
		return false
	}

	lector := csv.NewReader(archivo)
	registros, err := lector.ReadAll()
	archivo.Close()

	if err != nil {
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

		if registro[0] != token {
			continue
		}

		if registro[4] != "ACTIVO" {
			return false
		}

		registro[4] = "INACTIVO"
		encontrado = true
		break
	}

	if !encontrado {
		return false
	}

	archivoNuevo, err := os.Create(archivoSesiones)
	if err != nil {
		return false
	}
	defer archivoNuevo.Close()

	escritor := csv.NewWriter(archivoNuevo)

	if err := escritor.WriteAll(registros); err != nil {
		return false
	}

	escritor.Flush()

	return escritor.Error() == nil
}

func iniciarServidorUDP() {
	go func() {

		direccion := net.UDPAddr{
			IP:   net.ParseIP("127.0.0.1"),
			Port: puertoUDP,
		}

		servidor, err := net.ListenUDP("udp", &direccion)

		if err != nil {
			fmt.Println("Error al iniciar servidor UDP:", err)
			return
		}

		defer servidor.Close()

		fmt.Println(
			"Servidor UDP escuchando en 127.0.0.1:9001",
		)

		buffer := make([]byte, 1024)

		for {
			n, direccionCliente, err := servidor.ReadFromUDP(buffer)

			if err != nil {
				fmt.Println("Error al recibir UDP:", err)
				continue
			}

			mensaje := strings.TrimSpace(
				string(buffer[:n]),
			)

			partes := strings.SplitN(mensaje, " ", 2)

			if len(partes) != 2 || partes[0] != "HEARTBEAT" {
				fmt.Println("Heartbeat invalido")
				continue
			}

			token := partes[1]

			if actualizarHeartbeat(token) {
				fmt.Printf(
					"HB de %s valido: %s\n",
					direccionCliente.IP.String(),
					token,
				)
			} else {
				fmt.Println(
					"Heartbeat rechazado:",
					token,
				)
			}
		}
	}()
}
