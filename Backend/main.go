package main

import (
	"encoding/json"
	"net/http"
	"strings"

	analizador "github.com/EstuardoSon/P2_MIA_202003894/Analizador"
	"github.com/rs/cors"
)

type Respuesta struct {
	Res string
}

type ComandoJson struct {
	Comando string
}

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
					a := &analizador.Analizador{Comando: strings.TrimSpace(comando)}
					resultado += a.Analizar() + "\n\n"
				}
			}
			res := Respuesta{Res: resultado}
			json.NewEncoder(w).Encode(res)
		}
		return
	}
}

func initRoutes(mux *http.ServeMux) {
	(*mux).HandleFunc("/", index)
}

func New() *http.ServeMux {
	mux := http.NewServeMux()
	initRoutes(mux)

	return mux
}

func main() {
	mux := New()

	handler := cors.Default().Handler(mux)
	http.ListenAndServe(":8080", handler)
}
