package estructuras

type NodoMount struct {
	Nombre_disco     string
	Numero           int
	Part_start       int
	Part_type        byte
	Fichero          string
	Id               string
	IdCompleto       string
	Next             *NodoMount
	Nombre_particion string
}
