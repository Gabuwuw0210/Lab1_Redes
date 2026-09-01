package main

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"os"
	"time"
)

const archivoUsuarios = "usuarios.csv"

func usuarioExiste(username string) bool {

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

	for _, registro := range registros {

		if len(registro) >= 1 && registro[0] == username {
			return true
		}
	}

	return false
}

func guardarUsuario(username string, password string) error {

	archivo, err := os.OpenFile(
		archivoUsuarios,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0644,
	)

	if err != nil {
		return err
	}

	defer archivo.Close()

	escritor := csv.NewWriter(archivo)
	defer escritor.Flush()

	fecha := time.Now().Format("2006-01-02 15:04:05")

	registro := []string{
		username,
		password,
		fecha,
	}

	return escritor.Write(registro)
}

func registerHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseForm()

	if err != nil {
		http.Error(w, "Solicitud invalida", http.StatusBadRequest)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	if username == "" || password == "" {
		http.Error(w, "Faltan username o password", http.StatusBadRequest)
		return
	}

	if usuarioExiste(username) {
		http.Error(w, "El usuario ya existe", http.StatusConflict)
		return
	}

	err = guardarUsuario(username, password)

	if err != nil {
		http.Error(w, "Error al guardar usuario", http.StatusInternalServerError)
		return
	}

	fmt.Println("Nuevo usuario registrado:", username)

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintln(w, "Usuario registrado correctamente")
}

func historyHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		http.Error(
			w,
			"Metodo no permitido",
			http.StatusMethodNotAllowed,
		)
		return
	}

	archivo, err := os.ReadFile("historial.csv")

	if err != nil {

		http.Error(
			w,
			"Error al leer historial",
			http.StatusInternalServerError,
		)

		return
	}

	w.Header().Set(
		"Content-Type",
		"text/csv; charset=utf-8",
	)

	w.WriteHeader(http.StatusOK)

	w.Write(archivo)
}

func main() {

	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/history", historyHandler)

	fmt.Println("Servidor HTTP escuchando en 127.0.0.1:8080")

	err := http.ListenAndServe("127.0.0.1:8080", nil)

	if err != nil {
		fmt.Println("Error:", err)
	}
}
