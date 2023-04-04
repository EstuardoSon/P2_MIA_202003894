package estructuras

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type GestorArchivos struct {
	ListaMount *ListaMount
	Usuario    *Usuario
}

func (this *GestorArchivos) Mkfile(path_fichero, path_archivo string, r bool, size int, cont_fichero string, cont_archivo string) string {
	if this.Usuario.NombreG == "" && this.Usuario.NombreU == "" {
		return "No existe una sesion iniciada"
	}
	nodo := this.ListaMount.Buscar(this.Usuario.IdParticion)
	if nodo != nil {
		var archivo *os.File
		archivo, _ = os.OpenFile(nodo.Fichero+"/"+nodo.Nombre_disco, os.O_RDWR, 0777)

		if archivo != nil {
			inicioSB := 0
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
					inicioSB = int(BytetoI32(mbr.Mbr_partition[i].Part_start[:]))
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
				inicioSB = nodo.Part_start + binary.Size(ebr)
			}

			match, _ := regexp.MatchString("(\\/)(([a-zA-Z0-9_ñÑáéíóúÁÉÍÓÚ ]+(\\/))*[a-zA-Z0-9_ñÑáéíóúÁÉÍÓÚ ]+)?", path_fichero)
			if match {
				path_fichero = path_fichero[1:]
				ficheros := strings.Split(path_fichero, "/")

				//Texto a escribir
				bandera := true
				textoArchivo := ""

				match, _ = regexp.MatchString("(\\/)(([a-zA-Z0-9_ñÑáéíóúÁÉÍÓÚ ]+(\\/))*[a-zA-Z0-9_ñÑáéíóúÁÉÍÓÚ ]+)?", cont_fichero)
				matchF, _ := regexp.MatchString("[a-zA-Z0-9_ñÑáéíóúÁÉÍÓÚ ]+(\\.[a-zA-Z0-9_ñÑáéíóúÁÉÍÓÚ ]+)", cont_archivo)
				if cont_fichero != "" && cont_archivo != "" {
					if match && matchF {
						file, err := ioutil.ReadFile(cont_fichero + "/" + cont_archivo)
						if err != nil {
							return err.Error()
						}
						text := string(file)
						fmt.Println(text)
					}
				}
				if size >= 0 {
					for i := 0; i < size; i++ {
						modulo := i % 10
						textoArchivo += strconv.Itoa(modulo)
					}
				} else {
					bandera = false
					return "El valor de SIZE debe ser positivo"
				}

				if bandera {
					//Ejecutar
					if len(textoArchivo) <= 1024 {
						return this.buscarfichero(ficheros, path_archivo, r, &sb, inicioSB, int(BytetoI32(sb.S_inode_start[:])), archivo, textoArchivo)
					} else {
						return "El texto que desea ingresar excede el espacio disponible para un archivo"
					}
				}
			}

			archivo.Close()
		} else {
			this.ListaMount.Eliminar(nodo.IdCompleto)
			return "No fue posible encontrar el disco de la particion"
		}
	}

	return "No se encontro una Particion Montada con el ID ingresado"
}

// Verificar los permisos de escritura y lectura de un usuario en un inodo
func (this *GestorArchivos) verificarPermisos(inodo *TablaInodo, escritura *bool, lectura *bool) {
	if this.Usuario.NombreG == "root" && this.Usuario.NombreU == "root" {
		*escritura = true
		*lectura = true
		return
	}

	permisos := int(BytetoI32(inodo.I_perm[:]))
	propietario := permisos / 100
	permisos = permisos - (propietario * 100)
	grupo := permisos / 10
	permisos = permisos - (grupo * 10)
	otros := permisos

	if this.Usuario.IdU == int(BytetoI32(inodo.I_uid[:])) {
		if propietario > 3 {
			*lectura = true
		}
		if propietario == 2 || propietario == 3 || propietario == 6 || propietario == 7 {
			*escritura = true
		}
	} else if this.Usuario.IdG == int(BytetoI32(inodo.I_gid[:])) {
		if grupo > 3 {
			*lectura = true
		}
		if grupo == 2 || grupo == 3 || grupo == 6 || grupo == 7 {
			*escritura = true
		}
	} else {
		if otros > 3 {
			*lectura = true
		}
		if otros == 2 || otros == 3 || otros == 6 || otros == 7 {
			*escritura = true
		}
	}
}

