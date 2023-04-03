package estructuras

import (
	"bytes"
	"encoding/binary"
	"os"
	"strconv"
	"strings"
)

type Usuario struct {
	IdG         int
	IdU         int
	NombreG     string
	NombreU     string
	Password    string
	IdParticion string
}

func (this *Usuario) IngresarInfoU(idG int, idU int, nombreG string, nombreU string, password string, idParticion string) {
	this.IdG = idG
	this.IdU = idU
	this.NombreG = nombreG
	this.NombreU = nombreU
	this.Password = password
	this.IdParticion = idParticion
}

func (this *Usuario) BorrarInfoU() {
	this.IdG = 0
	this.IdU = 0
	this.NombreG = ""
	this.NombreU = ""
	this.Password = ""
	this.IdParticion = ""
}

type AdminUsuario struct {
	ListaMount *ListaMount
	Usuario    *Usuario
}

// Comado Login
func (this *AdminUsuario) Login(usuario, password, id string) string {
	usuario = strings.TrimSpace(usuario)
	password = strings.TrimSpace(password)
	id = strings.TrimSpace(id)
	if this.Usuario.IdParticion == "" {
		if usuario == "" || password == "" || id == "" {
			return "No fue posible ejecutar el comando con la informacion proporcionada"
		}

		nodo := this.ListaMount.Buscar(id)

		if nodo != nil {
			var archivo *os.File
			archivo, _ = os.OpenFile(nodo.Fichero+"/"+nodo.Nombre_disco, os.O_RDWR, 0777)

			if archivo != nil {
				sb := SuperBloque{}

				//Particion Primaria
				if nodo.Part_type == 'P' {
					mbr := MBR{}
					archivo.Seek(0, 0)
					binary.Read(extraerStruct(archivo, binary.Size(mbr)), binary.BigEndian, &mbr)
					var i int

					//Verificar la existencia de la particion
					for i = 0; i < 4; i++ {
						if string(bytes.Trim(mbr.Mbr_partition[i].Part_name[:], "\000")) == nodo.Nombre_particion {
							break
						}
					}

					//Error de posicion no encontrada
					if i == 5 {
						this.ListaMount.Eliminar(nodo.IdCompleto)
						archivo.Close()
						return "No fue posible encontrar la particion en el disco"
					} else { //Posicion si Encontrada
						if mbr.Mbr_partition[i].Part_status != '2' {
							archivo.Close()
							return "No se ha aplicado el comando mkfs a la particion"
						}
						//Recuperar la informacion del superbloque
						archivo.Seek(int64(BytetoI32(mbr.Mbr_partition[i].Part_start[:])), 0)
						binary.Read(extraerStruct(archivo, binary.Size(sb)), binary.BigEndian, &sb)
					}
				} else if nodo.Part_type == 'L' { //Particiones Logicas
					ebr := EBR{}
					archivo.Seek(int64(nodo.Part_start), 0)
					binary.Read(extraerStruct(archivo, binary.Size(ebr)), binary.BigEndian, &ebr)
					if ebr.Part_status != '2' {
						archivo.Close()
						return "No se ha aplicado el comando mkfs a la particion"
					}
					binary.Read(extraerStruct(archivo, binary.Size(sb)), binary.BigEndian, &sb)
				}

				//Acceder al archivo user.txt
				utxt := GetContentF(int(BytetoI32((sb.S_inode_start[:])))+binary.Size(TablaInodo{}), archivo)
				usuarios := []string{}
				grupos := []string{}
				ObtenerUG(utxt, &usuarios, &grupos)

				for i := 0; i < len(usuarios); i++ {
					uDatos := strings.Split(usuarios[i], ",")
					if uDatos[0] != "0" && uDatos[3] == usuario && uDatos[4] == password {
						for j := 0; j < len(grupos); j++ {
							gDatos := strings.Split(grupos[j], ",")
							if uDatos[2] == gDatos[2] && gDatos[0] != "0" {
								gid, _ := strconv.Atoi(gDatos[0])
								uid, _ := strconv.Atoi(uDatos[0])
								this.Usuario.IngresarInfoU(gid, uid, uDatos[2], uDatos[3], uDatos[4], id)

								archivo.Close()
								return "Sesion iniciada correctamente"
							}
						}
					}
				}

				archivo.Close()
				return "No fue posible iniciar sesion"
			} else {
				this.ListaMount.Eliminar(nodo.IdCompleto)
				return "No fue posible encontrar el disco de la particion"
			}
		}
	}
	return "Ya existe una sesion iniciada"
}

// Comando Logout
func (this *AdminUsuario) Logout() string {
	if this.Usuario.IdParticion == "" {
		return "No hay una sesion iniciada con anterioridad"
	}
	this.Usuario.BorrarInfoU()
	return "Sesion Cerrada"
}
