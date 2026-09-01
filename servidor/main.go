package main

const (
	archivoUsuarios  = "usuarios.csv"
	archivoSesiones  = "sesiones.csv"
	archivoHistorial = "historial.csv"
)

func main() {
	iniciarServidorHTTP()
	iniciarServidorTCP()
	iniciarServidorUDP()

	select {}
}
