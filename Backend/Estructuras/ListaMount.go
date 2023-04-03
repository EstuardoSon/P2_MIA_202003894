package estructuras

import (
	"bytes"
	"encoding/binary"
	"fmt"
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
