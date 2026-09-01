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
}

var clientesActivos = make(map[string]*Cliente)
var mutexClientes sync.Mutex

func verificarCredenciales(username string, password string) bool {

	archivo, err := os.Open(archivoUsuarios)

	if err != nil {
		fmt.Println("Error al abrir usuarios.csv:", err)
		return false
	}

	defer archivo.Close()

	lector := csv.NewReader(archivo)

	registros, err := lector.ReadAll()

	if err != nil {
		fmt.Println("Error al leer usuarios.csv:", err)
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

	err = escritor.Write(registro)

	escritor.Flush()

	return err
}

func validarSesion(token string) (bool, string) {

	archivo, err := os.Open(archivoSesiones)

	if err != nil {
		fmt.Println("Error al abrir sesiones.csv:", err)
		return false, ""
	}

	defer archivo.Close()

	lector := csv.NewReader(archivo)

	registros, err := lector.ReadAll()

	if err != nil {
		fmt.Println("Error al leer sesiones.csv:", err)
		return false, ""
	}

	for i, registro := range registros {

		if i == 0 {
			continue
		}

		if len(registro) < 5 {
			continue
		}

		tokenCSV := registro[0]
		username := registro[1]
		timestampCreacion := registro[2]
		estado := registro[4]

		// Buscar el token
		if tokenCSV != token {
			continue
		}

		fmt.Println("Token encontrado en sesiones.csv")
		fmt.Println("Usuario:", username)
		fmt.Println("Estado:", estado)
		fmt.Println("Creacion:", timestampCreacion)

		// Comprobar estado
		if estado != "ACTIVO" {
			fmt.Println("Sesion no esta ACTIVA")
			return false, ""
		}

		// Convertir timestamp
		tiempoCreacion, err := time.ParseInLocation(
			"2006-01-02 15:04:05",
			timestampCreacion,
			time.Local,
		)

		if err != nil {
			fmt.Println("Error al interpretar timestamp:", err)
			return false, ""
		}

		// Calcular antiguedad
		antiguedad := time.Since(tiempoCreacion)

		fmt.Println("Antiguedad de la sesion:", antiguedad)

		// TTL de 10 minutos
		if antiguedad > 10*time.Minute {

			fmt.Println("Sesion expirada por TTL")

			return false, ""
		}

		fmt.Println("Sesion valida")

		return true, username
	}

	fmt.Println("Token no encontrado en sesiones.csv")

	return false, ""
}

func invalidarSesion(token string) error {

	archivo, err := os.Open(archivoSesiones)

	if err != nil {
		return err
	}

	defer archivo.Close()

	lector := csv.NewReader(archivo)

	registros, err := lector.ReadAll()

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

	err = escritor.WriteAll(registros)

	if err != nil {
		return err
	}

	escritor.Flush()

	return nil
}

func manejarCliente(conexion net.Conn) {
	defer conexion.Close()
	fmt.Println("Cliente TCP conectado")
	lector := bufio.NewReader(conexion)

	// LOGIN
	linea, err := lector.ReadString('\n')

	if err != nil {
		fmt.Println("Error al recibir LOGIN:", err)
		return
	}

	linea = strings.TrimSpace(linea)
	fmt.Println("Mensaje recibido:", linea)
	partes := strings.SplitN(linea, " ", 3)

	if len(partes) != 3 || partes[0] != "LOGIN" {

		fmt.Fprintln(conexion, "ERROR INVALID COMMAND")
		return
	}

	username := partes[1]
	password := partes[2]

	if !verificarCredenciales(username, password) {
		fmt.Fprintln(conexion, "ERROR INVALID CREDENTIALS")
		return
	}

	token, err := generarToken()

	if err != nil {
		fmt.Println("Error al generar token:", err)
		fmt.Fprintln(conexion, "ERROR INTERNAL SERVER")
		return
	}

	err = guardarSesion(token, username)

	if err != nil {
		fmt.Println("Error al guardar sesión:", err)
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
	fmt.Println("Token:", token)
	fmt.Println("Puerto UDP:", puertoUDP)

	fmt.Fprintf(
		conexion,
		"OK %s %d\n",
		token,
		puertoUDP,
	)

	// CONEXIÓN PERMANENTE
	for {
		linea, err = lector.ReadString('\n')

		if err != nil {
			fmt.Println(
				"Cliente desconectado:",
				username,
			)

			mutexClientes.Lock()
			delete(clientesActivos, token)
			mutexClientes.Unlock()

			return
		}

		linea = strings.TrimSpace(linea)

		fmt.Println(
			"Mensaje recibido de",
			username,
			":",
			linea,
		)

		if linea == "LOGOUT" {
			err := invalidarSesion(token)

			if err != nil {
				fmt.Println(
					"Error al invalidar sesion:",
					err,
				)

				fmt.Fprintln(
					conexion,
					"ERROR INTERNAL SERVER",
				)

				return
			}

			mutexClientes.Lock()
			delete(clientesActivos, token)
			mutexClientes.Unlock()

			fmt.Println(
				"Sesion invalidada:",
				token,
			)

			fmt.Fprintln(
				conexion,
				"BYE",
			)

			return
		}

		partes = strings.SplitN(linea, " ", 3)

		if len(partes) == 3 && partes[0] == "MSG" {
			tokenMensaje := partes[1]
			contenido := partes[2]

			// Validar token
			if tokenMensaje != token {
				fmt.Fprintln(
					conexion,
					"ERROR INVALID TOKEN",
				)

				continue
			}

			// Validar sesión
			sesionValida, usuarioSesion := validarSesion(tokenMensaje)

			if !sesionValida {
				fmt.Fprintln(
					conexion,
					"ERROR SESSION EXPIRED",
				)

				continue
			}

			// Validar usuario-token
			if usuarioSesion != username {
				fmt.Fprintln(
					conexion,
					"ERROR INVALID TOKEN",
				)

				continue
			}

			// Guardar mensaje.
			err := guardarMensaje(
				username,
				contenido,
			)

			if err != nil {
				fmt.Println(
					"Error al guardar mensaje:",
					err,
				)

				fmt.Fprintln(
					conexion,
					"ERROR INTERNAL SERVER",
				)

				continue
			}

			fmt.Println(
				"Mensaje guardado:",
				username,
				"-",
				contenido,
			)

			// ACK al remitente
			fmt.Fprintln(
				conexion,
				"ACK",
			)

			// Mensaje para los demás clientes
			broadcast := fmt.Sprintf(
				"INCOMING %s %s\n",
				username,
				contenido,
			)

			mutexClientes.Lock()

			for tokenCliente, cliente := range clientesActivos {

				// No enviar el mensaje
				// nuevamente al remitente
				if tokenCliente == token {
					continue
				}

				_, err := cliente.conexion.Write(
					[]byte(broadcast),
				)

				if err != nil {

					fmt.Println(
						"Error enviando broadcast a",
						cliente.username,
					)
				}
			}

			mutexClientes.Unlock()

			continue
		}

		fmt.Fprintln(conexion, "ERROR INVALID COMMAND")
	}
}

func guardarMensaje(username string, mensaje string) error {

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

	timestamp := time.Now().Format("2006-01-02 15:04:05")

	registro := []string{
		timestamp,
		username,
		mensaje,
	}

	err = escritor.Write(registro)

	escritor.Flush()

	return err
}

func iniciarServidorTCP() {

	go func() {

		servidor, err := net.Listen(
			"tcp",
			"127.0.0.1:9000",
		)

		if err != nil {

			fmt.Println(
				"Error al iniciar servidor TCP:",
				err,
			)

			return
		}

		defer servidor.Close()

		fmt.Println(
			"Servidor TCP escuchando en 127.0.0.1:9000",
		)

		for {

			conexion, err := servidor.Accept()

			if err != nil {

				fmt.Println(
					"Error al aceptar conexión:",
					err,
				)

				continue
			}

			go manejarCliente(conexion)
		}
	}()
}
