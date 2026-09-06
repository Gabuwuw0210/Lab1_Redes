package main

import (
	"bufio"
	"crypto/rand"
	"encoding/base64"
	"encoding/csv"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

const puertoUDP = 9001

type Cliente struct {
	conexion net.Conn
	username string
	token    string
	mutex    sync.Mutex
}

var clientesActivos = make(map[string]*Cliente)
var mutexClientes sync.Mutex
var mutexArchivos sync.Mutex

func verificarCredenciales(username string, password string) bool {
	mutexArchivos.Lock()
	defer mutexArchivos.Unlock()

	archivo, err := os.Open(archivoUsuarios)
	if err != nil {
		return false
	}
	defer archivo.Close()

	lector := csv.NewReader(archivo)
	registros, err := lector.ReadAll()
	if err != nil {
		return false
	}

	for i, registro := range registros {
		if i == 0 {
			continue
		}

		if len(registro) < 2 {
			continue
		}

		if registro[0] == username && registro[1] == password {
			return true
		}
	}

	return false
}

func generarToken() (string, error) {
	bytes := make([]byte, 18)

	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

func guardarSesion(token string, username string) error {
	mutexArchivos.Lock()
	defer mutexArchivos.Unlock()

	archivo, err := os.OpenFile(
		archivoSesiones,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0644,
	)
	if err != nil {
		return err
	}
	defer archivo.Close()

	escritor := csv.NewWriter(archivo)

	ahora := time.Now().Format("2006-01-02 15:04:05")

	registro := []string{
		token,
		username,
		ahora,
		"",
		"ACTIVO",
	}

	if err := escritor.Write(registro); err != nil {
		return err
	}

	escritor.Flush()

	return escritor.Error()
}

func validarSesion(token string) (bool, string) {
	mutexArchivos.Lock()
	defer mutexArchivos.Unlock()

	archivo, err := os.Open(archivoSesiones)
	if err != nil {
		return false, ""
	}
	defer archivo.Close()

	lector := csv.NewReader(archivo)

	registros, err := lector.ReadAll()
	if err != nil {
		return false, ""
	}

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
			return false, ""
		}

		tiempoCreacion, err := time.ParseInLocation(
			"2006-01-02 15:04:05",
			registro[2],
			time.Local,
		)

		if err != nil {
			return false, ""
		}

		if time.Since(tiempoCreacion) >= 10*time.Minute {
			return false, ""
		}

		return true, registro[1]
	}

	return false, ""
}

func invalidarSesion(token string) error {
	mutexArchivos.Lock()
	defer mutexArchivos.Unlock()

	archivo, err := os.Open(archivoSesiones)
	if err != nil {
		return err
	}

	lector := csv.NewReader(archivo)

	registros, err := lector.ReadAll()
	archivo.Close()

	if err != nil {
		return err
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
			registro[4] = "INACTIVO"
			encontrado = true
			break
		}
	}

	if !encontrado {
		return fmt.Errorf("token no encontrado")
	}

	archivoNuevo, err := os.Create(archivoSesiones)
	if err != nil {
		return err
	}
	defer archivoNuevo.Close()

	escritor := csv.NewWriter(archivoNuevo)

	if err := escritor.WriteAll(registros); err != nil {
		return err
	}

	escritor.Flush()

	return escritor.Error()
}

func guardarMensaje(username string, mensaje string) error {
	mutexArchivos.Lock()
	defer mutexArchivos.Unlock()

	archivo, err := os.OpenFile(
		archivoHistorial,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0644,
	)
	if err != nil {
		return err
	}
	defer archivo.Close()

	escritor := csv.NewWriter(archivo)

	registro := []string{
		time.Now().Format("2006-01-02 15:04:05"),
		username,
		mensaje,
	}

	if err := escritor.Write(registro); err != nil {
		return err
	}

	escritor.Flush()

	return escritor.Error()
}

func enviar(cliente *Cliente, mensaje string) error {
	cliente.mutex.Lock()
	defer cliente.mutex.Unlock()

	_, err := cliente.conexion.Write([]byte(mensaje + "\n"))

	return err
}

