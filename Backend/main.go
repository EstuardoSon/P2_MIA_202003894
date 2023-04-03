package main

import (
	"encoding/json"
	"net/http"
	"strings"

	analizador "github.com/EstuardoSon/P2_MIA_202003894/Analizador"
	estructuras "github.com/EstuardoSon/P2_MIA_202003894/Estructuras"
	"github.com/rs/cors"
)

type Respuesta struct {
	Res string
}

type ComandoJson struct {
	Comando string
}

var listaMount *estructuras.ListaMount

func index(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ComandoJson{Comando: "Hola Mundo"})
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
					a := &analizador.Analizador{Comando: strings.TrimSpace(comando), ListaMount: listaMount}
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

	mux := http.NewServeMux()
	(*mux).HandleFunc("/", index)

	handler := cors.Default().Handler(mux)
	http.ListenAndServe(":8080", handler)
}
