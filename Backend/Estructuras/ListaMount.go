package estructuras

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"regexp"
	"strconv"
	"time"
)

type ListaMount struct {
	Inicio *NodoMount
	Fin    *NodoMount
}

// Obtener el numero y nombre del ID
func (this *ListaMount) obtenerIdNumero(nuevo *NodoMount) {
	valor := nuevo.IdCompleto[2:]
	var numero string

	var i int
	for i = 0; i < len(valor); i++ {
		if valor[i] >= 48 && valor[i] <= 57 {
			numero += string(valor[i])
		} else {
			break
		}
	}

	nuevo.Id = valor[i:]
	nuevo.Numero, _ = strconv.Atoi(numero[:])
}

// Verificar ID completo
func (this *ListaMount) verificarID(nuevo *NodoMount) (bool, string) {
	match, _ := regexp.MatchString("(94)[0-9]+[a-zA-ZñÑ]+", nuevo.IdCompleto)
	if match {
		this.obtenerIdNumero(nuevo)
		return true, ""
	}

	return false, "El ID no cumple con los requisitos minimos 94 + Num + Letra"
}

// Verificar que la particion no este actualmente montada
func (this *ListaMount) verificarLista(nuevo *NodoMount) (bool, string) {
	aux := this.Inicio
	for aux != nil {
		if aux.IdCompleto == nuevo.IdCompleto {
			return false, "El ID ya esta en uso"
		} else if aux.Fichero == nuevo.Fichero && aux.Nombre_disco == nuevo.Nombre_disco && aux.Nombre_particion == nuevo.Nombre_particion {
			return false, "La particion ya se encuentra montada"
		}
		aux = aux.Next
	}

	return true, ""
}

// Verificar la existencia de la parcticion entre Principales y Extendidas
func (this *ListaMount) verificarParticion(archivo *os.File, mbr *MBR, nuevo *NodoMount) (bool, string) {
	verificacion := false
	err := ""
	for i := 0; i < 4; i++ {
		if string(bytes.Trim(mbr.Mbr_partition[i].Part_name[:], "\000")) == nuevo.Nombre_particion && mbr.Mbr_partition[i].Part_status != '0' && mbr.Mbr_partition[i].Part_type != 'E' {
			if mbr.Mbr_partition[i].Part_status == '2' {
				sb := SuperBloque{}
				archivo.Seek(int64(BytetoI32(mbr.Mbr_partition[i].Part_start[:])), 0)

				binary.Read(extraerStruct(archivo, binary.Size(sb)), binary.BigEndian, &sb)

				sb.S_mtime = I64toByte(time.Now().Unix())
				sb.S_mnt_count = I32toByte(BytetoI32(sb.S_mnt_count[:]) + 1)

				archivo.Seek(int64(BytetoI32(mbr.Mbr_partition[i].Part_start[:])), 0)

				var bs bytes.Buffer
				binary.Write(&bs, binary.BigEndian, sb)
				_, _ = archivo.Write(bs.Bytes())
			}
			nuevo.Part_start = int(BytetoI32(mbr.Mbr_partition[i].Part_start[:]))
			nuevo.Part_type = mbr.Mbr_partition[i].Part_type
			verificacion = true
			break
		} else if string(bytes.Trim(mbr.Mbr_partition[i].Part_name[:], "\000")) == nuevo.Nombre_particion && mbr.Mbr_partition[i].Part_status == '0' {
			err = "La Particion no puede montarse ya que fue eliminada FAST"
			verificacion = false
			break
		} else if mbr.Mbr_partition[i].Part_type == 'E' && mbr.Mbr_partition[i].Part_status == '1' {
			verificacion, err = this.verificarParticionL(archivo, &mbr.Mbr_partition[i], nuevo)
			if verificacion {
				break
			}
		}
	}
	if !verificacion && err == "" {
		err = "No se encontro una Particion con el nombre establecido"
	}
	return verificacion, err
}

