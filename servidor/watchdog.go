package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"time"
)

func iniciarWatchdog(){
	go func(){
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for range ticker.C{
			archivo, err:=os.Open(archivoSesiones)
			if err!=nil{
				continue
			}

			lector:=csv.NewReader(archivo)
			registros, err:=lector.ReadAll()
			archivo.Close()

			if err != nil{
				fmt.Println("Error al leer sesiones.csv:", err)
				continue
			}
			
			ahora := time.Now()

			for i, registro:=range registros{
				if i==0{
					continue
				}
				if len(registro)<5{
					continue
				}
				
				token:=registro[0]
				timestampCreacion:=registro[2]
				timestampHeartbeat:=registro[3]
				estado:=registro[4]

				if estado!="ACTIVO"{
					continue
				}
				
				tiempoCreacion, err:=time.ParseInLocation(
					"2006-01-02 15:04:05",
					timestampCreacion,
					time.Local,
				)
				
				if err!=nil{
					fmt.Println("Error al interpretar timestamp:", err)
					continue
				}

				revocar:=false
				motivo:=""

				if ahora.Sub(tiempoCreacion)>=10*time.Minute{
					revocar=true
					motivo="TTL de 10 minutos expirado"
				} else if timestampHeartbeat=="" && ahora.Sub(tiempoCreacion)>=30*time.Second{
					revocar=true
					motivo="Sin primer datagrama tras 30 segundos"
				} else if timestampHeartbeat!=""{
					tiempoHeartbeat, err:=time.ParseInLocation(
						"2006-01-02 15:04:05",
						timestampHeartbeat,
						time.Local,
					)
					
					if err!=nil{
						fmt.Println("Error al interpretar timestamp heartbeat:", err)
						continue 
					}
					
					if ahora.Sub(tiempoHeartbeat)>60*time.Second{
						revocar=true
						motivo="Sin heartbeat por más de 60 segundos"
					}
				}

				if revocar{
					fmt.Printf("Revocando sesion %s. Motivo: %s\n", token, motivo)
					invalidarSesion(token)
					mutexClientes.Lock()
					if cliente, existe:=clientesActivos[token]; existe{
						fmt.Fprintf(cliente.conexion, "ERROR SESSION EXPIRED\n")
						cliente.conexion.Close()
						delete(clientesActivos, token)
						fmt.Println("Socket TCP cerrado para el token", token)
					}
					mutexClientes.Unlock()
				}
			}
		}
	}()
}