package estructuras

import (
	"bytes"
	"encoding/binary"
	"os"
	"regexp"
)

// Devuelve un numero de acuerdo al tipo d eunidades de la particion
func verificarTamanioP(u string) int {
	if u == "b" {
		return 1
	} else if u == "k" || u == "" {
		return 1024
	} else if u == "m" {
		return 1024 * 1024
	}
	return -1
}

// Devuelve un numero de acuerdo al tipo de particion
func verificarTipoP(t string) int {
	if t == "p" || t == "" {
		return 1
	} else if t == "e" {
		return 2
	} else if t == "l" {
		return 3
	}
	return -1
}

// Devuelve un numero deacuerdo al tipo de ajuste
func verificarAjusteP(f string) int {
	if f == "bf" {
		return 1
	} else if f == "ff" {
		return 2
	} else if f == "wf" || f == "" {
		return 3
	}
	return -1
}

// Obtener la posicion de inicio de la particion
func obtenerInicioP(path string) int {
	var file *os.File
	file, _ = os.OpenFile(path, os.O_RDWR, 0777)
	mbr := MBR{}

	binary.Read(extraerStruct(file, binary.Size(mbr)), binary.BigEndian, &mbr)

	file.Close()

	if BytetoI32(mbr.Mbr_partition[0].Part_start[:]) == -1 {
		return binary.Size(mbr)
	}
	for i := 1; i < 4; i++ {
		if BytetoI32(mbr.Mbr_partition[i].Part_start[:]) == -1 {
			return int(BytetoI32(mbr.Mbr_partition[i-1].Part_start[:]) + BytetoI32(mbr.Mbr_partition[i-1].Part_size[:]))
		}
	}

	return int(BytetoI32(mbr.Mbr_tamano[:]))
}

// Obtener el inicio de la particion I/E siguiente
func obtenerInincioSiguiente(mbr *MBR, posParticion int) int {
	for i := posParticion; i < 3; i++ {
		if BytetoI32(mbr.Mbr_partition[i+1].Part_start[:]) != -1 {
			return int(BytetoI32(mbr.Mbr_partition[i+1].Part_start[:]))
		}
	}
	return int(BytetoI32(mbr.Mbr_tamano[:]))
}

// Obtener la cantidad de bits libres despues de la particion
func obtenerSizeDP(mbr *MBR, name string) int {
	for i := 0; i < 4; i++ {
		if string(bytes.Trim(mbr.Mbr_partition[i].Part_name[:], "\000")) == name {
			return obtenerInincioSiguiente(mbr, i) - int(BytetoI32(mbr.Mbr_partition[i].Part_start[:])) - int(BytetoI32(mbr.Mbr_partition[i].Part_size[:]))
		}
	}
	return int(BytetoI32(mbr.Mbr_tamano[:])) - binary.Size(mbr)
}

// Buscar Particion Extendida
func buscarParticionE(mbr *MBR) (Partition, string) {
	for i := 0; i < 4; i++ {
		if mbr.Mbr_partition[i].Part_type == 'E' && mbr.Mbr_partition[i].Part_status != '0' {
			return mbr.Mbr_partition[i], ""
		}
	}
	p := Partition{}
	p.Part_status = '0'
	return p, "No existe una Particion Extendida dentro del disco"
}

// Crear Struct Partition
func crearStructPartition(tipo byte, f string, s int, path string, name string) Partition {
	p := Partition{}
	p.Part_status = '1'
	p.Part_type = tipo
	if verificarAjusteP(f) == 1 {
		p.Part_fit = 'B'
	} else if verificarAjusteP(f) == 2 {
		p.Part_fit = 'F'
	} else if verificarAjusteP(f) == 3 {
		p.Part_fit = 'W'
	}
	p.Part_start = I32toByte(int32(obtenerInicioP(path)))
	p.Part_size = I32toByte(int32(s))
	for i, char := range name {
		p.Part_name[i] = byte(char)
	}
	return p
}

// Verificar el nombre en particiones logicas
func verificarNombrePL(partition *Partition, tipo *byte, path string, name string) bool {
	ebr := EBR{}
	var file *os.File
	file, _ = os.OpenFile(path, os.O_RDWR, 0777)
	file.Seek(int64(BytetoI32(partition.Part_start[:])), 0)
	binary.Read(extraerStruct(file, binary.Size(ebr)), binary.BigEndian, &ebr)

	if string(bytes.Trim(ebr.Part_name[:], "\000")) == name {
		file.Close()
		*tipo = 'L'
		return true
	}
	for BytetoI32(ebr.Part_next[:]) != -1 {
		file.Seek(int64(BytetoI32(ebr.Part_next[:])), 0)
		binary.Read(extraerStruct(file, binary.Size(ebr)), binary.BigEndian, &ebr)

		if string(bytes.Trim(ebr.Part_name[:], "\000")) == name {
			file.Close()
			*tipo = 'L'
			return true
		}
	}

	return false
}