// Buscar una carpeta o archivo en una carpeta padre
func (this *GestorArchivos) buscarEnCarpeta(ti *TablaInodo, inicioInodo int, archivo *os.File, nombre string) int {
	bc := BloqueCarpeta{}

	ubicacion := -1
	for i := 0; i < 16; i++ {
		if int(BytetoI32(ti.I_block[i][:])) != -1 {
			archivo.Seek(int64(BytetoI32(ti.I_block[i][:])), 0)

			//Bloques directos
			binary.Read(extraerStruct(archivo, binary.Size(bc)), binary.BigEndian, &bc)
			for j := 0; j < 4; j++ {
				if string(bytes.Trim(bc.B_content[j].B_name[:], "\000")) == nombre {
					ubicacion = int(BytetoI32(bc.B_content[j].B_inodo[:]))
				}
			}
		}
	}
	ti.I_atime = I64toByte(time.Now().Unix())
	archivo.Seek(int64(inicioInodo), 0)

	var bs bytes.Buffer
	binary.Write(&bs, binary.BigEndian, ti)
	_, _ = archivo.Write(bs.Bytes())

	return ubicacion
}

// Realizar acciones del comando mkfile
func (this *GestorArchivos) buscarfichero(ficheros []string, nombreArchivo string, r bool, sb *SuperBloque, inicioSB, inicioInodo int, archivo *os.File, textoArchivo string) string {
	ti := TablaInodo{}

	//Obtener Inodo
	archivo.Seek(int64(inicioInodo), 0)
	binary.Read(extraerStruct(archivo, binary.Size(ti)), binary.BigEndian, &ti)

	escritura := false
	lectura := false
	this.verificarPermisos(&ti, &escritura, &lectura)
	res := ""

	if len(ficheros) > 0 {
		if ti.I_type == '0' {
			fichero := ficheros[0]
			newFicheros := []string{}
			newFicheros = append(newFicheros, ficheros[1:]...)
			ubicacion := this.buscarEnCarpeta(&ti, inicioInodo, archivo, fichero)

			if ubicacion != -1 {
				return this.buscarfichero(newFicheros, nombreArchivo, r, sb, inicioSB, ubicacion, archivo, textoArchivo)

			} else if ubicacion == -1 && r {
				if escritura {
					ubicacion = this.buscarEspacioCarpeta(&ti, inicioInodo, archivo, fichero, sb, inicioSB)
					if ubicacion != -1 {

						res = this.buscarfichero(newFicheros, nombreArchivo, r, sb, inicioSB, ubicacion, archivo, textoArchivo)

						ti.I_mtime = I64toByte(time.Now().Unix())
						archivo.Seek(int64(inicioInodo), 0)

						var bs bytes.Buffer
						binary.Write(&bs, binary.BigEndian, ti)
						_, _ = archivo.Write(bs.Bytes())

						return res
					} else {
						return "No fue posible crear el fichero: " + fichero
					}
				} else {
					return "El usuario no tiene permisos de Escritura"
				}
			} else {
				return "No se encontro el fichero "
			}
		} else {
			return "El inodo no corresponde a una carpeta"
		}
	}
	if ti.I_type == '0' && len(ficheros) == 0 {
		ubicacion := this.buscarEnCarpeta(&ti, inicioInodo, archivo, nombreArchivo)

		if ubicacion != -1 {
			return "El archivo ya existe"

		} else {
			if escritura {
				ubicacion = this.buscarEspacioArchivo(&ti, inicioInodo, archivo, nombreArchivo, sb, inicioSB,
					textoArchivo)
				if ubicacion != -1 {
					ti.I_mtime = I64toByte(time.Now().Unix())
					archivo.Seek(int64(inicioInodo), 0)

					var bs bytes.Buffer
					binary.Write(&bs, binary.BigEndian, ti)
					_, _ = archivo.Write(bs.Bytes())
					return "Archivo Creado: " + nombreArchivo
				}
				return "No fue posible Crear el archivo: " + nombreArchivo
			} else {
				return "El usuario no tiene permisos de Escritura"
			}
		}
	} else {
		return "El inodo no corresponde a una carpeta"
	}
}