func manejarCliente(conexion net.Conn) {
	defer conexion.Close()

	lector := bufio.NewReader(conexion)

	// LOGIN
	linea, err := lector.ReadString('\n')
	if err != nil {
		return
	}

	linea = strings.TrimSpace(linea)

	if !strings.HasPrefix(linea, "LOGIN ") {
		fmt.Fprintln(conexion, "ERROR INVALID COMMAND")
		return
	}

	credenciales := strings.TrimPrefix(linea, "LOGIN ")

	// El último espacio separa username de password.
	posicion := strings.LastIndex(credenciales, " ")

	if posicion <= 0 || posicion == len(credenciales)-1 {
		fmt.Fprintln(conexion, "ERROR INVALID COMMAND")
		return
	}

	username := credenciales[:posicion]
	password := credenciales[posicion+1:]

	if !verificarCredenciales(username, password) {
		fmt.Fprintln(conexion, "ERROR INVALID CREDENTIALS")
		return
	}

	token, err := generarToken()
	if err != nil {
		fmt.Fprintln(conexion, "ERROR INTERNAL SERVER")
		return
	}

	if err := guardarSesion(token, username); err != nil {
		fmt.Fprintln(conexion, "ERROR INTERNAL SERVER")
		return
	}

	cliente := &Cliente{
		conexion: conexion,
		username: username,
		token:    token,
	}

	mutexClientes.Lock()
	clientesActivos[token] = cliente
	mutexClientes.Unlock()

	fmt.Println("Usuario autenticado:", username)

	if err := enviar(cliente, fmt.Sprintf("OK %s %d", token, puertoUDP)); err != nil {
		mutexClientes.Lock()
		delete(clientesActivos, token)
		mutexClientes.Unlock()
		return
	}

	for {
		linea, err := lector.ReadString('\n')

		if err != nil {
			mutexClientes.Lock()
			delete(clientesActivos, token)
			mutexClientes.Unlock()
			return
		}

		linea = strings.TrimSpace(linea)

		if linea == "LOGOUT" {
			invalidarSesion(token)

			mutexClientes.Lock()
			delete(clientesActivos, token)
			mutexClientes.Unlock()

			enviar(cliente, "BYE")
			return
		}

		if !strings.HasPrefix(linea, "MSG ") {
			enviar(cliente, "ERROR INVALID COMMAND")
			continue
		}

		partes := strings.SplitN(linea, " ", 3)

		if len(partes) != 3 || partes[2] == "" {
			enviar(cliente, "ERROR INVALID COMMAND")
			continue
		}

		tokenMensaje := partes[1]
		contenido := partes[2]

		if tokenMensaje != token {
			enviar(cliente, "ERROR INVALID TOKEN")
			continue
		}

		sesionValida, usuarioSesion := validarSesion(token)

		if !sesionValida {
			enviar(cliente, "ERROR SESSION EXPIRED")
			continue
		}

		if usuarioSesion != username {
			enviar(cliente, "ERROR INVALID TOKEN")
			continue
		}

		if err := guardarMensaje(username, contenido); err != nil {
			enviar(cliente, "ERROR INTERNAL SERVER")
			continue
		}

		// ACK al emisor.
		enviar(cliente, "ACK")

		// Obtener una copia de los clientes activos.
		mutexClientes.Lock()

		destinatarios := make([]*Cliente, 0, len(clientesActivos))

		for tokenCliente, clienteActivo := range clientesActivos {
			if tokenCliente != token {
				destinatarios = append(destinatarios, clienteActivo)
			}
		}

		mutexClientes.Unlock()

		// Broadcast.
		mensajeBroadcast := fmt.Sprintf(
			"INCOMING %s %s",
			username,
			contenido,
		)

		for _, destinatario := range destinatarios {
			if err := enviar(destinatario, mensajeBroadcast); err != nil {
				fmt.Println(
					"Error enviando mensaje a",
					destinatario.username,
				)
			}
		}
	}
}

func iniciarServidorTCP() {
	go func() {
		servidor, err := net.Listen(
			"tcp",
			"127.0.0.1:9000",
		)

		if err != nil {
			fmt.Println("Error al iniciar servidor TCP:", err)
			return
		}

		defer servidor.Close()

		fmt.Println(
			"Servidor TCP escuchando en 127.0.0.1:9000",
		)

		for {
			conexion, err := servidor.Accept()

			if err != nil {
				continue
			}

			go manejarCliente(conexion)
		}
	}()
}
