package analizador

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	estructuras "github.com/EstuardoSon/P2_MIA_202003894/Estructuras"
)

type error interface {
	Error() string
}

type Analizador struct {
	Comando    string
	ListaMount *estructuras.ListaMount
	Usuario    *estructuras.Usuario
}

func (analizador *Analizador) recoInstrucion(cadena string) (int, error) {
	if len(cadena) >= 6 && cadena[:6] == "mkdisk" {
		return 1, nil
	} else if len(cadena) >= 6 && cadena[:6] == "rmdisk" {
		return 2, nil
	} else if len(cadena) >= 5 && cadena[:5] == "fdisk" {
		return 3, nil
	} else if len(cadena) >= 5 && cadena[:5] == "mount" {
		return 4, nil
	} else if len(cadena) >= 4 && cadena[:4] == "mkfs" {
		return 5, nil
	} else if len(cadena) >= 5 && cadena[:5] == "login" {
		return 6, nil
	} else if len(cadena) >= 6 && cadena[:6] == "logout" {
		return 7, nil
	} else if len(cadena) >= 5 && cadena[:5] == "mkgrp" {
		return 8, nil
	} else if len(cadena) >= 5 && cadena[:5] == "rmgrp" {
		return 9, nil
	} else if len(cadena) >= 5 && cadena[:5] == "mkusr" {
		return 10, nil
	} else if len(cadena) >= 5 && cadena[:5] == "rmusr" {
		return 11, nil
	} else if len(cadena) >= 6 && cadena[:6] == "mkfile" {
		return 12, nil
	} else if len(cadena) >= 5 && cadena[:5] == "mkdir" {
		return 13, nil
	} else if len(cadena) >= 5 && cadena[:5] == "pause" {
		return 14, nil
	} else if len(cadena) >= 3 && cadena[:3] == "rep" {
		return 15, nil
	}
	return -1, fmt.Errorf("Comando no reconocido")
}

// Obtener la posicion del siguiente espacio en blanco
func obtenerPosEspacio(cadena string) int {
	for x := 0; x < len(cadena); x++ {
		if cadena[x] == ' ' || cadena[x] == '\n' || cadena[x] == '\t' || cadena[x] == '\r' {
			return x
		}
	}
	return len(cadena)
}

// Verificar si la cadena es un comentario
func (analizador *Analizador) verificarComentario(cadena string) bool {
	if cadena[0] == '#' {
		fmt.Println(cadena[:])
		return true
	}
	return false
}

// Obtener la posicion de la ultima "
func (analizador *Analizador) obtenerPosEnd(cadena string) int {
	for x := 0; x < len(cadena); x++ {
		if cadena[x] == '"' {
			return x
		}
	}
	return len(cadena)
}

// Obtener un parametro que puede estar entre comillas o sin comillas
func (analizador *Analizador) obtenerDatoParamC(parametro *string, tamanio int) {
	analizador.Comando = strings.TrimSpace(analizador.Comando[tamanio:])

	if analizador.Comando[0] == '"' {
		analizador.Comando = analizador.Comando[1:]
		posComilla := analizador.obtenerPosEnd(analizador.Comando)
		*parametro = analizador.Comando[:posComilla]
		analizador.Comando = strings.TrimSpace(analizador.Comando[posComilla+1:])
	} else {
		posEspacio := obtenerPosEspacio(strings.TrimSpace(analizador.Comando))
		*parametro = analizador.Comando[:posEspacio]
		analizador.Comando = strings.TrimSpace(analizador.Comando[posEspacio:])
	}
}

// Obtener un parametro que no puede estar entre comillas
func (analizador *Analizador) obtenerDatoParamS(parametro *string, tamanio int) {
	analizador.Comando = strings.TrimSpace(analizador.Comando[tamanio:])
	posEspacio := obtenerPosEspacio(analizador.Comando)
	*parametro = strings.ToLower(analizador.Comando[:posEspacio])
	analizador.Comando = strings.TrimSpace(analizador.Comando[posEspacio:])
}

// Obtener un parametro numerico
func (analizador *Analizador) obtenerDatoParamN(parametro *int, tamanio int) {
	analizador.Comando = strings.TrimSpace(analizador.Comando[tamanio:])
	posEspacio := obtenerPosEspacio(analizador.Comando)
	*parametro, _ = strconv.Atoi(analizador.Comando[:posEspacio])
	analizador.Comando = strings.TrimSpace(analizador.Comando[posEspacio:])
}

