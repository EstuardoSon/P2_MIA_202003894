package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	analizador "github.com/EstuardoSon/P2_MIA_202003894/Analizador"
	estructuras "github.com/EstuardoSon/P2_MIA_202003894/Estructuras"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

type Respuesta struct {
	Res string
}

type RespuestaRe struct {
	Res []string
}

type ComandoJson struct {
	Comando string
}

type Login struct {
	Id   string
	User string
	Pass string
}

type Archivo struct {
	Uri string
}

type ResLogin struct {
	Archivos []Archivo
	Res      bool
}

var listaMount *estructuras.ListaMount
var usuario *estructuras.Usuario

func find(cadena string, substring string) int {
	i := strings.Index(cadena, substring)
	if i == -1 {
		i = len(cadena)
	}
	return i
}

func login(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		operacion := &Login{}
		err := json.NewDecoder(r.Body).Decode(operacion)
		if err != nil {
			fmt.Println(err)
		}

		log := estructuras.AdminUsuario{ListaMount: listaMount, Usuario: usuario}

		ra := make([]Archivo, 0)
		if log.LoginEspecial(operacion.User, operacion.Pass, operacion.Id) == "Sesion Iniciada" {
			reportes, er := os.ReadDir("Reportes/")
			if er != nil {
				fmt.Println(er)
			}
			for _, reporte := range reportes {
				i := find(reporte.Name(), ".")
				if i < len(reporte.Name()) {
					if strings.Index(reporte.Name(), operacion.Id) != -1 {
						ra = append(ra, Archivo{Uri: "http://localhost:8080/Reportes/" + reporte.Name()})
					}
				}

			}
			res := ResLogin{Res: true, Archivos: ra}
			json.NewEncoder(w).Encode(res)
		} else {
			res := ResLogin{Res: false, Archivos: ra}
			json.NewEncoder(w).Encode(res)
		}
	}
}

func index(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		operacion := &ComandoJson{}
		err := json.NewDecoder(r.Body).Decode(operacion)

		if err != nil {
			panic(err)
		} else {
			w.Header().Set("Content-Type", "application/json")
			Comandos := strings.Split(operacion.Comando, "\n")
			var resultado string

			for _, comando := range Comandos {
				if strings.TrimSpace(comando) != "" {
					a := &analizador.Analizador{Comando: strings.TrimSpace(comando), ListaMount: listaMount, Usuario: usuario}
					resultado += a.Analizar() + "\n\n"
				}
			}
			res := Respuesta{Res: resultado}
			json.NewEncoder(w).Encode(res)
		}
		return
	} else if r.Method == "GET" {
		json.NewEncoder(w).Encode("Estuardo Gabriel Son Mux 202003894")
	}
}

func main() {
	listaMount = &estructuras.ListaMount{}
	usuario = &estructuras.Usuario{}

	server := mux.NewRouter()
	server.HandleFunc("/", index)
	server.HandleFunc("/Login", login)

	handler := cors.Default().Handler(server)

	server.PathPrefix("/Reportes/").Handler(http.StripPrefix("/Reportes/", http.FileServer(http.Dir("./Reportes/"))))

	fmt.Println("Servidor Ejecutandose en el Puerto 8080")
	http.ListenAndServe(":8080", handler)
}
