package estructuras

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

type Reporte struct {
	ListaMount *ListaMount
	Name       string
	Path       string
	Ruta       string
	Id         string
}

// Generacion de Reportes
func (this *Reporte) GenerarReporte() string {
	if this.Name == "disk" {
		return this.reporteDisk()
	} else if this.Name == "sb" {
		return this.reporteSb()
	} else if this.Name == "tree" {
		return this.reporteTree()
	} else if this.Name == "file" {
		return this.reporteFile()
	}
	return "No fue posible identificar el reporte deseado"
}

// Generar Reporte Disk
func (this *Reporte) reporteDisk() string {
	if this.Path != "" {
		nodo := this.ListaMount.Buscar(this.Id)

		if nodo != nil {
			archivoDisco, err := os.OpenFile(nodo.Fichero+"/"+nodo.Nombre_disco, os.O_RDONLY, 0777)

			if err == nil {
				//Creacion de los ficheros y dando permisos
				_, n_reporte := DivPath(this.Path)
				archivoReporte, _ := os.OpenFile("./Reportes/DOTS/reporteDisk.dot", os.O_RDWR|os.O_CREATE, 0777)

				mbr := MBR{}
				binary.Read(extraerStruct(archivoDisco, binary.Size(mbr)), binary.BigEndian, &mbr)
				dot := ""

				dot += "digraph G {\n"
				dot += "node[shape=none]\n"
				dot += "node[shape=none]\n"
				dot += "start[label=<<table><tr>"
				dot += "<td rowspan=\"2\">MBR</td>"

				tamanioT := int(BytetoI32(mbr.Mbr_tamano[:]))
				inicio := binary.Size(mbr)
				for i := 0; i < 4; i++ {
					if int(BytetoI32(mbr.Mbr_partition[i].Part_start[:])) != -1 {
						if mbr.Mbr_partition[i].Part_type == 'P' {
							p1 := (float64(BytetoI32(mbr.Mbr_partition[i].Part_size[:]))) / float64(tamanioT)
							porcentaje := p1 * 100.0
							name1 := string(bytes.Trim(mbr.Mbr_partition[i].Part_name[:], "\000"))
							dot += fmt.Sprintf("<td rowspan=\"2\">%s <br/>%.2f %%</td>", name1, porcentaje)
							if i != 3 {
								aux := i
								for i = i + 1; i < 4; i++ {
									if int(BytetoI32(mbr.Mbr_partition[i].Part_start[:])) != -1 {
										if (int(BytetoI32(mbr.Mbr_partition[aux].Part_start[:]) + BytetoI32(mbr.Mbr_partition[aux].Part_size[:]))) <
											int(BytetoI32(mbr.Mbr_partition[i].Part_start[:])) {
											porcentaje = (float64((BytetoI32(mbr.Mbr_partition[i].Part_start[:]) -
												(BytetoI32(mbr.Mbr_partition[aux].Part_start[:]) +
													BytetoI32(mbr.Mbr_partition[aux].Part_size[:])))) / float64(tamanioT))
											dot += fmt.Sprintf("<td rowspan=\"2\">LIBRE <br/>%.2f %%</td>",
												porcentaje)
											i = i - 1
											break
										} else if int(BytetoI32(mbr.Mbr_partition[aux].Part_start[:])+BytetoI32(mbr.Mbr_partition[aux].Part_size[:])) ==
											int(BytetoI32(mbr.Mbr_partition[i].Part_start[:])) {
											i = i - 1
											break
										}
									}
								}
								if i == 4 {
									porcentaje =
										(float64(int32(tamanioT)-
											(BytetoI32(mbr.Mbr_partition[aux].Part_start[:])+BytetoI32(mbr.Mbr_partition[aux].Part_size[:]))) /
											float64(tamanioT))
									porcentaje = porcentaje * 100.0
									dot += fmt.Sprintf("<td rowspan=\"2\">LIBRE <br/>%.2f %%</td>", porcentaje)
									goto t0
								}
							} else if (int(BytetoI32(mbr.Mbr_partition[i].Part_start[:]) + BytetoI32(mbr.Mbr_partition[i].Part_size[:]))) < tamanioT {
								porcentaje = (float64(int32(tamanioT)-
									(BytetoI32(mbr.Mbr_partition[i].Part_start[:])+BytetoI32(mbr.Mbr_partition[i].Part_size[:]))) / float64(tamanioT)) * 100
								dot += fmt.Sprintf("<td rowspan=\"2\">LIBRE <br/>%.2f %%</td>", porcentaje)
							}

						} else if mbr.Mbr_partition[i].Part_type == 'E' {
							porcentaje := (float64(BytetoI32(mbr.Mbr_partition[i].Part_size[:])) / float64(tamanioT)) * 100.0
							dot += "<td rowspan=\"2\">EXTENDIDA</td>"
							ebr := EBR{}
							ebrAux := EBR{}
							archivoDisco.Seek(int64(BytetoI32(mbr.Mbr_partition[i].Part_start[:])), 0)
							binary.Read(extraerStruct(archivoDisco, binary.Size(ebr)), binary.BigEndian, &ebr)

							if !(BytetoI32(ebr.Part_size[:]) == 0 && BytetoI32(ebr.Part_next[:]) == -1) {
								name1 := string(bytes.Trim(ebr.Part_name[:], "\000"))
								dot += "<td rowspan=\"2\">EBR <br/>" + name1 + "</td>"
								porcentaje = (float64(BytetoI32(ebr.Part_size[:])) / float64(tamanioT)) * 100.0
								dot += fmt.Sprintf("<td rowspan=\"2\">Logica <br/>%.2f %%</td>", porcentaje)

								finAnterior := 0
								for BytetoI32(ebr.Part_next[:]) != -1 {
									archivoDisco.Seek(int64(BytetoI32(ebr.Part_next[:])), 0)
									binary.Read(extraerStruct(archivoDisco, binary.Size(ebrAux)), binary.BigEndian, &ebrAux)

									finAnterior := int(BytetoI32(ebr.Part_size[:]) + BytetoI32(ebr.Part_start[:]))
									if int(BytetoI32(ebrAux.Part_start[:])) > finAnterior {
										porcentaje = (float64(int(BytetoI32(ebrAux.Part_start[:]))-finAnterior) / float64(tamanioT)) * 100.0
										dot += fmt.Sprintf("<td rowspan=\"2\">Libre <br/>%.2f %%</td>", porcentaje)
									}

									name1 = string(bytes.Trim(ebrAux.Part_name[:], "\000"))
									dot += "<td rowspan=\"2\">EBR <br/>" + name1 + "</td>"
									porcentaje = (float64(BytetoI32((ebrAux.Part_size[:]))) / float64(tamanioT)) * 100.0
									dot += fmt.Sprintf("<td rowspan=\"2\">Logica <br/>%.2f %%</td>", porcentaje)
									ebr = ebrAux
								}

								finAnterior = int(BytetoI32(ebr.Part_size[:]) + BytetoI32(ebr.Part_start[:]))
								if int(BytetoI32(mbr.Mbr_partition[i].Part_size[:])+BytetoI32(mbr.Mbr_partition[i].Part_start[:])) > finAnterior {
									porcentaje =
										(float64(BytetoI32(mbr.Mbr_partition[i].Part_size[:])+BytetoI32(mbr.Mbr_partition[i].Part_start[:])-int32(finAnterior)) / float64(tamanioT)) * 100.0
									dot += fmt.Sprintf("<td rowspan=\"2\">Libre <br/>%.2f %%</td>", porcentaje)
								}
							} else {
								dot += fmt.Sprintf("<td rowspan=\"2\">Libre <br/>%.2f %%</td>", porcentaje)
							}
							dot += "<td rowspan=\"2\">FIN EXTENDIDA</td>"
						}
						inicio = int(BytetoI32(mbr.Mbr_partition[i].Part_start[:]) + BytetoI32(mbr.Mbr_partition[i].Part_size[:]))
					} else {
						for i = i + 1; i < 4; i++ {
							if BytetoI32(mbr.Mbr_partition[i].Part_start[:]) != -1 {
								porcentaje :=
									(float64(int(BytetoI32(mbr.Mbr_partition[i].Part_start[:]))-inicio) / float64(tamanioT)) * 100.0
								dot += fmt.Sprintf("<td rowspan=\"2\">LIBRE <br/>%.2f %%</td>", porcentaje)
								i = i - 1
								break
							}
						}
						if i == 4 {
							porcentaje := ((float64(tamanioT-inicio) * 1.0) / float64(tamanioT)) * (100.0)
							dot += fmt.Sprintf("<td rowspan=\"2\">LIBRE <br/>%.2f %%</td>", porcentaje)
							goto t0
						}
					}
				}
			t0:
				dot += "</tr></table>>];\n"
				dot += "}"

				archivoReporte.Truncate(0)
				_, _ = archivoReporte.WriteString(dot)

				archivoDisco.Close()
				archivoReporte.Close()

				_, err := exec.Command("dot", "-T"+GetExtension(n_reporte), "Reportes/DOTS/reporteDisk.dot", "-o", "Reportes/"+nodo.IdCompleto+"_"+n_reporte).Output()
				if err != nil {
					return "No fue posible completar la creacion del Reporte DISK"
				}
				return "Reporte de DISK generado con Exito"
			} else {
				this.ListaMount.Eliminar(this.Id)
				return "No fue posible encontrar el Disco... "
			}
		}
		return "No fue posible encontrar una Particion con el ID ingresado"
	}
	return "Valor invalido en el parametro PATH"
}