func (analizador *Analizador) Analizar() string {
	if analizador.verificarComentario(strings.TrimSpace(analizador.Comando)) {
		return strings.TrimSpace(analizador.Comando)
	}

	nInst, err := analizador.recoInstrucion(strings.ToLower(analizador.Comando))

	if err != nil {
		return err.Error()
	} else {
		fmt.Println(analizador.Comando)
		if nInst == 1 { //Mkdisk
			analizador.Comando = strings.TrimSpace(analizador.Comando[6:])
			size := -1
			var fit, unit, path string

			for len(analizador.Comando) > 0 {
				if analizador.verificarComentario(analizador.Comando) {
					break
				} else if len(analizador.Comando) >= 6 && strings.ToLower(analizador.Comando[:6]) == ">size=" {
					analizador.obtenerDatoParamN(&size, 6)
				} else if len(analizador.Comando) >= 5 && strings.ToLower(analizador.Comando[:5]) == ">fit=" {
					analizador.obtenerDatoParamS(&fit, 5)
				} else if len(analizador.Comando) >= 6 && strings.ToLower(analizador.Comando[:6]) == ">unit=" {
					analizador.obtenerDatoParamS(&unit, 6)
				} else if len(analizador.Comando) >= 6 && strings.ToLower(analizador.Comando[:6]) == ">path=" {
					analizador.obtenerDatoParamC(&path, 6)
				} else {
					return analizador.Comando + " Ingreso un parametro no reconocido"
				}

			}
			return estructuras.Mkdisk(size, path, fit, unit)
		} else if nInst == 2 { //Rmdisk
			analizador.Comando = strings.TrimSpace(analizador.Comando[6:])
			var path string

			for len(analizador.Comando) > 0 {
				if analizador.verificarComentario(analizador.Comando) {
					break
				} else if len(analizador.Comando) >= 6 && strings.ToLower(analizador.Comando[:6]) == ">path=" {
					analizador.obtenerDatoParamC(&path, 6)
				} else {
					return analizador.Comando + " Ingreso un parametro no reconocido"
				}
			}
			return estructuras.Rmdisk(path)
		} else if nInst == 3 { //Fdisk
			analizador.Comando = strings.TrimSpace(analizador.Comando[5:])
			var size int
			var tipo, unit, path, fit, name string

			for len(analizador.Comando) > 0 {
				if analizador.verificarComentario(analizador.Comando) {
					break
				} else if len(analizador.Comando) >= 6 && strings.ToLower(analizador.Comando[:6]) == ">size=" {
					analizador.obtenerDatoParamN(&size, 6)
				} else if len(analizador.Comando) >= 5 && strings.ToLower(analizador.Comando[:5]) == ">fit=" {
					analizador.obtenerDatoParamS(&fit, 5)
				} else if len(analizador.Comando) >= 6 && strings.ToLower(analizador.Comando[:6]) == ">unit=" {
					analizador.obtenerDatoParamS(&unit, 6)
				} else if len(analizador.Comando) >= 6 && strings.ToLower(analizador.Comando[:6]) == ">path=" {
					analizador.obtenerDatoParamC(&path, 6)
				} else if len(analizador.Comando) >= 6 && strings.ToLower(analizador.Comando[:6]) == ">name=" {
					analizador.obtenerDatoParamC(&name, 6)
				} else if len(analizador.Comando) >= 6 && strings.ToLower(analizador.Comando[:6]) == ">type=" {
					analizador.obtenerDatoParamS(&tipo, 6)
				} else {
					return analizador.Comando + " Ingreso un parametro no reconocido"
				}
			}

			return estructuras.Fdisk(size, unit, path, tipo, fit, name)
		} else if nInst == 4 { //Mount
			analizador.Comando = strings.TrimSpace(analizador.Comando[5:])
			var path, name, id string

			for len(analizador.Comando) > 0 {
				if analizador.verificarComentario(analizador.Comando) {
					break
				} else if len(analizador.Comando) >= 6 && strings.ToLower(analizador.Comando[:6]) == ">path=" {
					analizador.obtenerDatoParamC(&path, 6)
				} else if len(analizador.Comando) >= 6 && strings.ToLower(analizador.Comando[:6]) == ">name=" {
					analizador.obtenerDatoParamC(&name, 6)
				} else if len(analizador.Comando) >= 4 && strings.ToLower(analizador.Comando[:4]) == ">id=" {
					analizador.obtenerDatoParamC(&id, 4)
				} else {
					return analizador.Comando + " Ingreso un parametro no reconocido"
				}
			}
			directorios, nombre := estructuras.DivPath(path)
			nuevo := estructuras.NodoMount{Fichero: directorios, Nombre_disco: nombre, IdCompleto: id, Nombre_particion: name}

			return analizador.ListaMount.Agregar(&nuevo)
		} else if nInst == 5 { //Mkfs
			analizador.Comando = strings.TrimSpace(analizador.Comando[4:])
			var tipo, id string
			tipo = "full"

			for len(analizador.Comando) > 0 {
				if analizador.verificarComentario(analizador.Comando) {
					break
				} else if len(analizador.Comando) >= 6 && strings.ToLower(analizador.Comando[:6]) == ">type=" {
					analizador.obtenerDatoParamS(&tipo, 6)
				} else if len(analizador.Comando) >= 4 && strings.ToLower(analizador.Comando[:4]) == ">id=" {
					analizador.obtenerDatoParamC(&id, 4)
				} else {
					return analizador.Comando + " Ingreso un parametro no reconocido"
				}
			}
			return analizador.ListaMount.Mkfs(id, tipo)
		} else if nInst == 6 { //Login
			analizador.Comando = strings.TrimSpace(analizador.Comando[5:])
			var user, pwd, id string

			for len(analizador.Comando) > 0 {
				if analizador.verificarComentario(analizador.Comando) {
					break
				} else if len(analizador.Comando) >= 6 && strings.ToLower(analizador.Comando[:6]) == ">user=" {
					analizador.obtenerDatoParamC(&user, 6)
				} else if len(analizador.Comando) >= 5 && strings.ToLower(analizador.Comando[:5]) == ">pwd=" {
					analizador.obtenerDatoParamC(&pwd, 5)
				} else if len(analizador.Comando) >= 4 && strings.ToLower(analizador.Comando[:4]) == ">id=" {
					analizador.obtenerDatoParamC(&id, 4)
				} else {
					return analizador.Comando + " Ingreso un parametro no reconocido"
				}
			}
			admin := estructuras.AdminUsuario{ListaMount: analizador.ListaMount, Usuario: analizador.Usuario}
			return admin.Login(user, pwd, id)
		} else if nInst == 7 { //Logout
			analizador.Comando = strings.TrimSpace(analizador.Comando[6:])

			if analizador.Comando != "" {
				return analizador.Comando + " Ingreso un parametro no reconocido"
			}
			admin := estructuras.AdminUsuario{ListaMount: analizador.ListaMount, Usuario: analizador.Usuario}
			return admin.Logout()
		} else if nInst == 8 { //Mkgrp
			analizador.Comando = strings.TrimSpace(analizador.Comando[5:])
			var name string

			for len(analizador.Comando) > 0 {
				if analizador.verificarComentario(analizador.Comando) {
					break
				} else if len(analizador.Comando) >= 6 && strings.ToLower(analizador.Comando[:6]) == ">name=" {
					analizador.obtenerDatoParamC(&name, 6)
				} else {
					return analizador.Comando + " Ingreso un parametro no reconocido"
				}
			}
			return fmt.Sprintf("MKGRP %s \n", name)
		} else if nInst == 9 { //Rmgrp
			analizador.Comando = strings.TrimSpace(analizador.Comando[5:])
			var name string

			for len(analizador.Comando) > 0 {
				if analizador.verificarComentario(analizador.Comando) {
					break
				} else if len(analizador.Comando) >= 6 && strings.ToLower(analizador.Comando[:6]) == ">name=" {
					analizador.obtenerDatoParamC(&name, 6)
				} else {
					return analizador.Comando + " Ingreso un parametro no reconocido"
				}
			}
			return fmt.Sprintf("RMGRP %s \n", name)
		} else if nInst == 10 { //Mkusr
			analizador.Comando = strings.TrimSpace(analizador.Comando[5:])
			var user, pwd, grp string

			for len(analizador.Comando) > 0 {
				if analizador.verificarComentario(analizador.Comando) {
					break
				} else if len(analizador.Comando) >= 6 && strings.ToLower(analizador.Comando[:6]) == ">user=" {
					analizador.obtenerDatoParamC(&user, 6)
				} else if len(analizador.Comando) >= 5 && strings.ToLower(analizador.Comando[:5]) == ">pwd=" {
					analizador.obtenerDatoParamC(&pwd, 5)
				} else if len(analizador.Comando) >= 5 && strings.ToLower(analizador.Comando[:5]) == ">grp=" {
					analizador.obtenerDatoParamC(&grp, 5)
				} else {
					return analizador.Comando + " Ingreso un parametro no reconocido"
				}
			}
			return fmt.Sprintf("MKUSR %s %s %s \n", user, pwd, grp)
		} else if nInst == 11 { //Rmusr
			analizador.Comando = strings.TrimSpace(analizador.Comando[5:])
			var user string

			for len(analizador.Comando) > 0 {
				if analizador.verificarComentario(analizador.Comando) {
					break
				} else if len(analizador.Comando) >= 6 && strings.ToLower(analizador.Comando[:6]) == ">user=" {
					analizador.obtenerDatoParamC(&user, 6)
				} else {
					return analizador.Comando + " Ingreso un parametro no reconocido"
				}
			}
			return fmt.Sprintf("RMUSR %s \n", user)
		} else if nInst == 12 { //Mkfile
			analizador.Comando = strings.TrimSpace(analizador.Comando[6:])
			r := false
			var size int
			var path, cont string

			for len(analizador.Comando) > 0 {
				if analizador.verificarComentario(analizador.Comando) {
					break
				} else if len(analizador.Comando) >= 6 && strings.ToLower(analizador.Comando[:6]) == ">path=" {
					analizador.obtenerDatoParamC(&path, 6)
				} else if len(analizador.Comando) >= 6 && strings.ToLower(analizador.Comando[:6]) == ">cont=" {
					analizador.obtenerDatoParamC(&cont, 6)
				} else if len(analizador.Comando) >= 6 && strings.ToLower(analizador.Comando[:6]) == ">size=" {
					analizador.obtenerDatoParamN(&size, 6)
				} else if len(analizador.Comando) >= 2 && strings.ToLower(analizador.Comando[:2]) == ">r" {
					r = true
					analizador.Comando = strings.TrimSpace(analizador.Comando[2:])
				} else {
					return analizador.Comando + " Ingreso un parametro no reconocido"
				}
			}
			return fmt.Sprintf("MKFILE %s %s %d %t \n", path, cont, size, r)
		} else if nInst == 13 { //Mkdir
			analizador.Comando = strings.TrimSpace(analizador.Comando[5:])
			r := false
			var path string

			for len(analizador.Comando) > 0 {
				if analizador.verificarComentario(analizador.Comando) {
					break
				} else if len(analizador.Comando) >= 6 && strings.ToLower(analizador.Comando[:6]) == ">path=" {
					analizador.obtenerDatoParamC(&path, 6)
				} else if len(analizador.Comando) >= 2 && strings.ToLower(analizador.Comando[:2]) == ">r" {
					r = true
					analizador.Comando = strings.TrimSpace(analizador.Comando[2:])
				} else {
					return analizador.Comando + " Ingreso un parametro no reconocido"
				}
			}
			return fmt.Sprintf("MKDIR %s %t \n", path, r)
		} else if nInst == 14 { //Pause
			analizador.Comando = strings.TrimSpace(analizador.Comando[5:])

			if analizador.Comando != "" {
				return analizador.Comando + " Ingreso un parametro no reconocido"
			}
			fmt.Printf("Pause\n")
			bufio.NewReader(os.Stdin).ReadBytes('\n')
		} else if nInst == 15 { //Rep
			analizador.Comando = strings.TrimSpace(analizador.Comando[3:])
			var name, path, id, ruta string

			for len(analizador.Comando) > 0 {
				if analizador.verificarComentario(analizador.Comando) {
					break
				} else if len(analizador.Comando) >= 6 && strings.ToLower(analizador.Comando[:6]) == ">name=" {
					analizador.obtenerDatoParamC(&name, 6)
				} else if len(analizador.Comando) >= 6 && strings.ToLower(analizador.Comando[:6]) == ">path=" {
					analizador.obtenerDatoParamC(&path, 6)
				} else if len(analizador.Comando) >= 4 && strings.ToLower(analizador.Comando[:4]) == ">id=" {
					analizador.obtenerDatoParamC(&id, 4)
				} else if len(analizador.Comando) >= 6 && strings.ToLower(analizador.Comando[:6]) == ">ruta=" {
					analizador.obtenerDatoParamC(&ruta, 6)
				} else {
					return analizador.Comando + " Ingreso un parametro no reconocido"
				}
			}
			return fmt.Sprintf("REP %s %s %s %s \n", name, path, id, ruta)
		}
		return "Esto no deberia pasar :v"
	}

}