// Crear Archivos
func (this *GestorArchivos) crearArchivo(archivo *os.File, textoArchivo string, sb *SuperBloque, inicioSB int, ubicacion *int) {
	var caracter byte
	caracter = '1'
	sb.S_first_ino = I32toByte(int32(BuscarBM_i(sb, archivo)))

	archivo.Seek(int64(BytetoI32(sb.S_bm_inode_start[:])+BytetoI32(sb.S_first_ino[:])), 0)

	var bs bytes.Buffer
	binary.Write(&bs, binary.BigEndian, caracter)
	_, _ = archivo.Write(bs.Bytes())

	sb.S_free_inodes_count = I32toByte(BytetoI32(sb.S_free_inodes_count[:]) - int32(1))

	tiArchivo := TablaInodo{}
	tiArchivo.I_uid = I32toByte(int32(this.Usuario.IdU))
	tiArchivo.I_gid = I32toByte(int32(this.Usuario.IdG))
	tiArchivo.I_size = I32toByte(0)
	tiArchivo.I_atime = I64toByte(time.Now().Unix())
	tiArchivo.I_ctime = I64toByte(time.Now().Unix())
	tiArchivo.I_mtime = I64toByte(time.Now().Unix())
	for i := 0; i < 16; i++ {
		tiArchivo.I_block[i] = I32toByte(-1)
	}
	tiArchivo.I_type = '1'
	tiArchivo.I_perm = I32toByte(664)

	archivo.Seek(int64(BytetoI32(sb.S_inode_start[:])+(BytetoI32(sb.S_first_ino[:])+int32(binary.Size(tiArchivo)))), 0)
	bs.Reset()
	binary.Write(&bs, binary.BigEndian, tiArchivo)
	_, _ = archivo.Write(bs.Bytes())

	WriteInFile(textoArchivo, sb, inicioSB, int(BytetoI32(sb.S_inode_start[:])+(BytetoI32(sb.S_first_ino[:])+int32(binary.Size(tiArchivo)))), archivo)

	*ubicacion = int(BytetoI32(sb.S_inode_start[:]) + (BytetoI32(sb.S_first_ino[:]) + int32(binary.Size(tiArchivo))))
}