// Generar Reporte Sb
func (this *Reporte) reporteSb() string {
	if this.Path != "" {
		nodo := this.ListaMount.Buscar(this.Id)

		if nodo != nil {
			archivoDisco, err := os.OpenFile(nodo.Fichero+"/"+nodo.Nombre_disco, os.O_RDONLY, 0777)

			if err == nil {
				//Creacion de los ficheros y dando permisos
				mbr := MBR{}
				dot := ""
				binary.Read(extraerStruct(archivoDisco, binary.Size(mbr)), binary.BigEndian, &mbr)

				sb := SuperBloque{}
				verificar := false
				res := ""
				for i := 0; i < 4; i++ {
					if string(bytes.Trim(mbr.Mbr_partition[i].Part_name[:], "\000")) == nodo.Nombre_particion &&
						nodo.Part_type == mbr.Mbr_partition[i].Part_type {
						if mbr.Mbr_partition[i].Part_status == '2' {
							archivoDisco.Seek(int64(BytetoI32(mbr.Mbr_partition[i].Part_start[:])), 0)
							binary.Read(extraerStruct(archivoDisco, binary.Size(sb)), binary.BigEndian, &sb)
							verificar = true
							break
						} else if mbr.Mbr_partition[i].Part_status == '1' {
							res = "No se ha aplicado el comando MKFS a la Particion"
							goto t0
						} else if mbr.Mbr_partition[i].Part_status == '0' {
							res = "No se encontro la Particion en el Disco... Desmontando la Praticion"
							this.ListaMount.Eliminar(nodo.IdCompleto)
							goto t0
						}
					} else if mbr.Mbr_partition[i].Part_type == 'E' &&
						nodo.Part_type == 'L' {
						ebr := EBR{}
						archivoDisco.Seek(int64(nodo.Part_start), 0)
						binary.Read(extraerStruct(archivoDisco, binary.Size(ebr)), binary.BigEndian, &ebr)

						if string(bytes.Trim(ebr.Part_name[:], "\000")) == nodo.Nombre_particion {
							if ebr.Part_status == '2' {
								archivoDisco.Seek(int64(BytetoI32(ebr.Part_start[:])+int32(binary.Size(ebr))), 0)
								binary.Read(extraerStruct(archivoDisco, binary.Size(sb)), binary.BigEndian, &sb)
								verificar = true
								break
							} else if ebr.Part_status == '1' {
								res = "No se ha aplicado el comando MKFS a la Particion"
								goto t0
							} else if ebr.Part_status == '0' {
								res = "No se encontro la Particion en el Disco... Desmontando la Praticion"
								this.ListaMount.Eliminar(nodo.IdCompleto)
								goto t0
							}
						}
					}
				}

				if verificar {
					_, n_reporte := DivPath(this.Path)
					archivoReporte, _ := os.OpenFile("./Reportes/DOTS/reporteSb.dot", os.O_RDWR|os.O_CREATE, 0777)
					dot += "digraph G {\n"
					dot += "node[shape=none]\n"
					dot += "start[label=<<table>\n"
					dot +=
						"<tr><td colspan=\"2\" bgcolor=\"#145a32\"><font color=\"white\">REPORTE DE SUPERBLOQUE</font></td></tr>\n"

					dot +=
						("<tr><td color=\"white\">sb_nombre_hd</td><td color=\"white\">" + nodo.Nombre_disco +
							"</td></tr>\n")

					dot +=
						fmt.Sprintf("<tr><td bgcolor=\"#27ae60\" color=\"white\">s_filesystem_type</td><td bgcolor=\"#27ae60\" color=\"white\">%d</td></tr>\n", int(BytetoI32(sb.S_filesystem_type[:])))

					dot +=
						fmt.Sprintf("<tr><td color=\"white\">s_inodes_count</td><td color=\"white\">%d</td></tr>\n", int(BytetoI32(sb.S_inodes_count[:])))

					dot +=
						fmt.Sprintf("<tr><td bgcolor=\"#27ae60\" color=\"white\">s_blocks_count</td><td bgcolor=\"#27ae60\" color=\"white\">%d</td></tr>\n", int(BytetoI32(sb.S_blocks_count[:])))

					dot +=
						fmt.Sprintf("<tr><td color=\"white\">s_free_blocks_count</td><td color=\"white\">%d</td></tr>\n", int(BytetoI32(sb.S_free_blocks_count[:])))

					dot +=
						fmt.Sprintf("<tr><td bgcolor=\"#27ae60\" color=\"white\">s_free_inodes_count</td><td bgcolor=\"#27ae60\" color=\"white\">%d</td></tr>\n", int(BytetoI32(sb.S_free_inodes_count[:])))

					fecha := time.Unix(BytetoI64(sb.S_mtime[:]), 0)
					dot +=
						("<tr><td color=\"white\">s_mtime</td><td color=\"white\">" +
							fecha.String() + "</td></tr>\n")

					dot +=
						fmt.Sprintf("<tr><td color=\"white\">s_mnt_count</td><td color=\"white\">%d</td></tr>\n", BytetoI32(sb.S_mnt_count[:]))

					dot +=
						fmt.Sprintf("<tr><td bgcolor=\"#27ae60\" color=\"white\">s_magic</td><td bgcolor=\"#27ae60\" color=\"white\">%d</td></tr>\n", int(BytetoI32(sb.S_magic[:])))

					dot +=
						fmt.Sprintf("<tr><td color=\"white\">s_inode_size</td><td color=\"white\">%d</td></tr>\n", BytetoI32(sb.S_inode_size[:]))

					dot +=
						fmt.Sprintf("<tr><td bgcolor=\"#27ae60\" color=\"white\">s_block_size</td><td bgcolor=\"#27ae60\" color=\"white\">%d</td></tr>\n", BytetoI32(sb.S_block_size[:]))

					dot +=
						fmt.Sprintf("<tr><td color=\"white\">s_first_ino</td><td color=\"white\">%d</td></tr>\n", BytetoI32(sb.S_first_ino[:]))
					dot +=
						fmt.Sprintf("<tr><td bgcolor=\"#27ae60\" color=\"white\">s_first_blo</td><td bgcolor=\"#27ae60\" color=\"white\">%d</td></tr>\n", BytetoI32(sb.S_first_blo[:]))

					dot +=
						fmt.Sprintf("<tr><td color=\"white\">s_bm_inode_start</td><td color=\"white\">%d</td></tr>\n", BytetoI32(sb.S_bm_inode_start[:]))

					dot +=
						fmt.Sprintf("<tr><td bgcolor=\"#27ae60\" color=\"white\">s_bm_block_start</td><td bgcolor=\"#27ae60\" color=\"white\">%d</td></tr>\n", BytetoI32(sb.S_bm_block_start[:]))

					dot +=
						fmt.Sprintf("<tr><td color=\"white\">s_inode_start</td><td color=\"white\">%d</td></tr>\n", BytetoI32(sb.S_inode_start[:]))

					dot +=
						fmt.Sprintf("<tr><td bgcolor=\"#27ae60\" color=\"white\">s_block_start</td><td bgcolor=\"#27ae60\" color=\"white\">%d</td></tr>\n", BytetoI32(sb.S_block_start[:]))

					dot += "</table>>];\n"
					dot += "}"

					archivoReporte.Truncate(0)
					_, _ = archivoReporte.WriteString(dot)

					archivoReporte.Close()
					_, err := exec.Command("dot", "-T"+GetExtension(n_reporte), "Reportes/DOTS/reporteSb.dot", "-o", "Reportes/"+nodo.IdCompleto+"_"+n_reporte).Output()
					if err != nil {
						res = "No fue posible completar la creacion del Reporte SB"
					} else {
						res = "Reporte de SB generado con Exito"
					}
				} else {
					res = "La Particion no exite dentro del disco... Desmontando la particion"
					this.ListaMount.Eliminar(nodo.IdCompleto)
				}
			t0:
				archivoDisco.Close()
				return res
			} else {
				this.ListaMount.Eliminar(this.Id)
				return "No fue posible encontrar el Disco... Desmontando particion"
			}
		}
		return "No fue posible encontrar una particion montada con el ID especificado"
	}

	return "Ingreso un valor erroneo en Path"
}

