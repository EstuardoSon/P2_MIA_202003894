package estructuras

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"regexp"
	"time"
)

func verificarAjusteD(f string) int {
	if f == "bf" {
		return 1
	} else if f == "ff" || f == "" {
		return 2
	} else if f == "wf" {
		return 3
	}
	return -1
}

func verificarTamanio(u string) int {
	if u == "m" || u == "" {
		return 1024
	} else if u == "k" {
		return 1
	}
	return -1
}

func Mkdisk(s int, path, f, u string) string {
	//Verificacion de datos para la cracion del disco
	tamanio := verificarTamanio(u)
	ajusteD := verificarAjusteD(f)
	if s <= 0 || ajusteD == -1 || tamanio == -1 {
		return "No fue posible crear el disco con la informacion proporcionada"

	}

	match, _ := regexp.MatchString("(\\/[a-zA-Z0-9_ñÑáéíóúÁÉÍÓÚ ]+\\/)([a-zA-Z0-9_ñÑáéíóúÁÉÍÓÚ ]+(\\/))*[a-zA-Z0-9_ñÑáéíóúÁÉÍÓÚ ]+(\\.dsk)", path)

	if !match {
		return "La direccion ingresada no es valida o la extension del archivo no es la adecuada"
	}

	//Verificacion de la existencia del archivo
	var file *os.File
	file, _ = os.OpenFile(path, os.O_RDONLY, 0777)

	if file != nil {
		file.Close()
		return "El archivo ya existe"
	}

	//Creacion de los ficheros y dando permisos
	directorios, nombre := DivPath(path)
	err := os.MkdirAll(directorios, 0777)

	if err != nil {
		return fmt.Sprintf("%s", err)
	}

	//Creacion del Archivo y llenandolo de caracteres vacios
	file, err = os.OpenFile(directorios+"/"+nombre, os.O_CREATE|os.O_RDWR, 0664)
	err = os.Chmod(path, 0777)

	if err != nil {
		return fmt.Sprintf("%s", err)
	}

	var buff [1024]byte
	for i := 0; i < 1024; i++ {
		buff[i] = 0
	}

	s = s * tamanio
	for i := 0; i < s; i++ {
		var bs bytes.Buffer
		binary.Write(&bs, binary.BigEndian, buff)
		_, _ = file.Write(bs.Bytes())
	}

	//Regresar el buffer de lectura del archivo al inicio
	file.Seek(0, 0)

	//Crear el MBR para
	var mbr MBR
	mbr.Mbr_tamano = I32toByte(1024 * int32(s))
	mbr.Mbr_fecha_creacion = I64toByte(time.Now().Unix())
	mbr.Mbr_dsk_signature = I64toByte(time.Now().Unix())
	if ajusteD == 1 {
		mbr.Dsk_fit = '0'
	} else if ajusteD == 2 {
		mbr.Dsk_fit = 'F'
	} else if ajusteD == 3 {
		mbr.Dsk_fit = 'W'
	}
	for i := 0; i < 4; i++ {
		mbr.Mbr_partition[i].Part_status = '0'
		mbr.Mbr_partition[i].Part_start = I32toByte(-1)
		mbr.Mbr_partition[i].Part_size = I32toByte(0)
	}

	var bs bytes.Buffer
	binary.Write(&bs, binary.BigEndian, mbr)
	file.Write(bs.Bytes())

	file.Close()

	//cout << this->obtenerNombreRaid() << endl
	return fmt.Sprintf("- Disco creado con exito -")
}

func Rmdisk(path string) string {
	match, _ := regexp.MatchString("(\\/[a-zA-Z0-9_ñÑáéíóúÁÉÍÓÚ ]+\\/)([a-zA-Z0-9_ñÑáéíóúÁÉÍÓÚ ]+(\\/))*[a-zA-Z0-9_ñÑáéíóúÁÉÍÓÚ ]+(\\.dsk)", path)
	if !match {
		return "La direccion ingresada no es valida o la extension del archivo no es la adecuada"
	}

	//Verificando la existencia del archivo
	file, _ := os.OpenFile(path, os.O_RDONLY, 0777)

	if file != nil {
		file.Close()
		os.Remove(path)
		return "Archivo eliminado con Exito"
	} else {
		return "El archivo especificado no existe"
	}
}