// Verificar la existencia de la particion entre Logicas
func (this *ListaMount) verificarParticionL(archivo *os.File, p *Partition, nuevo *NodoMount) (bool, string) {
	ebr := EBR{}
	archivo.Seek(int64(BytetoI32(p.Part_start[:])), 0)

	binary.Read(extraerStruct(archivo, binary.Size(ebr)), binary.BigEndian, &ebr)

	if string(bytes.Trim(ebr.Part_name[:], "\000")) == nuevo.Nombre_particion && ebr.Part_status != '0' {
		if ebr.Part_status == '2' {
			sb := SuperBloque{}
			archivo.Seek(int64(int(BytetoI32(ebr.Part_start[:]))+binary.Size(ebr)), 0)

			sb.S_mtime = I64toByte(time.Now().Unix())
			sb.S_mnt_count = I32toByte(BytetoI32(sb.S_mnt_count[:]) + 1)

			archivo.Seek(int64(int(BytetoI32(ebr.Part_start[:]))+binary.Size(ebr)), 0)

			var bs bytes.Buffer
			binary.Write(&bs, binary.BigEndian, sb)
			_, _ = archivo.Write(bs.Bytes())
		}
		nuevo.Part_type = 'L'
		nuevo.Part_start = int(BytetoI32(ebr.Part_start[:]))
		return true, ""
	} else if string(bytes.Trim(ebr.Part_name[:], "\000")) == nuevo.Nombre_particion && ebr.Part_status == '0' {
		return false, "La Particion no puede montarse ya que fue eliminada FAST"
	}
	for BytetoI32(ebr.Part_next[:]) != -1 {
		archivo.Seek(int64(BytetoI32(ebr.Part_next[:])), 0)
		binary.Read(extraerStruct(archivo, binary.Size(ebr)), binary.BigEndian, &ebr)

		if string(bytes.Trim(ebr.Part_name[:], "\000")) == nuevo.Nombre_particion && ebr.Part_status != '0' {
			if ebr.Part_status == '2' {
				sb := SuperBloque{}
				archivo.Seek(int64(int(BytetoI32(ebr.Part_start[:]))+binary.Size(ebr)), 0)

				sb.S_mtime = I64toByte(time.Now().Unix())
				sb.S_mnt_count = I32toByte(BytetoI32(sb.S_mnt_count[:]) + 1)

				archivo.Seek(int64(int(BytetoI32(ebr.Part_start[:]))+binary.Size(ebr)), 0)

				var bs bytes.Buffer
				binary.Write(&bs, binary.BigEndian, sb)
				_, _ = archivo.Write(bs.Bytes())
			}
			nuevo.Part_type = 'L'
			nuevo.Part_start = int(BytetoI32(ebr.Part_start[:]))
			return true, ""
		}
	}
	return false, "No se Encontro la Particion Deseada"
}

// Verificar la Existencia del disco
func (this *ListaMount) verificarRuta(nuevo *NodoMount) (bool, string) {
	match, _ := regexp.MatchString("(\\/[a-zA-Z0-9_ñÑáéíóúÁÉÍÓÚ ]+\\/)([a-zA-Z0-9_ñÑáéíóúÁÉÍÓÚ ]+(\\/))*[a-zA-Z0-9_ñÑáéíóúÁÉÍÓÚ ]+(\\.dsk)", nuevo.Fichero+"/"+nuevo.Nombre_disco)
	if !match {
		return false, "La direccion ingresada no es valida o la extension del archivo no es la adecuada"
	}

	var file *os.File
	file, _ = os.OpenFile(nuevo.Fichero+"/"+nuevo.Nombre_disco, os.O_RDWR, 0777)

	if file != nil {
		mbr := MBR{}
		file.Seek(0, 0)

		binary.Read(extraerStruct(file, binary.Size(mbr)), binary.BigEndian, &mbr)

		verificar, err := this.verificarParticion(file, &mbr, nuevo)
		file.Close()
		return verificar, err
	} else {
		return false, "El archivo no se encontro en la ruta establecida"
	}
}

// Imprimir la lista
func (this *ListaMount) imprimirLista() string {
	aux := this.Inicio
	var lista string
	for aux != nil {
		lista += fmt.Sprintf("Id: %s Nombre: %s Numero: %d \n", aux.IdCompleto, aux.Id, aux.Numero)
		aux = aux.Next
	}

	return lista
}

// Agregar Nodos a la lista
func (this *ListaMount) Agregar(nuevo *NodoMount) string {
	verificar, err := this.verificarID(nuevo)
	if verificar {
		verificar, err = this.verificarLista(nuevo)
		if verificar {
			verificar, err = this.verificarRuta(nuevo)
			if verificar {
				if this.Inicio == nil {
					this.Inicio = nuevo
					this.Fin = nuevo
				} else {
					this.Fin.Next = nuevo
					this.Fin = nuevo
				}

				return this.imprimirLista()
			}
		}
	}
	return err
}

