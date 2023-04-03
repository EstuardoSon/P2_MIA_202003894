package estructuras

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"strings"
	"time"
)

type Partition struct {
	Part_status byte
	Part_type   byte
	Part_fit    byte
	Part_start  [4]byte
	Part_size   [4]byte
	Part_name   [16]byte
}

type MBR struct {
	Mbr_tamano         [4]byte
	Mbr_fecha_creacion [8]byte
	Mbr_dsk_signature  [8]byte
	Dsk_fit            byte
	Mbr_partition      [4]Partition
}

type EBR struct {
	Part_status byte
	Part_fit    byte
	Part_start  [4]byte
	Part_size   [4]byte
	Part_next   [4]byte
	Part_name   [16]byte
}

type SuperBloque struct {
	S_filesystem_type   [4]byte
	S_inodes_count      [4]byte
	S_blocks_count      [4]byte
	S_free_blocks_count [4]byte
	S_free_inodes_count [4]byte
	S_mtime             [8]byte
	S_mnt_count         [4]byte
	S_magic             [4]byte
	S_inode_size        [4]byte
	S_block_size        [4]byte
	S_first_ino         [4]byte
	S_first_blo         [4]byte
	S_bm_inode_start    [4]byte
	S_bm_block_start    [4]byte
	S_inode_start       [4]byte
	S_block_start       [4]byte
}

type TablaInodo struct {
	I_uid   [4]byte
	I_gid   [4]byte
	I_size  [4]byte
	I_atime [8]byte
	I_ctime [8]byte
	I_mtime [8]byte
	I_block [16][4]byte
	I_type  byte
	I_perm  [4]byte
}

type Content struct {
	B_name  [12]byte
	B_inodo [4]byte
}

type BloqueCarpeta struct {
	B_content [4]Content
}

type BloqueArchivo struct {
	B_content [64]byte
}

func I32toByte(numero int32) [4]byte {
	var buff bytes.Buffer
	binary.Write(&buff, binary.BigEndian, numero)
	return [4]byte(buff.Bytes())
}

func BytetoI32(arr []byte) int32 {
	var n int32
	binary.Read(bytes.NewBuffer(arr), binary.BigEndian, &n)
	return n
}

func I64toByte(numero int64) [8]byte {
	var buff bytes.Buffer
	binary.Write(&buff, binary.BigEndian, numero)
	return [8]byte(buff.Bytes())
}

func BytetoI64(arr []byte) int64 {
	var n int64
	binary.Read(bytes.NewBuffer(arr), binary.BigEndian, &n)
	return n
}

func DivPath(path string) (directorios, nombre string) {
	pos := strings.LastIndex(path, "/")

	if pos != -1 {
		nombre := path[pos+1:]
		directorios := path[:pos]

		return directorios, nombre
	}
	return "", path
}

func extraerStruct(file *os.File, number int) *bytes.Buffer {
	data := make([]byte, number)
	_, err := file.Read(data)
	if err != nil {
		fmt.Println(err)
	}
	buff := bytes.NewBuffer(data)
	return buff
}

// Obtener la informacion de un archivo en disco
func GetContentF(inicioInodo int, archivo *os.File) string {
	ti := TablaInodo{}
	ba := BloqueArchivo{}
	contenido := ""

	//Obtener Inodo
	archivo.Seek(int64(inicioInodo), 0)
	binary.Read(extraerStruct(archivo, binary.Size(ti)), binary.BigEndian, &ti)

	//Recorrer Array de Bloques de inodo
	for i := 0; i < 16; i++ {
		if BytetoI32(ti.I_block[i][:]) != -1 {
			archivo.Seek(int64(BytetoI32(ti.I_block[i][:])), 0)
			binary.Read(extraerStruct(archivo, binary.Size(ba)), binary.BigEndian, &ba)
			for j := 0; j < 64; j++ {
				if ba.B_content[j] != '\000' {
					contenido += string(ba.B_content[j])
					continue
				}
				break
			}
		}
	}

	ti.I_atime = I64toByte(time.Now().Unix())

	archivo.Seek(int64(inicioInodo), 0)
	var bs bytes.Buffer
	binary.Write(&bs, binary.BigEndian, ti)
	_, _ = archivo.Write(bs.Bytes())

	return contenido
}

// Separar en vectores los grupos y usuarios
func ObtenerUG(cadena string, usuarios *[]string, grupos *[]string) {
	info := strings.Split(cadena, "\n")
	for i := 0; i < len(info); i++ {
		aux := strings.Split(info[i], ",")
		if len(aux) == 3 {
			(*grupos) = append(*grupos, info[i])
		} else if len(aux) == 5 {
			(*usuarios) = append(*usuarios, info[i])
		}
	}
}