// Buscar un espacio libre en carpeta para crear un archivo
func (this *GestorArchivos) buscarEspacioArchivo(ti *TablaInodo, inicioInodo int, archivo *os.File, nombre string, sb *SuperBloque,
	inicioSB int, textoArchivo string) int {
	bc := BloqueCarpeta{}

	ubicacion := -1
	bandera := true
	var caracter byte
	caracter = '1'

	for i := 0; i <= 15; i++ {
		if ubicacion == -1 && bandera {
			if int(BytetoI32(ti.I_block[i][:])) != -1 {
				archivo.Seek(int64(BytetoI32(ti.I_block[i][:])), 0)
				binary.Read(extraerStruct(archivo, binary.Size(bc)), binary.BigEndian, &bc)

				for j := 0; j < 4; j++ {
					if int(BytetoI32(bc.B_content[j].B_inodo[:])) == -1 && ubicacion == -1 && bandera {
						if int(BytetoI32(sb.S_free_inodes_count[:])) > 0 {
							sb.S_first_ino = I32toByte(int32(BuscarBM_i(sb, archivo)))
							for a, char := range nombre {
								bc.B_content[j].B_name[a] = byte(char)
							}
							bc.B_content[j].B_inodo = I32toByte(int32(int(BytetoI32(sb.S_inode_start[:])) + (int(BytetoI32(sb.S_first_ino[:])) * binary.Size(ti))))

							this.crearArchivo(archivo, textoArchivo, sb, inicioSB, &ubicacion)
							bandera = false
							break
						} else {
							bandera = false
							break
						}
					}
				}

				archivo.Seek(int64(BytetoI32(ti.I_block[i][:])), 0)
				var bs bytes.Buffer
				binary.Write(&bs, binary.BigEndian, bc)
				_, _ = archivo.Write(bs.Bytes())
			} else if ubicacion == -1 && bandera {
				if int(BytetoI32(sb.S_free_inodes_count[:])) > 0 && int(BytetoI32(sb.S_free_blocks_count[:])) > 0 {
					// Bloque carpeta
					sb.S_first_blo = I32toByte(int32(BuscarBM_b(sb, archivo)))
					ti.I_block[i] = I32toByte(BytetoI32(sb.S_block_start[:]) + (BytetoI32(sb.S_first_blo[:]) * int32(binary.Size(bc))))
					archivo.Seek(int64(BytetoI32(sb.S_bm_block_start[:])+BytetoI32(sb.S_first_blo[:])), 0)

					var bs bytes.Buffer
					binary.Write(&bs, binary.BigEndian, caracter)
					_, _ = archivo.Write(bs.Bytes())

					sb.S_free_blocks_count = I32toByte(BytetoI32(sb.S_free_blocks_count[:]) - int32(1))

					sb.S_first_ino = I32toByte(int32(BuscarBM_i(sb, archivo)))
					bc.B_content[0].B_inodo = I32toByte(BytetoI32(sb.S_inode_start[:]) + (BytetoI32(sb.S_first_ino[:]) * int32(binary.Size(ti))))
					for a, char := range nombre {
						bc.B_content[0].B_name[a] = byte(char)
					}
					for j := 1; j < 4; j++ {
						bc.B_content[j].B_inodo = I32toByte(-1)
					}

					archivo.Seek(int64(BytetoI32(ti.I_block[i][:])), 0)
					bs.Reset()
					binary.Write(&bs, binary.BigEndian, bc)
					_, _ = archivo.Write(bs.Bytes())

					this.crearArchivo(archivo, textoArchivo, sb, inicioSB, &ubicacion)
					bandera = false
				} else {
					bandera = false
					break
				}
			}
		} else {
			break
		}
	}

	ti.I_mtime = I64toByte(time.Now().Unix())

	archivo.Seek(int64(inicioInodo), 0)
	var bs bytes.Buffer
	binary.Write(&bs, binary.BigEndian, ti)
	_, _ = archivo.Write(bs.Bytes())

	sb.S_first_blo = I32toByte(int32(BuscarBM_b(sb, archivo)))
	sb.S_first_ino = I32toByte(int32(BuscarBM_i(sb, archivo)))
	archivo.Seek(int64(inicioSB), 0)
	bs.Reset()
	binary.Write(&bs, binary.BigEndian, sb)
	_, _ = archivo.Write(bs.Bytes())

	return ubicacion
}