// Buscar Nodo en la lista
func (this *ListaMount) Buscar(idCompleto string) *NodoMount {
	aux := this.Inicio
	for aux != nil {
		if aux.IdCompleto == idCompleto {
			return aux
		}
		aux = aux.Next
	}

	return nil
}

// Eliminar Nodo de la lista
func (this *ListaMount) Eliminar(idCompleto string) *NodoMount {
	aux := this.Inicio
	var ant *NodoMount
	for aux != nil {
		if aux.IdCompleto == idCompleto {
			if ant != nil {
				if this.Inicio == aux {
					this.Inicio = aux.Next
				}
				if this.Fin == aux {
					this.Fin = ant
				}
				ant.Next = aux.Next
				aux.Next = nil
				return aux
			} else {
				if this.Inicio == aux {
					this.Inicio = aux.Next
				}
				if this.Fin == aux {
					this.Fin = ant
				}
				return aux
			}
		}
		ant = aux
		aux = aux.Next
	}
	return nil
}

// Comando Mkfs
func (this *ListaMount) Mkfs(id string, tipo string) string {
	nodo := this.Buscar(id)

	if tipo == "full" {
		if nodo != nil {
			mbr := MBR{}
			var archivo *os.File
			archivo, _ = os.OpenFile(nodo.Fichero+"/"+nodo.Nombre_disco, os.O_RDWR, 0777)
			if archivo != nil {
				archivo.Seek(0, 0)
				binary.Read(extraerStruct(archivo, binary.Size(mbr)), binary.BigEndian, &mbr)

				sb := SuperBloque{}
				ti := TablaInodo{}
				bc := BloqueCarpeta{}

				sb.S_mtime = I64toByte(time.Now().Unix())
				sb.S_mnt_count = I32toByte(1)
				sb.S_magic = I32toByte(0xEF53)

				//Formatear particion Primaria
				if nodo.Part_type == 'P' {
					p := -1
					for i := 0; i < 4; i++ {
						if string(bytes.Trim(mbr.Mbr_partition[i].Part_name[:], "\000")) == nodo.Nombre_particion &&
							mbr.Mbr_partition[i].Part_status != '0' && mbr.Mbr_partition[i].Part_type == 'P' {
							p = i
							break
						}
					}

					if p != -1 && p != 4 {
						//Llenar de \0
						archivo.Seek(0, 0)
						var vacio byte
						vacio = '\000'
						for i := 0; i < int(BytetoI32(mbr.Mbr_partition[p].Part_size[:])); i++ {
							var bs bytes.Buffer
							binary.Write(&bs, binary.BigEndian, vacio)
							_, _ = archivo.Write(bs.Bytes())
						}

						//Calcular el valor de N y Numero de Estructuras
						var n float64
						n = float64(int(BytetoI32(mbr.Mbr_partition[p].Part_size[:]))-binary.Size(sb)) / float64(4+(binary.Size(ti))+(3*binary.Size(bc)))
						sb.S_filesystem_type = I32toByte(2)

						num_estructura := int32(math.Floor(n))
						bloques := 3 * num_estructura

						//Inicializar SuperBloque
						sb.S_inodes_count = I32toByte(num_estructura)
						sb.S_blocks_count = I32toByte(bloques)
						sb.S_free_inodes_count = I32toByte(num_estructura - 2)
						sb.S_free_blocks_count = I32toByte(bloques - 2)
						sb.S_inode_size = I32toByte(int32(binary.Size(ti)))
						sb.S_block_size = I32toByte(int32(binary.Size(bc)))

						sb.S_bm_inode_start = I32toByte(BytetoI32(mbr.Mbr_partition[p].Part_start[:]) + int32(binary.Size(sb)))

						sb.S_first_ino = I32toByte(2)
						sb.S_first_blo = I32toByte(2)
						sb.S_bm_block_start = I32toByte(BytetoI32(sb.S_bm_inode_start[:]) + num_estructura)
						sb.S_inode_start = I32toByte(BytetoI32(sb.S_bm_block_start[:]) + bloques)
						sb.S_block_start = I32toByte(BytetoI32(sb.S_inode_start[:]) + (num_estructura * int32(binary.Size(ti))))

						//Inicializar Inodo Raiz
						ti.I_uid = I32toByte(1)
						ti.I_gid = I32toByte(1)
						ti.I_atime = I64toByte(time.Now().Unix())
						ti.I_ctime = I64toByte(time.Now().Unix())
						ti.I_mtime = I64toByte(time.Now().Unix())
						ti.I_perm = I32toByte(664)
						ti.I_block[0] = sb.S_block_start
						for i := 1; i < 16; i++ {
							ti.I_block[i] = I32toByte(-1)
						}
						ti.I_type = '0'
						ti.I_size = I32toByte(0)

						archivo.Seek(int64(BytetoI32(sb.S_inode_start[:])), 0)

						var bs bytes.Buffer
						binary.Write(&bs, binary.BigEndian, ti)
						_, _ = archivo.Write(bs.Bytes())

						for i, char := range "." {
							bc.B_content[0].B_name[i] = byte(char)
						}
						bc.B_content[0].B_inodo = sb.S_inode_start
						for i, char := range ".." {
							bc.B_content[1].B_name[i] = byte(char)
						}
						bc.B_content[1].B_inodo = sb.S_inode_start
						for i, char := range "users.txt" {
							bc.B_content[2].B_name[i] = byte(char)
						}
						bc.B_content[2].B_inodo = I32toByte(BytetoI32(sb.S_inode_start[:]) + int32(binary.Size(ti)))
						bc.B_content[3].B_inodo = I32toByte(-1)
						archivo.Seek(int64(BytetoI32(sb.S_block_start[:])), 0)

						bs.Reset()
						binary.Write(&bs, binary.BigEndian, bc)
						_, _ = archivo.Write(bs.Bytes())

						//Generar Inodo y archivo Users.txt
						tiUsers := TablaInodo{}
						baUsers := BloqueArchivo{}

						tiUsers.I_uid = I32toByte(1)
						tiUsers.I_gid = I32toByte(1)
						tiUsers.I_atime = I64toByte(time.Now().Unix())
						tiUsers.I_ctime = I64toByte(time.Now().Unix())
						tiUsers.I_mtime = I64toByte(time.Now().Unix())
						tiUsers.I_perm = I32toByte(700)
						tiUsers.I_block[0] = I32toByte(BytetoI32(sb.S_block_start[:]) + int32(binary.Size(bc)))
						for i := 1; i < 15; i++ {
							tiUsers.I_block[i] = I32toByte(-1)
						}
						s := "1,G,root\n1,U,root,root,123\n"
						tiUsers.I_size = I32toByte(int32(len(s)))
						tiUsers.I_type = '1'

						archivo.Seek(int64(BytetoI32(sb.S_inode_start[:])+int32(binary.Size(ti))), 0)

						bs.Reset()
						binary.Write(&bs, binary.BigEndian, tiUsers)
						_, _ = archivo.Write(bs.Bytes())

						for i, char := range s {
							baUsers.B_content[i] = byte(char)
						}
						archivo.Seek(int64(BytetoI32(sb.S_block_start[:])+int32(binary.Size(bc))), 0)

						bs.Reset()
						binary.Write(&bs, binary.BigEndian, baUsers)
						_, _ = archivo.Write(bs.Bytes())

						mbr.Mbr_partition[p].Part_status = '2'

						//Actualizacion del MBR y escritura del superbloque en disco
						archivo.Seek(0, 0)
						bs.Reset()
						binary.Write(&bs, binary.BigEndian, mbr)
						_, _ = archivo.Write(bs.Bytes())

						archivo.Seek(int64(BytetoI32(mbr.Mbr_partition[p].Part_start[:])), 0)
						bs.Reset()
						binary.Write(&bs, binary.BigEndian, sb)
						_, _ = archivo.Write(bs.Bytes())

						//Escritura de los BITMAP
						var ch0, ch1 byte
						ch0 = '0'
						ch1 = '1'
						archivo.Seek(int64(BytetoI32(sb.S_bm_inode_start[:])), 0)
						for i := 0; i < int(num_estructura); i++ {
							bs.Reset()
							binary.Write(&bs, binary.BigEndian, ch0)
							_, _ = archivo.Write(bs.Bytes())
						}
						archivo.Seek(int64(BytetoI32(sb.S_bm_inode_start[:])), 0)
						bs.Reset()
						binary.Write(&bs, binary.BigEndian, ch1)
						_, _ = archivo.Write(bs.Bytes())
						bs.Reset()
						binary.Write(&bs, binary.BigEndian, ch1)
						_, _ = archivo.Write(bs.Bytes())

						archivo.Seek(int64(BytetoI32(sb.S_bm_block_start[:])), 0)
						for i := 0; i < int(bloques); i++ {
							bs.Reset()
							binary.Write(&bs, binary.BigEndian, ch0)
							_, _ = archivo.Write(bs.Bytes())
						}
						archivo.Seek(int64(BytetoI32(sb.S_bm_block_start[:])), 0)
						bs.Reset()
						binary.Write(&bs, binary.BigEndian, ch1)
						_, _ = archivo.Write(bs.Bytes())
						bs.Reset()
						binary.Write(&bs, binary.BigEndian, ch1)
						_, _ = archivo.Write(bs.Bytes())

						return "La particion se formateo exitosamente"
					} else {
						this.Eliminar(nodo.IdCompleto)
						return "No fue posible encontrar la Particion en el Disco... Se demontara la Particion"
					}
				} else if nodo.Part_type == 'L' { //Formatear Particion Logica
					ebr := EBR{}
					archivo.Seek(int64(nodo.Part_start), 0)
					binary.Read(extraerStruct(archivo, binary.Size(ebr)), binary.BigEndian, &ebr)

					if string(bytes.Trim(ebr.Part_name[:], "\000")) == nodo.Nombre_particion && ebr.Part_status != '0' {
						//Llenar de \0
						archivo.Seek(int64(BytetoI32(ebr.Part_start[:])+int32(binary.Size(ebr))), 0)
						var vacio byte
						vacio = '\000'
						for i := 0; i < int(BytetoI32(ebr.Part_start[:])-int32(binary.Size(ebr))); i++ {
							var bs bytes.Buffer
							binary.Write(&bs, binary.BigEndian, vacio)
							_, _ = archivo.Write(bs.Bytes())
						}

						//Calcular el valor de N y Numero de Estructuras
						var n float64
						n = float64(int(BytetoI32(ebr.Part_size[:]))-binary.Size(ebr)-binary.Size(sb)) / float64(4+(binary.Size(ti))+(3*binary.Size(bc)))
						sb.S_filesystem_type = I32toByte(2)

						num_estructura := int32(math.Floor(n))
						bloques := 3 * num_estructura

						//Inicializar SuperBloque
						sb.S_inodes_count = I32toByte(num_estructura)
						sb.S_blocks_count = I32toByte(bloques)
						sb.S_free_inodes_count = I32toByte(num_estructura - 2)
						sb.S_free_blocks_count = I32toByte(bloques - 2)
						sb.S_inode_size = I32toByte(int32(binary.Size(ti)))
						sb.S_block_size = I32toByte(int32(binary.Size(bc)))

						sb.S_bm_inode_start = I32toByte(BytetoI32(ebr.Part_start[:]) + int32(binary.Size(ebr)+binary.Size(sb)))

						sb.S_first_ino = I32toByte(2)
						sb.S_first_blo = I32toByte(2)
						sb.S_bm_block_start = I32toByte(BytetoI32(sb.S_bm_inode_start[:]) + num_estructura)
						sb.S_inode_start = I32toByte(BytetoI32(sb.S_bm_block_start[:]) + bloques)
						sb.S_block_start = I32toByte(BytetoI32(sb.S_inode_start[:]) + (num_estructura * int32(binary.Size(ti))))

						//Inicializar Inodo Raiz
						ti.I_uid = I32toByte(1)
						ti.I_gid = I32toByte(1)
						ti.I_atime = I64toByte(time.Now().Unix())
						ti.I_ctime = I64toByte(time.Now().Unix())
						ti.I_mtime = I64toByte(time.Now().Unix())
						ti.I_perm = I32toByte(664)
						ti.I_block[0] = sb.S_block_start
						for i := 1; i < 16; i++ {
							ti.I_block[i] = I32toByte(-1)
						}
						ti.I_type = '0'
						ti.I_size = I32toByte(0)

						archivo.Seek(int64(BytetoI32(sb.S_inode_start[:])), 0)

						var bs bytes.Buffer
						binary.Write(&bs, binary.BigEndian, ti)
						_, _ = archivo.Write(bs.Bytes())

						for i, char := range "." {
							bc.B_content[0].B_name[i] = byte(char)
						}
						bc.B_content[0].B_inodo = sb.S_inode_start
						for i, char := range ".." {
							bc.B_content[1].B_name[i] = byte(char)
						}
						bc.B_content[1].B_inodo = sb.S_inode_start
						for i, char := range "users.txt" {
							bc.B_content[2].B_name[i] = byte(char)
						}
						bc.B_content[2].B_inodo = I32toByte(BytetoI32(sb.S_inode_start[:]) + int32(binary.Size(ti)))
						bc.B_content[3].B_inodo = I32toByte(-1)
						archivo.Seek(int64(BytetoI32(sb.S_block_start[:])), 0)

						bs.Reset()
						binary.Write(&bs, binary.BigEndian, bc)
						_, _ = archivo.Write(bs.Bytes())

						//Generar Inodo y archivo Users.txt
						tiUsers := TablaInodo{}
						baUsers := BloqueArchivo{}

						tiUsers.I_uid = I32toByte(1)
						tiUsers.I_gid = I32toByte(1)
						tiUsers.I_atime = I64toByte(time.Now().Unix())
						tiUsers.I_ctime = I64toByte(time.Now().Unix())
						tiUsers.I_mtime = I64toByte(time.Now().Unix())
						tiUsers.I_perm = I32toByte(700)
						tiUsers.I_block[0] = I32toByte(BytetoI32(sb.S_block_start[:]) + int32(binary.Size(bc)))
						for i := 1; i < 15; i++ {
							tiUsers.I_block[i] = I32toByte(-1)
						}
						s := "1,G,root\n1,U,root,root,123\n"
						tiUsers.I_size = I32toByte(int32(len(s)))
						tiUsers.I_type = '1'

						archivo.Seek(int64(BytetoI32(sb.S_inode_start[:])+int32(binary.Size(ti))), 0)

						bs.Reset()
						binary.Write(&bs, binary.BigEndian, tiUsers)
						_, _ = archivo.Write(bs.Bytes())

						for i, char := range s {
							baUsers.B_content[i] = byte(char)
						}
						archivo.Seek(int64(BytetoI32(sb.S_block_start[:])+int32(binary.Size(bc))), 0)

						bs.Reset()
						binary.Write(&bs, binary.BigEndian, baUsers)
						_, _ = archivo.Write(bs.Bytes())

						ebr.Part_status = '2'

						//Escritura en disco del ebr y el superbloque
						archivo.Seek(int64(BytetoI32(ebr.Part_start[:])), 0)
						bs.Reset()
						binary.Write(&bs, binary.BigEndian, ebr)
						_, _ = archivo.Write(bs.Bytes())
						archivo.Seek(int64(BytetoI32(ebr.Part_start[:])+int32(binary.Size(ebr))), 0)
						bs.Reset()
						binary.Write(&bs, binary.BigEndian, sb)
						_, _ = archivo.Write(bs.Bytes())

						//Escritura de los BITMAP
						var ch0, ch1 byte
						ch0 = '0'
						ch1 = '1'
						archivo.Seek(int64(BytetoI32(sb.S_bm_inode_start[:])), 0)
						for i := 0; i < int(num_estructura); i++ {
							bs.Reset()
							binary.Write(&bs, binary.BigEndian, ch0)
							_, _ = archivo.Write(bs.Bytes())
						}
						archivo.Seek(int64(BytetoI32(sb.S_bm_inode_start[:])), 0)
						bs.Reset()
						binary.Write(&bs, binary.BigEndian, ch1)
						_, _ = archivo.Write(bs.Bytes())
						bs.Reset()
						binary.Write(&bs, binary.BigEndian, ch1)
						_, _ = archivo.Write(bs.Bytes())

						archivo.Seek(int64(BytetoI32(sb.S_bm_block_start[:])), 0)
						for i := 0; i < int(bloques); i++ {
							bs.Reset()
							binary.Write(&bs, binary.BigEndian, ch0)
							_, _ = archivo.Write(bs.Bytes())
						}
						archivo.Seek(int64(BytetoI32(sb.S_bm_block_start[:])), 0)
						bs.Reset()
						binary.Write(&bs, binary.BigEndian, ch1)
						_, _ = archivo.Write(bs.Bytes())
						bs.Reset()
						binary.Write(&bs, binary.BigEndian, ch1)
						_, _ = archivo.Write(bs.Bytes())

						return "La particion se formateo exitosamente"
					} else {
						this.Eliminar(nodo.IdCompleto)
						return "No fue posible encontrar la Particion en el Disco... Se demontara la Particion"

					}
				}
				archivo.Close()
			} else {
				this.Eliminar(nodo.IdCompleto)
				return "No fue posible encontrar el Disco... Se demontara la Particion"
			}
		}
		return "No se encontro una particion montada con el ID ingresado"
	}
	return "No se reconoce el parametro de formateo"
}
