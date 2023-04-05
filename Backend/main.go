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

type ComandoJson struct {
	Comando string
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

func index(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		r := make([]string, 0)
		reportes, er := os.ReadDir("Reportes/")
		if er != nil {
			fmt.Println(er)
		}
		for _, reporte := range reportes {
			i := find(reporte.Name(), ".")
			if i < len(reporte.Name()) {
				r = append(r, reporte.Name())
			}

		}
		json.NewEncoder(w).Encode(r)
	} else if r.Method == "POST" {
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
	}
}

func main() {
	listaMount = &estructuras.ListaMount{}
	usuario = &estructuras.Usuario{}

	server := mux.NewRouter()
	server.HandleFunc("/", index)

	handler := cors.Default().Handler(server)

	server.PathPrefix("/Reportes/").Handler(http.StripPrefix("/Reportes/", http.FileServer(http.Dir("./Reportes/"))))

	http.ListenAndServe(":8080", handler)
}
