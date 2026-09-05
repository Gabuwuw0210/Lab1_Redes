package main

import (
	"os"
	"path/filepath"
)

var (
	archivoUsuarios  string
	archivoSesiones  string
	archivoHistorial string
)

func configurarArchivos() {

	directorio, err := os.Getwd()

	if err != nil {
		panic(err)
	}

	if filepath.Base(directorio) == "servidor" {
		directorio = filepath.Dir(directorio)
	}

	archivoUsuarios = filepath.Join(directorio, "usuarios.csv")
	archivoSesiones = filepath.Join(directorio, "sesiones.csv")
	archivoHistorial = filepath.Join(directorio, "historial.csv")
}

func main() {

	configurarArchivos()

	iniciarServidorHTTP()
	iniciarServidorTCP()
	iniciarServidorUDP()
	iniciarWatchdog()

	select {}
}