func (this *Reporte) buscarFichero(ficheros *[]string, sb *SuperBloque, inicioSB, inicioInodo int, archivo *os.File) int {
	ti := TablaInodo{}

	//Obtener Inodo
	archivo.Seek(int64(inicioInodo), 0)
	binary.Read(extraerStruct(archivo, binary.Size(ti)), binary.BigEndian, &ti)

	if len(*ficheros) > 0 {
		if ti.I_type == '0' {
			fichero := (*ficheros)[0]
			*ficheros = (*ficheros)[1:]
			ubicacion := this.buscarEnCarpeta(&ti, inicioInodo, archivo, fichero)

			if ubicacion != -1 {
				return this.buscarFichero(ficheros, sb, inicioSB, ubicacion, archivo)

			} else {
				return -1
			}
		} else {
			return -1
		}
	} else {
		return inicioInodo
	}
}

// Buscar una carpeta o archivo en una carpeta padre
func (this *Reporte) buscarEnCarpeta(ti *TablaInodo, inicioInodo int, archivo *os.File, nombre string) int {
	bc := BloqueCarpeta{}
	ubicacion := -1
	for i := 0; i < 16; i++ {
		if BytetoI32(ti.I_block[i][:]) != -1 {
			archivo.Seek(int64(BytetoI32(ti.I_block[i][:])), 0)
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

// Generar Reporte File
func (this *Reporte) reporteFile() string {
	if this.Path != "" && this.Ruta != "" {
		nodo := this.ListaMount.Buscar(this.Id)

		if nodo != nil {
			archivoDisco, err := os.OpenFile(nodo.Fichero+"/"+nodo.Nombre_disco, os.O_RDONLY, 0777)

			if err == nil {
				//Creacion de los ficheros y dando permisos
				mbr := MBR{}
				dot := ""
				binary.Read(extraerStruct(archivoDisco, binary.Size(mbr)), binary.BigEndian, &mbr)

				sb := SuperBloque{}
				inicioSB := -1
				res := ""
				for i := 0; i < 4; i++ {
					if string(bytes.Trim(mbr.Mbr_partition[i].Part_name[:], "\000")) == nodo.Nombre_particion &&
						nodo.Part_type == mbr.Mbr_partition[i].Part_type {
						if mbr.Mbr_partition[i].Part_status == '2' {
							archivoDisco.Seek(int64(BytetoI32(mbr.Mbr_partition[i].Part_start[:])), 0)
							binary.Read(extraerStruct(archivoDisco, binary.Size(sb)), binary.BigEndian, &sb)
							inicioSB = int(BytetoI32(mbr.Mbr_partition[i].Part_start[:]))
							break
						} else if mbr.Mbr_partition[i].Part_status == '1' {
							res = "No se ha aplicado el comando MKFS a la Particion"
							goto t0
						} else if mbr.Mbr_partition[i].Part_status == '0' {
							res = "No se encontro la Particion en el Disco... Desmontando la Praticion"
							this.ListaMount.Eliminar(nodo.IdCompleto)
							goto t0
						}
					} else if mbr.Mbr_partition[i].Part_type == 'E' &&
						nodo.Part_type == 'L' {
						ebr := EBR{}
						archivoDisco.Seek(int64(nodo.Part_start), 0)
						binary.Read(extraerStruct(archivoDisco, binary.Size(ebr)), binary.BigEndian, &ebr)

						if string(bytes.Trim(ebr.Part_name[:], "\000")) == nodo.Nombre_particion {
							if ebr.Part_status == '2' {
								archivoDisco.Seek(int64(BytetoI32(ebr.Part_start[:])+int32(binary.Size(ebr))), 0)
								binary.Read(extraerStruct(archivoDisco, binary.Size(sb)), binary.BigEndian, &sb)
								inicioSB = nodo.Part_start + binary.Size(ebr)
								break
							} else if ebr.Part_status == '1' {
								res = "No se ha aplicado el comando MKFS a la Particion"
								goto t0
							} else if ebr.Part_status == '0' {
								res = "No se encontro la Particion en el Disco... Desmontando la Praticion"
								this.ListaMount.Eliminar(nodo.IdCompleto)
								goto t0
							}
						}
					}
				}

				dot = ""

				if inicioSB != -1 {
					f_ruta, n_ruta := DivPath(this.Ruta)
					_, n_archivo := DivPath(this.Path)
					archivoReporte, _ := os.OpenFile("./Reportes/"+nodo.IdCompleto+"_"+n_archivo, os.O_RDWR|os.O_CREATE, 0777)

					ficheros := strings.Split(f_ruta[1:], "/")
					if n_ruta != "" {
						ficheros = append(ficheros, n_ruta)
					}

					ubicacion := this.buscarFichero(&ficheros, &sb, inicioSB, int(BytetoI32(sb.S_inode_start[:])), archivoDisco)

					if ubicacion != -1 {
						dot += GetContentF(ubicacion, archivoDisco)
					}

					_ = archivoReporte.Truncate(0)
					_, _ = archivoReporte.WriteString(dot)

					archivoReporte.Close()
					res = "Reporte de File generado con Exito"
				} else {
					this.ListaMount.Eliminar(nodo.IdCompleto)
					res = "La Particion no exite dentro del disco... Desmontando la particion"
				}
			t0:
				archivoDisco.Close()
				return res
			} else {
				this.ListaMount.Eliminar(this.Id)
				return "No fue posible encontrar el Disco... Desmontando particion"
			}
		}
		return "No fue posible encontrar una particion montada con el ID especificado"
	}
	return "Ingreso valores invalidos en Path o Ruta"
}

// Generar Reporte Tree
func (this *Reporte) reporteTree() string {
	if this.Path != "" {
		nodo := this.ListaMount.Buscar(this.Id)

		if nodo != nil {
			archivoDisco, err := os.OpenFile(nodo.Fichero+"/"+nodo.Nombre_disco, os.O_RDONLY, 0777)

			if err == nil {
				//Creacion de los ficheros y dando permisos
				mbr := MBR{}
				dot := ""
				binary.Read(extraerStruct(archivoDisco, binary.Size(mbr)), binary.BigEndian, &mbr)

				sb := SuperBloque{}
				verificar := false
				res := ""
				for i := 0; i < 4; i++ {
					if string(bytes.Trim(mbr.Mbr_partition[i].Part_name[:], "\000")) == nodo.Nombre_particion &&
						nodo.Part_type == mbr.Mbr_partition[i].Part_type {
						if mbr.Mbr_partition[i].Part_status == '2' {
							archivoDisco.Seek(int64(BytetoI32(mbr.Mbr_partition[i].Part_start[:])), 0)
							binary.Read(extraerStruct(archivoDisco, binary.Size(sb)), binary.BigEndian, &sb)
							verificar = true
							break
						} else if mbr.Mbr_partition[i].Part_status == '1' {
							res = "No se ha aplicado el comando MKFS a la Particion"
							goto t0
						} else if mbr.Mbr_partition[i].Part_status == '0' {
							res = "No se encontro la Particion en el Disco... Desmontando la Praticion"
							this.ListaMount.Eliminar(nodo.IdCompleto)
							goto t0
						}
					} else if mbr.Mbr_partition[i].Part_type == 'E' &&
						nodo.Part_type == 'L' {
						ebr := EBR{}
						archivoDisco.Seek(int64(nodo.Part_start), 0)
						binary.Read(extraerStruct(archivoDisco, binary.Size(ebr)), binary.BigEndian, &ebr)

						if string(bytes.Trim(ebr.Part_name[:], "\000")) == nodo.Nombre_particion {
							if ebr.Part_status == '2' {
								archivoDisco.Seek(int64(BytetoI32(ebr.Part_start[:])+int32(binary.Size(ebr))), 0)
								binary.Read(extraerStruct(archivoDisco, binary.Size(sb)), binary.BigEndian, &sb)
								verificar = true
								break
							} else if ebr.Part_status == '1' {
								res = "No se ha aplicado el comando MKFS a la Particion"
								goto t0
							} else if ebr.Part_status == '0' {
								res = "No se encontro la Particion en el Disco... Desmontando la Praticion"
								this.ListaMount.Eliminar(nodo.IdCompleto)
								goto t0
							}
						}
					}
				}

				if verificar {
					_, n_reporte := DivPath(this.Path)
					archivoReporte, _ := os.OpenFile("./Reportes/DOTS/reporteTree.dot", os.O_RDWR|os.O_CREATE, 0777)
					conexiones := ""

					dot += "digraph G {\n"
					dot += "rankdir=LR;\n"
					dot += "node[shape=none]\n"

					for a := 0; a < int(BytetoI32(sb.S_inodes_count[:])); a++ {
						var caracter byte
						ti := TablaInodo{}

						archivoDisco.Seek(int64(BytetoI32(sb.S_bm_inode_start[:])+int32(a)), 0)
						binary.Read(extraerStruct(archivoDisco, binary.Size(caracter)), binary.BigEndian, &caracter)

						if caracter == '1' {
							archivoDisco.Seek(int64(BytetoI32(sb.S_inode_start[:])+(int32(a)*int32(binary.Size(ti)))), 0)
							binary.Read(extraerStruct(archivoDisco, binary.Size(ti)), binary.BigEndian, &ti)

							dot += this.treeInodo(int(BytetoI32(sb.S_inode_start[:])+(int32(a)*int32(binary.Size(ti)))), a, archivoDisco, archivoReporte, &conexiones)
							for i := 0; i < 16; i++ {
								if BytetoI32(ti.I_block[i][:]) != -1 {
									if ti.I_type == '0' {
										dot += this.treeCarpeta(int(BytetoI32(ti.I_block[i][:])), archivoDisco, archivoReporte, &conexiones)
									} else if ti.I_type == '1' {
										dot += this.treeArchivo(int(BytetoI32(ti.I_block[i][:])), archivoDisco, archivoReporte, &conexiones)
									}

								}
							}
						}
					}
					dot += conexiones
					dot += "}"
					_ = archivoReporte.Truncate(0)
					_, _ = archivoReporte.WriteString(dot)

					archivoReporte.Close()
					_, err := exec.Command("dot", "-T"+GetExtension(n_reporte), "Reportes/DOTS/reporteTree.dot", "-o", "Reportes/"+nodo.IdCompleto+"_"+n_reporte).Output()
					if err != nil {
						res = "No fue posible completar la creacion del Reporte Tree\n"
						res += err.Error()
					} else {
						res = "Reporte de Tree generado con Exito"
					}
				} else {
					this.ListaMount.Eliminar(nodo.IdCompleto)
					return "La Particion no exite dentro del disco... Desmontando la particion"
				}
			t0:
				archivoDisco.Close()
				return res
			} else {
				this.ListaMount.Eliminar(this.Id)
				return "No fue posible encontrar el Disco... Desmontando particion"
			}
		}

		return "No fue posible encontrar una particion montada con el ID especificado"
	}
	return "Ingreso un valor invalido en PATH"
}

func (this *Reporte) treeInodo(posicion, noInodo int, archivoDisco *os.File, archivoReporte *os.File, conexiones *string) string {
	dot := ""

	ti := TablaInodo{}
	archivoDisco.Seek(int64(posicion), 0)
	binary.Read(extraerStruct(archivoDisco, binary.Size(ti)), binary.BigEndian, &ti)

	dot += fmt.Sprintf("n%d[label=<<table><tr><td colspan=\"2\" bgcolor=\"#376ef3\">INODO %d</td></tr>\n", posicion, noInodo)

	dot += "<tr>\n"
	dot += "<td>i_uid</td>\n"
	dot += fmt.Sprintf("<td>%d</td>\n", BytetoI32(ti.I_uid[:]))
	dot += "</tr>\n"

	dot += "<tr>\n"
	dot += "<td>i_gid</td>\n"
	dot += fmt.Sprintf("<td>%d</td>\n", BytetoI32(ti.I_gid[:]))
	dot += "</tr>\n"

	dot += "<tr>\n"
	dot += "<td>i_s</td>\n"
	dot += fmt.Sprintf("<td>%d</td>\n", BytetoI32(ti.I_size[:]))
	dot += "</tr>\n"

	fecha := time.Unix(BytetoI64(ti.I_atime[:]), 0)
	dot += "<tr>\n"
	dot += "<td>i_atime</td>\n"
	dot += ("<td>" + fecha.String() + "</td>\n")
	dot += "</tr>\n"

	fecha = time.Unix(BytetoI64(ti.I_ctime[:]), 0)
	dot += "<tr>\n"
	dot += "<td>i_ctime</td>\n"
	dot += ("<td>" + fecha.String() + "</td>\n")
	dot += "</tr>\n"

	fecha = time.Unix(BytetoI64(ti.I_mtime[:]), 0)
	dot += "<tr>\n"
	dot += "<td>i_mtime</td>\n"
	dot += ("<td>" + fecha.String() + "</td>\n")
	dot += "</tr>\n"

	for j := 0; j < 16; j++ {
		if BytetoI32(ti.I_block[j][:]) != -1 {
			*conexiones += fmt.Sprintf("n%d -> n%d\n", posicion, BytetoI32(ti.I_block[j][:]))
			dot += "<tr>\n"
			dot += fmt.Sprintf("<td>ap%d</td>\n", j)
			dot += fmt.Sprintf("<td port=\"%d\">%d</td>\n", BytetoI32(ti.I_block[j][:]), BytetoI32(ti.I_block[j][:]))
			dot += "</tr>\n"
		} else {
			dot += "<tr>\n"
			dot += "<td>i_block</td>\n"
			dot += "<td>-1</td>\n"
			dot += "</tr>\n"
		}

	}

	dot += "<tr>\n"
	dot += "<td>i_type</td>\n"
	dot += ("<td>" + string(ti.I_type) + "</td>\n")
	dot += "</tr>\n"

	dot += "<tr>\n"
	dot += "<td>i_perm</td>\n"
	dot += fmt.Sprintf("<td>%d</td>\n", BytetoI32(ti.I_perm[:]))
	dot += "</tr>\n"

	dot += "</table>>]\n"
	return dot
}

func (this *Reporte) treeArchivo(posicion int, archivoDisco *os.File, archivoReporte *os.File, conexiones *string) string {
	content := ""
	dot := ""
	archivo := BloqueArchivo{}
	archivoDisco.Seek(int64(posicion), 0)
	binary.Read(extraerStruct(archivoDisco, binary.Size(archivo)), binary.BigEndian, &archivo)

	for i := 0; i < 64; i++ {
		if archivo.B_content[i] == '\000' {
			break
		}
		content += string(archivo.B_content[i])
	}

	dot += fmt.Sprintf("n%d[label=<<table>\n", posicion)
	dot += "<tr>\n"
	dot += "<td colspan=\"2\" bgcolor=\"#c3f8b6\">Bloque Archivo</td>"
	dot += "</tr>\n<tr>\n"
	dot += ("<td>" + content + "</td>\n")
	dot += "</tr>\n</table>>]\n"
	return dot
}

func (this *Reporte) treeCarpeta(posicion int, archivoDisco *os.File, archivoReporte *os.File, conexiones *string) string {
	carpeta := BloqueCarpeta{}
	dot := ""
	archivoDisco.Seek(int64(posicion), 0)
	binary.Read(extraerStruct(archivoDisco, binary.Size(carpeta)), binary.BigEndian, &carpeta)

	dot += fmt.Sprintf("n%d[label=<<table>\n", posicion)
	dot += "<tr>\n"
	dot += "<td colspan=\"2\" bgcolor=\"#f34037\">Bloque Carpeta</td>"
	dot += "</tr>\n"
	for i := 0; i < 4; i++ {
		if BytetoI32(carpeta.B_content[i].B_inodo[:]) != -1 {
			*conexiones += fmt.Sprintf("n%d -> n%d\n", posicion, BytetoI32(carpeta.B_content[i].B_inodo[:]))
		}

		dot += "<tr>\n"
		dot += ("<td>" + string(bytes.Trim(carpeta.B_content[i].B_name[:], "\000")) + "</td>\n")
		dot += fmt.Sprintf("<td port=\"%d\">%d</td>\n", BytetoI32(carpeta.B_content[i].B_inodo[:]), BytetoI32(carpeta.B_content[i].B_inodo[:]))
		dot += "</tr>\n"
	}
	dot += "</table>>]\n"

	return dot
}
