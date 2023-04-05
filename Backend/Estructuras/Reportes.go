package estructuras

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"os/exec"
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
		//return this.reporteTree()
	} else if this.Name == "file" {
		//return this.reporteFile()
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

				_, _ = archivoReporte.WriteString(dot)

				archivoDisco.Close()
				archivoReporte.Close()

				_, err := exec.Command("dot", "-T"+GetExtension(n_reporte), "Reportes/DOTS/reporteDisk.dot", "-o", "Reportes/"+nodo.IdCompleto+"_DISK."+GetExtension(n_reporte)).Output()
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

				dot = ""
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

					_, _ = archivoReporte.WriteString(dot)

					archivoReporte.Close()
					_, err := exec.Command("dot", "-T"+GetExtension(n_reporte), "Reportes/DOTS/reporteSb.dot", "-o", "Reportes/"+nodo.IdCompleto+"_SB."+GetExtension(n_reporte)).Output()
					if err != nil {
						res = "No fue posible completar la creacion del Reporte SB"
					} else {
						res = "Reporte de SB generado con Exito"
					}
				} else {
					res = "La Particion no exite dentro del disco... Desmontando la particion"
					this.ListaMount.Eliminar(nodo.Id)
				}
			t0:
				archivoDisco.Close()
				return res
			} else {
				this.ListaMount.Eliminar(this.Id)
				return "No fue posible encontrar el Disco... Desmontando particion"
			}
		}
	}

	return "Ingreso un valor erroneo en Path"
}