// Buscar un espacio libre en carpeta para crear una carpeta
func (this *GestorArchivos) buscarEspacioCarpeta(ti *TablaInodo, inicioInodo int, archivo *os.File, nombre string, sb *SuperBloque,
	inicioSB int) int {
	bc := BloqueCarpeta{}

	ubicacion := -1
	bandera := true

	var caracter byte
	caracter = '1'

	for i := 0; i <= 15; i++ {
		if ubicacion == -1 && bandera {
			if int(BytetoI32(ti.I_block[i][:])) != -1 {
				archivo.Seek(int64(BytetoI32(ti.I_block[i][:])), 0)
				binary.Read(extraerStruct(archivo, binary.Size(bc)), binary.BigEndian, &bc)

				for j := 0; j < 4; j++ {
					if int(BytetoI32(bc.B_content[j].B_inodo[:])) == -1 && ubicacion == -1 && bandera {
						if int(BytetoI32(sb.S_free_inodes_count[:])) > 0 && int(BytetoI32(sb.S_free_blocks_count[:])) > 0 {
							sb.S_first_ino = I32toByte(int32(BuscarBM_i(sb, archivo)))
							for a, char := range nombre {
								bc.B_content[j].B_name[a] = byte(char)
							}
							bc.B_content[j].B_inodo = I32toByte(BytetoI32(sb.S_inode_start[:]) + (BytetoI32(sb.S_first_ino[:]) * int32(binary.Size(ti))))
							this.crearCarpeta(archivo, sb, &ubicacion, inicioInodo)
							bandera = false
							break
						} else {
							bandera = false
							break
						}
					}
				}

				archivo.Seek(int64(BytetoI32(ti.I_block[i][:])), 0)
				var bs bytes.Buffer
				binary.Write(&bs, binary.BigEndian, bc)
				_, _ = archivo.Write(bs.Bytes())
			} else if ubicacion == -1 && bandera {
				if int(BytetoI32(sb.S_free_inodes_count[:])) > 0 && int(BytetoI32(sb.S_free_blocks_count[:])) > 1 {
					// Bloque carpeta
					sb.S_first_blo = I32toByte(int32(BuscarBM_b(sb, archivo)))
					ti.I_block[i] = I32toByte(BytetoI32(sb.S_block_start[:]) + (BytetoI32(sb.S_first_blo[:]) * int32(binary.Size(bc))))
					archivo.Seek(int64(BytetoI32(sb.S_bm_block_start[:])+BytetoI32(sb.S_first_blo[:])), 0)

					var bs bytes.Buffer
					binary.Write(&bs, binary.BigEndian, caracter)
					_, _ = archivo.Write(bs.Bytes())

					sb.S_free_blocks_count = I32toByte(BytetoI32(sb.S_free_blocks_count[:]) - int32(1))

					sb.S_first_ino = I32toByte(int32(BuscarBM_i(sb, archivo)))
					bc.B_content[0].B_inodo = I32toByte(BytetoI32(sb.S_inode_start[:]) + (BytetoI32(sb.S_first_ino[:]) * int32(binary.Size(ti))))
					sb.S_first_ino = I32toByte(int32(BuscarBM_i(sb, archivo)))
					for a, char := range nombre {
						bc.B_content[0].B_name[a] = byte(char)
					}
					for j := 1; j < 4; j++ {
						bc.B_content[j].B_inodo = I32toByte(int32(-1))
					}

					archivo.Seek(int64(BytetoI32(ti.I_block[i][:])), 0)
					bs.Reset()
					binary.Write(&bs, binary.BigEndian, bc)
					_, _ = archivo.Write(bs.Bytes())

					this.crearCarpeta(archivo, sb, &ubicacion, inicioInodo)
					bandera = false
				} else {
					bandera = false
					break
				}
			}
		} else {
			break
		}
	}

	ti.I_mtime = I64toByte(time.Now().Unix())

	archivo.Seek(int64(inicioInodo), 0)
	var bs bytes.Buffer
	binary.Write(&bs, binary.BigEndian, ti)
	_, _ = archivo.Write(bs.Bytes())

	sb.S_first_blo = I32toByte(int32(BuscarBM_b(sb, archivo)))
	sb.S_first_ino = I32toByte(int32(BuscarBM_i(sb, archivo)))
	archivo.Seek(int64(inicioSB), 0)
	bs.Reset()
	binary.Write(&bs, binary.BigEndian, sb)
	_, _ = archivo.Write(bs.Bytes())

	return ubicacion
}