// Verificar nombre en particiones
func verificarNombreP(mbr *MBR, tipo *byte, path string, name string) bool {
	verificacion := false
	for i := 0; i < 4; i++ {
		if string(bytes.Trim(mbr.Mbr_partition[i].Part_name[:], "\000")) == name {
			*tipo = mbr.Mbr_partition[i].Part_type
			verificacion = true
			break
		} else if mbr.Mbr_partition[i].Part_type == 'E' {
			verificacion = verificarNombrePL(&mbr.Mbr_partition[i], tipo, path, name)
			if verificacion {
				break
			}
		}
	}
	return verificacion
}

// Verificar que exista espacio en el disco
func verificarEspacio(path string, tipoP int, s *int) bool {
	var file *os.File
	file, _ = os.OpenFile(path, os.O_RDWR, 0777)
	mbr := MBR{}
	file.Seek(0, 0)
	binary.Read(extraerStruct(file, binary.Size(mbr)), binary.BigEndian, &mbr)

	var ebr EBR
	if tipoP == 2 {
		*s = *s + binary.Size(ebr)
	}

	espacioD := int(BytetoI32(mbr.Mbr_tamano[:])) - binary.Size(mbr)
	if int(BytetoI32(mbr.Mbr_partition[0].Part_start[:])) == -1 {
		espacioD = obtenerInincioSiguiente(&mbr, 0) - binary.Size(mbr)
	} else {
		for i := 1; i < 4; i++ {
			if int(BytetoI32(mbr.Mbr_partition[i].Part_start[:])) == -1 {
				espacioD = obtenerInincioSiguiente(&mbr, i) - int(BytetoI32(mbr.Mbr_partition[i].Part_start[:])) - int(BytetoI32(mbr.Mbr_partition[i].Part_size[:]))
				break
			}
		}
	}
	file.Close()
	return espacioD >= *s
}

// Verificar que exista espacio en una partida Extendida
func verificarEspacioPE(p *Partition, nextEBR *EBR, path, name string, s int, f string) bool {
	var file *os.File
	file, _ = os.OpenFile(path, os.O_RDWR, 0777)
	ebrAux := EBR{}
	file.Seek(int64(BytetoI32(p.Part_start[:])), 0)
	binary.Read(extraerStruct(file, binary.Size(ebrAux)), binary.BigEndian, &ebrAux)

	for int(BytetoI32(ebrAux.Part_next[:])) != -1 {
		file.Seek(int64(BytetoI32(ebrAux.Part_next[:])), 0)
		binary.Read(extraerStruct(file, binary.Size(ebrAux)), binary.BigEndian, &ebrAux)
	}

	nextEBR.Part_next = I32toByte(-1)
	nextEBR.Part_start = I32toByte(BytetoI32(ebrAux.Part_start[:]) + BytetoI32(ebrAux.Part_size[:]))
	for i, char := range name {
		nextEBR.Part_name[i] = byte(char)
	}
	s = s + binary.Size(nextEBR)
	nextEBR.Part_size = I32toByte(int32(s))
	if verificarAjusteP(f) == 1 {
		nextEBR.Part_fit = 'B'
	} else if verificarAjusteP(f) == 2 {
		nextEBR.Part_fit = 'F'
	} else if verificarAjusteP(f) == 3 {
		nextEBR.Part_fit = 'W'
	}
	nextEBR.Part_status = '1'

	if int(BytetoI32(p.Part_size[:])) >= s {
		ebrAux.Part_next = nextEBR.Part_start
		file.Seek(int64(BytetoI32(ebrAux.Part_start[:])), 0)

		var bs bytes.Buffer
		binary.Write(&bs, binary.BigEndian, ebrAux)
		_, _ = file.Write(bs.Bytes())

		file.Close()
		return true
	} else {
		file.Close()
		return false
	}
}

// Verificar que no exitan otra partidas extendidas
func verificarPE(mbr *MBR) bool {
	if (mbr.Mbr_partition[0].Part_type == 'E' && mbr.Mbr_partition[0].Part_status != '0') || (mbr.Mbr_partition[1].Part_type == 'E' && mbr.Mbr_partition[1].Part_status != '0') || (mbr.Mbr_partition[2].Part_type == 'E' && mbr.Mbr_partition[2].Part_status != '0') || (mbr.Mbr_partition[3].Part_type == 'E' && mbr.Mbr_partition[3].Part_status != '0') {
		return true
	} else {
		return false
	}
}

