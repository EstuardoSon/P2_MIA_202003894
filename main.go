package main

import (
	"strings"

	analizador "github.com/EstuardoSon/P2_MIA_202003894/Analizador"
)

func main() {
	var strComandos string
	strComandos = `
	mkdisk >size=5 >unit=M >path="/home/mis discos/Disco3.dsk" #Comentario aqui
	rmdisk >path="/home/mis discos/Disco4.dsk"
	fdisk >size=1 >type=L >unit=M >fit=BF >path="/mis discos/Disco3.dsk" >name="Particion3"
	fdisk >type=E >path=/home/Disco2.dsk >name=Part3 >Unit=K >size=200
	#Comentario Random
	mount >path=/home/Disco3.dk >name=Part2 >id=063a
	mkfs >type=full >id=061A
	login >user="mi usuario" >pwd="mi pwd" >id=062A
	login >user=root >pwd=123 >id=062A
	logout 
	mkgrp >name=usuarios
	mkgrp >name="grupo 1"
	rmgrp >name=usuarios
	mkusr >user=user1 >pwd=usuario >grp=usuarios
	mkfile >path=/home/user/docs/b.txt >r >cont=/home/Documents/b.txt
	mkdir >r >path=/home/user/docs/usac
	mkdir >path="/home/mis documentos/archivos diciembre"
	rep >id=061A >Path=/home/user/reports/reporte1.jpg >name=mbr
	Pause`

	Comandos := strings.Split(strComandos, "\n")

	for _, comando := range Comandos {
		if strings.TrimSpace(comando) != "" {
			a := &analizador.Analizador{Comando: strings.TrimSpace(comando)}
			a.Analizar()
		}
	}
	//strconv.Atoi(string(content.B_inodo))
	//string(content.B_name)
}