// Crear Carpetas
func (this *GestorArchivos) crearCarpeta(archivo *os.File, sb *SuperBloque, ubicacion *int, inicioInodo int) {
	var caracter byte
	caracter = '1'
	sb.S_first_blo = I32toByte(int32(BuscarBM_b(sb, archivo)))
	sb.S_first_ino = I32toByte(int32(BuscarBM_i(sb, archivo)))
	archivo.Seek(int64(BytetoI32(sb.S_bm_block_start[:])+BytetoI32(sb.S_first_blo[:])), 0)
	var bs bytes.Buffer
	binary.Write(&bs, binary.BigEndian, caracter)
	_, _ = archivo.Write(bs.Bytes())
	archivo.Seek(int64(BytetoI32(sb.S_bm_inode_start[:])+BytetoI32(sb.S_first_ino[:])), 0)
	bs.Reset()
	binary.Write(&bs, binary.BigEndian, caracter)
	_, _ = archivo.Write(bs.Bytes())
	sb.S_free_inodes_count = I32toByte(BytetoI32(sb.S_free_inodes_count[:]) - int32(1))
	sb.S_free_blocks_count = I32toByte(BytetoI32(sb.S_free_blocks_count[:]) - int32(1))

	tCarpetaN := TablaInodo{}
	tCarpetaN.I_uid = I32toByte(int32(this.Usuario.IdU))
	tCarpetaN.I_gid = I32toByte(int32(this.Usuario.IdG))
	tCarpetaN.I_size = I32toByte(0)
	tCarpetaN.I_atime = I64toByte(time.Now().Unix())
	tCarpetaN.I_ctime = I64toByte(time.Now().Unix())
	tCarpetaN.I_mtime = I64toByte(time.Now().Unix())
	tCarpetaN.I_block[0] = I32toByte(BytetoI32(sb.S_block_start[:]) + (BytetoI32(sb.S_first_blo[:]) * int32(binary.Size(BloqueCarpeta{}))))
	for i := 1; i < 16; i++ {
		tCarpetaN.I_block[i] = I32toByte(-1)
	}
	tCarpetaN.I_type = '0'
	tCarpetaN.I_perm = I32toByte(664)

	archivo.Seek(int64(BytetoI32(sb.S_inode_start[:])+(BytetoI32(sb.S_first_ino[:])*int32(binary.Size(tCarpetaN)))), 0)
	bs.Reset()
	binary.Write(&bs, binary.BigEndian, tCarpetaN)
	_, _ = archivo.Write(bs.Bytes())

	nbc := BloqueCarpeta{}
	for i, char := range "." {
		nbc.B_content[0].B_name[i] = byte(char)
	}
	nbc.B_content[0].B_inodo = I32toByte(BytetoI32(sb.S_inode_start[:]) + (BytetoI32(sb.S_first_ino[:]) * int32(binary.Size(tCarpetaN))))
	for i, char := range ".." {
		nbc.B_content[1].B_name[i] = byte(char)
	}
	nbc.B_content[1].B_inodo = I32toByte(int32(inicioInodo))
	nbc.B_content[2].B_inodo = I32toByte(-1)
	nbc.B_content[3].B_inodo = I32toByte(-1)

	archivo.Seek(int64(BytetoI32(sb.S_block_start[:])+(BytetoI32(sb.S_first_blo[:])*int32(binary.Size(nbc)))), 0)
	bs.Reset()
	binary.Write(&bs, binary.BigEndian, nbc)
	_, _ = archivo.Write(bs.Bytes())

	*ubicacion = int(BytetoI32(sb.S_inode_start[:]) + (BytetoI32(sb.S_first_ino[:]) * int32(binary.Size(tCarpetaN))))
}