func Fdisk(s int, u, p, t, f, name string) string {
	tamanio := verificarTamanio(u)
	ajusteP := verificarAjusteP(f)
	tipoP := verificarTipoP(t)

	//Verificacion de los parametros S | T | F | DELETE
	if tamanio == -1 || ajusteP == -1 || tipoP == -1 || len(name) == 0 || len(name) > 16 {
		return "No fue posible ejecutar el comando con la informacion proporcionada"
	}

	match, _ := regexp.MatchString("(\\/)([a-zA-Z0-9_ñÑáéíóúÁÉÍÓÚ ]+(\\/))*[a-zA-Z0-9_ñÑáéíóúÁÉÍÓÚ ]+(\\.dsk)", p)
	if !match {
		return "La direccion ingresada no es valida o la extension del archivo no es la adecuada"
	}

	var file *os.File
	file, _ = os.OpenFile(p, os.O_RDWR, 0777)

	if file != nil {
		mbr := MBR{}
		binary.Read(extraerStruct(file, binary.Size(mbr)), binary.BigEndian, &mbr)
		var tipoParticion byte
		tipoParticion = '\000'

		//Crear Particion
		if s >= 0 {
			if !(verificarNombreP(&mbr, &tipoParticion, p, name)) {
				s = s * tamanio

				//Particiones Principales y Extendidas
				if tipoP == 1 || tipoP == 2 {
					if verificarEspacio(p, tipoP, &s) {

						//Particion Primaria
						if tipoP == 1 {
							for i := 0; i < 4; i++ {
								if int(BytetoI32(mbr.Mbr_partition[i].Part_start[:])) == -1 {
									mbr.Mbr_partition[i] = crearStructPartition('P', f, s, p, name)
									file.Seek(0, 0)

									var bs bytes.Buffer
									binary.Write(&bs, binary.BigEndian, mbr)
									_, _ = file.Write(bs.Bytes())

									file.Close()
									return "Particion Primaria Creada"
								}
							}
							file.Close()
							return "Actualmente ya existen 4 particiones"
						} else if tipoP == 2 { //Particion Extendida
							if !verificarPE(&mbr) {
								ebr := EBR{}
								for x := 0; x < 4; x++ {
									if int(BytetoI32(mbr.Mbr_partition[x].Part_start[:])) == -1 {
										mbr.Mbr_partition[x] = crearStructPartition('E', f, s, p, name)
										ebr.Part_start = mbr.Mbr_partition[x].Part_start
										ebr.Part_next = I32toByte(-1)
										ebr.Part_status = '0'
										ebr.Part_size = I32toByte(0)
										for i := 0; i < binary.Size(ebr.Part_name)/binary.Size(ebr.Part_name[0]); i++ {
											ebr.Part_name[i] = '\000'
										}
										file.Seek(0, 0)
										var bs bytes.Buffer
										binary.Write(&bs, binary.BigEndian, mbr)
										_, _ = file.Write(bs.Bytes())

										file.Seek(int64(BytetoI32(ebr.Part_start[:])), 0)
										bs.Reset()
										binary.Write(&bs, binary.BigEndian, ebr)
										_, _ = file.Write(bs.Bytes())
										file.Close()
										return "Particion Extendida Creada"
									}
								}
								file.Close()
								return "Actualmente ya existen 4 particiones"
							}
						}
					} else {
						return "La particion excede el espacio disponible"
					}
				} else if tipoP == 3 { //Particion Logica
					particionE, err := buscarParticionE(&mbr)
					if err == "" {
						if particionE.Part_status != '0' {
							ebr := EBR{}
							if verificarEspacioPE(&particionE, &ebr, p, name, s, f) {
								file.Seek(0, 0)

								var bs bytes.Buffer
								binary.Write(&bs, binary.BigEndian, mbr)
								_, _ = file.Write(bs.Bytes())

								bs.Reset()
								file.Seek(int64(BytetoI32(ebr.Part_start[:])), 0)
								binary.Write(&bs, binary.BigEndian, ebr)
								_, _ = file.Write(bs.Bytes())

								return "Particion Logica Creada"
							} else {
								return "La particion logica excede el espacio disponible en la Particion Extendida"
							}
						}
					} else {
						return err
					}
				}
			} else {
				return "El nombre de la particion ya esta en uso"
			}
		} else {
			return "El valor de tamanio para crear la Particion debe ser positivo diferente de 0"
		}
		file.Close()
	}

	return "El archivo no fue encontrado en la direccion establecida"
}
