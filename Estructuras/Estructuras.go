package estructuras

type Partition struct {
	Part_status []byte
	Part_type   []byte
	Part_fit    []byte
	Part_start  []byte
	Part_size   []byte
	Part_name   []byte
}

type MBR struct {
	Mbr_tamano         []byte
	Mbr_fecha_creacion []byte
	Mbr_dsk_signature  []byte
	Dsk_fit            []byte
	Mbr_partition      [4]Partition
}

type EBR struct {
	part_status []byte
	part_fit    []byte
	part_start  []byte
	part_size   []byte
	part_next   []byte
	part_name   []byte
}

type SuperBloque struct {
	S_filesystem_type   []byte
	S_inodes_count      []byte
	S_blocks_count      []byte
	S_free_blocks_count []byte
	S_free_inodes_count []byte
	S_mtime             []byte
	S_mnt_count         []byte
	S_magic             []byte
	S_inode_size        []byte
	S_block_size        []byte
	S_firts_ino         []byte
	S_first_blo         []byte
	S_bm_inode_start    []byte
	S_bm_block_start    []byte
	S_inode_start       []byte
	S_block_start       []byte
}

type TablaInodo struct {
	I_uid   []byte
	I_gid   []byte
	I_size  []byte
	I_atime []byte
	I_ctime []byte
	I_mtime []byte
	I_block [16]byte
	I_type  []byte
	I_perm  []byte
}

type Content struct {
	B_name  []byte
	B_inodo []byte
}

type BloqueCarpeta struct {
	B_content []Content
}

type BloqueArchivo struct {
	B_content []byte
}
