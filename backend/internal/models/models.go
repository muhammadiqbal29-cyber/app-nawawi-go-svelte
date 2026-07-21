package models

import "time"

type User struct {
	ID        int       `gorm:"primaryKey;autoIncrement" json:"id"`
	Email     string    `gorm:"type:varchar(255);unique;not null" json:"email"`
	Password  string    `gorm:"type:varchar(255);not null" json:"-"`
	Role      string    `gorm:"type:varchar(50);not null" json:"role"`
	CreatedAt time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"created_at"`
}

type Sekolah struct {
	ID            int       `gorm:"primaryKey;autoIncrement" json:"id"`
	NamaSekolah   string    `gorm:"type:varchar(255);not null" json:"nama_sekolah" validate:"required,max=255"`
	NPSN          string    `gorm:"type:varchar(8);unique;not null" json:"npsn" validate:"required,len=8,numeric"`
	Jenjang       string    `gorm:"type:varchar(50);not null" json:"jenjang" validate:"required,oneof=TK SD SMP SMA SMK MI MTs MA"`
	Status        string    `gorm:"type:varchar(50);not null" json:"status" validate:"required,oneof=Negeri Swasta"`
	Alamat        string    `gorm:"type:text;not null" json:"alamat" validate:"required"`
	KodePos       string    `gorm:"type:varchar(10)" json:"kode_pos"`
	Email         string    `gorm:"type:varchar(255)" json:"email" validate:"omitempty,email"`
	Website       string    `gorm:"type:varchar(255)" json:"website"`
	KepalaSekolah string    `gorm:"type:varchar(255)" json:"kepala_sekolah"`
	Fitur         string    `gorm:"type:json" json:"fitur"`
	CreatedAt     time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt     time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"updated_at"`
	Kelas         []Kelas   `gorm:"foreignKey:SekolahID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"-"`
	Siswa         []Siswa   `gorm:"foreignKey:SekolahID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"-"`
}

type KonfigurasiPresensi struct {
	ID          int       `gorm:"primaryKey;autoIncrement" json:"id"`
	JamMasuk    string    `gorm:"type:varchar(5);not null;default:'07:00'" json:"jam_masuk"`
	JamPulang   string    `gorm:"type:varchar(5);not null;default:'16:00'" json:"jam_pulang"`
	BatasTelat  int       `gorm:"type:int;not null;default:15" json:"batas_telat"`
	Latitude    float64   `gorm:"type:decimal(10,7);not null" json:"latitude"`
	Longitude   float64   `gorm:"type:decimal(10,7);not null" json:"longitude"`
	RadiusMeter int       `gorm:"type:int;not null;default:50" json:"radius_meter"`
	CreatedAt   time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt   time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

type KonfigurasiNilaiTahfiz struct {
	ID                    int       `gorm:"primaryKey;autoIncrement" json:"id"`
	SekolahID             int       `gorm:"type:int(11);not null;index;uniqueIndex:idx_konfig_nilai_tahfiz" json:"sekolah_id"`
	JenisPenilaian        string    `gorm:"type:varchar(50);not null;default:'kelulusan';uniqueIndex:idx_konfig_nilai_tahfiz" json:"jenis_penilaian"`
	Pembagi               int       `gorm:"type:int;not null;default:1" json:"pembagi"`
	TingkatLevelKelulusan string    `gorm:"type:varchar(255);not null;default:'[]'" json:"-"`
	KriteriaPenilaian     string    `gorm:"type:text;not null" json:"-"`
	Aktif                 bool      `gorm:"type:tinyint(1);not null;default:1" json:"aktif"`
	CreatedAt             time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt             time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

type TahunAjaran struct {
	ID         int       `gorm:"primaryKey;autoIncrement" json:"id"`
	TahunAwal  int       `gorm:"type:int;not null;uniqueIndex" json:"tahun_awal"`
	TahunAkhir int       `gorm:"type:int;not null" json:"tahun_akhir"`
	NamaTahun  string    `gorm:"type:varchar(20);not null" json:"nama_tahun"`
	Status     string    `gorm:"type:varchar(50);not null;default:'Draft'" json:"status"`
	CreatedAt  time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt  time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"updated_at"`

	Semesters []Semester `gorm:"foreignKey:TahunAjaranID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"semesters"`
}

type Semester struct {
	ID            int       `gorm:"primaryKey;autoIncrement" json:"id"`
	TahunAjaranID int       `gorm:"not null;index;uniqueIndex:idx_tahun_tipe" json:"tahun_ajaran_id"`
	TipeSemester  string    `gorm:"type:varchar(50);not null;uniqueIndex:idx_tahun_tipe" json:"tipe_semester"`
	StatusAktif   bool      `gorm:"type:boolean;not null;default:false" json:"status_aktif"`
	CreatedAt     time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt     time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

type Kelas struct {
	ID           int       `gorm:"primaryKey;autoIncrement" json:"id"`
	SekolahID    int       `gorm:"not null;index" json:"sekolah_id"`
	Tingkat      string    `gorm:"type:varchar(10);not null" json:"tingkat" validate:"required,max=10"`
	TingkatLevel int       `gorm:"type:int;not null;default:1" json:"tingkat_level" validate:"required,gt=0"`
	NamaKelas    string    `gorm:"type:varchar(50);not null" json:"nama_kelas" validate:"required,max=50"`
	CreatedAt    time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt    time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

type Siswa struct {
	ID           int        `gorm:"primaryKey;autoIncrement" json:"id"`
	NISN         string     `gorm:"type:varchar(10);index" json:"nisn" validate:"omitempty,numeric,max=10"`
	NIS          string     `gorm:"type:varchar(20);uniqueIndex:idx_siswa_sekolah_nis;not null" json:"nis" validate:"required,max=20"`
	NamaLengkap  string     `gorm:"type:varchar(100);not null" json:"nama_lengkap" validate:"required"`
	Status       string     `gorm:"type:varchar(50);not null;default:'Aktif'" json:"status"`
	TempatLahir  string     `gorm:"type:varchar(50)" json:"tempat_lahir"`
	TanggalLahir *string    `gorm:"type:date" json:"tanggal_lahir"` // Using string for YYYY-MM-DD, pointer to handle null
	JenisKelamin string     `gorm:"type:varchar(50);not null" json:"jenis_kelamin" validate:"required,oneof=L P"`
	Agama        string     `gorm:"type:varchar(20)" json:"agama"`
	Alamat       string     `gorm:"type:text" json:"alamat"`
	NamaAyah     string     `gorm:"type:varchar(100)" json:"nama_ayah"`
	NamaIbu      string     `gorm:"type:varchar(100)" json:"nama_ibu"`
	TeleponOrtu  string     `gorm:"type:varchar(15)" json:"telepon_ortu"`
	CreatedAt    time.Time  `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt    time.Time  `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"updated_at"`
	DeletedAt    *time.Time `gorm:"index" json:"deleted_at,omitempty"`

	// Virtual Fields for Frontend
	SekolahID   int    `gorm:"uniqueIndex:idx_siswa_sekolah_nis" json:"sekolah_id"`
	KelasID     *int   `gorm:"->" json:"kelas_id"`
	Rombel      string `gorm:"->" json:"rombel"`
	KelasAsal   string `gorm:"->" json:"kelas_asal"`
	NamaKelas   string `gorm:"->" json:"nama_kelas"`
	TahunAjaran string `gorm:"->" json:"tahun_ajaran"`

	RiwayatKelas []RiwayatKelas `gorm:"foreignKey:SiswaID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"-"`
}

type RiwayatKelas struct {
	ID            int       `gorm:"primaryKey;autoIncrement" json:"id"`
	SiswaID       int       `gorm:"not null;uniqueIndex:idx_siswa_tahun" json:"siswa_id"`
	SekolahID     int       `gorm:"not null;index" json:"sekolah_id"`
	KelasID       *int      `gorm:"index" json:"kelas_id"` // Boleh NULL (Belum Rombel)
	Rombel        string    `gorm:"type:varchar(5)" json:"rombel"`
	KelasAsal     string    `gorm:"type:varchar(50)" json:"kelas_asal"` // Menyimpan nama kelas sebelumnya (misal: "X IPA 1")
	TahunAjaranID int       `gorm:"not null;uniqueIndex:idx_siswa_tahun" json:"tahun_ajaran_id"`
	StatusAkhir   string    `gorm:"type:varchar(50);not null;default:'Aktif'" json:"status_akhir"`
	CreatedAt     time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"created_at"`
}

type Nilai struct {
	ID             int       `gorm:"primaryKey;autoIncrement" json:"id"`
	SiswaID        int       `gorm:"not null;index;uniqueIndex:idx_nilai_scope" json:"siswa_id" validate:"required"`
	SekolahID      int       `gorm:"not null;index;uniqueIndex:idx_nilai_scope" json:"sekolah_id" validate:"required"`
	TahunAjaranID  int       `gorm:"not null;index;uniqueIndex:idx_nilai_scope" json:"tahun_ajaran_id"`
	SemesterID     int       `gorm:"not null;index;uniqueIndex:idx_nilai_scope" json:"semester_id"`
	JenisPenilaian string    `gorm:"type:varchar(50);not null;default:'semester';uniqueIndex:idx_nilai_scope" json:"jenis_penilaian"`
	MataPelajaran  string    `gorm:"type:varchar(100);not null;uniqueIndex:idx_nilai_scope" json:"mata_pelajaran" validate:"required"`
	Skor1          int       `gorm:"type:int;not null;default:0" json:"skor1" validate:"gte=0,lte=100"`
	Skor2          int       `gorm:"type:int;not null;default:0" json:"skor2" validate:"gte=0,lte=100"`
	Skor3          int       `gorm:"type:int;not null;default:0" json:"skor3" validate:"gte=0,lte=100"`
	Skor4          int       `gorm:"type:int;not null;default:0" json:"skor4" validate:"gte=0,lte=100"`
	Skor5          int       `gorm:"type:int;not null;default:0" json:"skor5" validate:"gte=0,lte=100"`
	NilaiAkhir     int       `gorm:"type:int;not null;default:0" json:"nilai_akhir"`
	Predikat       string    `gorm:"type:varchar(20)" json:"predikat"`
	Catatan        string    `gorm:"type:text" json:"catatan"`
	CreatedAt      time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt      time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"updated_at"`

	// Virtual fields for joins
	NamaSiswa string `gorm:"->" json:"nama_siswa,omitempty"`
	NISN      string `gorm:"->" json:"nisn,omitempty"`
}

type NilaiTahfiz struct {
	ID                    int       `gorm:"primaryKey;autoIncrement" json:"id"`
	SiswaID               int       `gorm:"not null;index;uniqueIndex:idx_nilai_tahfiz_scope" json:"siswa_id"`
	SekolahID             int       `gorm:"not null;index;uniqueIndex:idx_nilai_tahfiz_scope" json:"sekolah_id"`
	TahunAjaranID         int       `gorm:"not null;index;uniqueIndex:idx_nilai_tahfiz_scope" json:"tahun_ajaran_id"`
	SemesterID            int       `gorm:"not null;index;uniqueIndex:idx_nilai_tahfiz_scope" json:"semester_id"`
	JenisPenilaian        string    `gorm:"type:varchar(50);not null;default:'semester';uniqueIndex:idx_nilai_tahfiz_scope" json:"jenis_penilaian"`
	MataPelajaran         string    `gorm:"type:varchar(100);not null;default:'Tahfiz';uniqueIndex:idx_nilai_tahfiz_scope" json:"mata_pelajaran"`
	JumlahMateri          int       `gorm:"type:int;not null;default:0" json:"jumlah_materi"`
	TotalPoin             int       `gorm:"type:int;not null;default:0" json:"total_poin"`
	NilaiAkhir            float64   `gorm:"type:decimal(5,2);not null;default:0" json:"nilai_akhir"`
	Predikat              string    `gorm:"type:varchar(20)" json:"predikat"`
	Catatan               string    `gorm:"type:text" json:"catatan"`
	PembaginYangDigunakan int       `gorm:"type:int;not null;default:0" json:"pembagi_yang_digunakan"`
	CreatedAt             time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt             time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"updated_at"`

	Details []NilaiTahfizDetail `gorm:"foreignKey:NilaiTahfizID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"details,omitempty"`
}

type NilaiTahfizDetail struct {
	ID                  int       `gorm:"primaryKey;autoIncrement" json:"id"`
	NilaiTahfizID       int       `gorm:"not null;index;uniqueIndex:idx_nilai_tahfiz_detail" json:"nilai_tahfiz_id"`
	NamaMateri          string    `gorm:"type:varchar(150);not null;uniqueIndex:idx_nilai_tahfiz_detail" json:"nama_materi"`
	Juz                 int       `gorm:"type:int;not null;default:0" json:"juz"`
	Urutan              int       `gorm:"type:int;not null;default:0" json:"urutan"`
	SkorHafalan         *int      `gorm:"type:int;default:null" json:"skor_hafalan"`
	SkorKelancaran      *int      `gorm:"type:int;default:null" json:"skor_kelancaran"`
	SkorMad             *int      `gorm:"type:int;default:null" json:"skor_mad"`
	SkorTajwid          *int      `gorm:"type:int;default:null" json:"skor_tajwid"`
	SkorMakharijulHuruf *int      `gorm:"type:int;default:null" json:"skor_makharijul_huruf"`
	TotalPoin           int       `gorm:"type:int;not null;default:0" json:"total_poin"`
	StatusSetor         string    `gorm:"type:varchar(30);not null;default:'draft'" json:"status_setor"`
	CreatedAt           time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt           time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

type NilaiTahfizV2 struct {
	ID                    int       `gorm:"primaryKey;autoIncrement" json:"id"`
	SiswaID               int       `gorm:"not null;index;uniqueIndex:idx_nilai_tahfiz_v2_scope" json:"siswa_id"`
	SekolahID             int       `gorm:"not null;index;uniqueIndex:idx_nilai_tahfiz_v2_scope" json:"sekolah_id"`
	TahunAjaranID         int       `gorm:"not null;index;uniqueIndex:idx_nilai_tahfiz_v2_scope" json:"tahun_ajaran_id"`
	SemesterID            int       `gorm:"not null;index;uniqueIndex:idx_nilai_tahfiz_v2_scope" json:"semester_id"`
	JenisPenilaian        string    `gorm:"type:varchar(50);not null;default:'semester';uniqueIndex:idx_nilai_tahfiz_v2_scope" json:"jenis_penilaian"`
	MataPelajaran         string    `gorm:"type:varchar(100);not null;default:'Tahfiz';uniqueIndex:idx_nilai_tahfiz_v2_scope" json:"mata_pelajaran"`
	JumlahMateri          int       `gorm:"type:int;not null;default:0" json:"jumlah_materi"`
	TotalPoin             int       `gorm:"type:int;not null;default:0" json:"total_poin"`
	NilaiAkhir            float64   `gorm:"type:decimal(5,2);not null;default:0" json:"nilai_akhir"`
	Predikat              string    `gorm:"type:varchar(20)" json:"predikat"`
	Catatan               string    `gorm:"type:text" json:"catatan"`
	PembaginYangDigunakan int       `gorm:"type:int;not null;default:0" json:"pembagi_yang_digunakan"`
	CreatedAt             time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt             time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"updated_at"`

	Details []NilaiTahfizDetailV2 `gorm:"foreignKey:NilaiTahfizV2ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"details,omitempty"`
}

type NilaiTahfizDetailV2 struct {
	ID                  int       `gorm:"primaryKey;autoIncrement" json:"id"`
	NilaiTahfizV2ID     int       `gorm:"not null;index;uniqueIndex:idx_nilai_tahfiz_detail_v2" json:"nilai_tahfiz_v2_id"`
	NamaMateri          string    `gorm:"type:varchar(150);not null;uniqueIndex:idx_nilai_tahfiz_detail_v2" json:"nama_materi"`
	Juz                 int       `gorm:"type:int;not null;default:0" json:"juz"`
	Urutan              int       `gorm:"type:int;not null;default:0" json:"urutan"`
	SkorKelancaran      *int      `gorm:"type:int;default:null" json:"skor_kelancaran"`
	SkorTajwid          *int      `gorm:"type:int;default:null" json:"skor_tajwid"`
	SkorMakharijulHuruf *int      `gorm:"type:int;default:null" json:"skor_makharijul_huruf"`
	TotalPoin           int       `gorm:"type:int;not null;default:0" json:"total_poin"`
	StatusSetor         string    `gorm:"type:varchar(30);not null;default:'draft'" json:"status_setor"`
	CreatedAt           time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt           time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

type MateriTahfiz struct {
	ID         int       `gorm:"primaryKey;autoIncrement" json:"id"`
	SekolahID  int       `gorm:"not null;index;uniqueIndex:idx_materi_tahfiz_scope" json:"sekolah_id"`
	NamaMateri string    `gorm:"type:varchar(150);not null;uniqueIndex:idx_materi_tahfiz_scope" json:"nama_materi"`
	Juz        int       `gorm:"type:int;not null;default:0" json:"juz"`
	Urutan     int       `gorm:"type:int;not null;default:0" json:"urutan"`
	Aktif      bool      `gorm:"type:boolean;not null;default:true" json:"aktif"`
	CreatedAt  time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt  time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

type Guru struct {
	ID           int     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID       int   `gorm:"uniqueIndex;not null" json:"user_id"`
	NamaLengkap  string    `gorm:"type:varchar(100);not null" json:"nama_lengkap"`
	NIK          *string   `gorm:"type:varchar(50);uniqueIndex" json:"nik"`
	NoHP         *string   `gorm:"type:varchar(20)" json:"no_hp"`
	JamMasuk     string    `gorm:"type:varchar(5)" json:"jam_masuk"`
	JamPulang    string    `gorm:"type:varchar(5)" json:"jam_pulang"`
	AksesFitur   string    `gorm:"type:json" json:"akses_fitur"`
	AksesSekolah string    `gorm:"type:json" json:"akses_sekolah"`
	StatusAktif  bool      `gorm:"type:tinyint(1);not null;default:1" json:"status_aktif"`
	CreatedAt    time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt    time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"updated_at"`

	User User `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"user"`
}

type PresensiGuru struct {
	ID           int     `gorm:"primaryKey;autoIncrement" json:"id"`
	GuruID       int     `gorm:"not null;uniqueIndex:idx_presensi_guru_tanggal;index" json:"guru_id"`
	Tanggal      string    `gorm:"type:date;not null;uniqueIndex:idx_presensi_guru_tanggal" json:"tanggal"`
	WaktuMasuk   *string   `gorm:"type:varchar(10)" json:"waktu_masuk"`
	WaktuPulang  *string   `gorm:"type:varchar(10)" json:"waktu_pulang"`
	Status       string    `gorm:"type:varchar(20);not null;default:'Hadir'" json:"status"`
	LokasiMasuk  string    `gorm:"type:varchar(255)" json:"lokasi_masuk"`
	LokasiPulang string    `gorm:"type:varchar(255)" json:"lokasi_pulang"`
	CreatedAt    time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt    time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

type SetoranTahfizHarian struct {
	ID                  int       `gorm:"primaryKey;autoIncrement" json:"id"`
	SiswaID             int       `gorm:"not null;index" json:"siswa_id"`
	KelompokID          *int      `gorm:"index" json:"kelompok_id"`
	GuruID              *int      `gorm:"index" json:"guru_id"`
	Tanggal             string    `gorm:"type:date;not null;index" json:"tanggal"`        // format YYYY-MM-DD
	JenisSetoran        string    `gorm:"type:varchar(50);not null" json:"jenis_setoran"` // Ziyadah, Murojaah
	Juz                 int       `gorm:"type:int;not null;default:0" json:"juz"`
	Surah               string    `gorm:"type:varchar(150);not null" json:"surah"`
	SkorHafalan         int       `gorm:"type:int;not null;default:0" json:"skor_hafalan"`
	SkorKelancaran      int       `gorm:"type:int;not null;default:0" json:"skor_kelancaran"`
	SkorMad             int       `gorm:"type:int;not null;default:0" json:"skor_mad"`
	SkorTajwid          int       `gorm:"type:int;not null;default:0" json:"skor_tajwid"`
	SkorMakharijulHuruf int       `gorm:"type:int;not null;default:0" json:"skor_makharijul_huruf"`
	TotalPoin           int       `gorm:"type:int;not null;default:0" json:"total_poin"`
	Predikat            string    `gorm:"type:varchar(50);not null" json:"predikat"`
	Catatan             string    `gorm:"type:text" json:"catatan"`
	CreatedAt           time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt           time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

type Halaqah struct {
	ID            int       `gorm:"primaryKey;autoIncrement" json:"id"`
	SekolahID     int       `gorm:"not null;index" json:"sekolah_id"`
	GuruID        int       `gorm:"not null;index" json:"guru_id"`
	TahunAjaranID int       `gorm:"not null;index" json:"tahun_ajaran_id"`
	SemesterID    int       `gorm:"not null;index" json:"semester_id"`
	NamaHalaqah   string    `gorm:"type:varchar(150);not null" json:"nama_halaqah"`
	CreatedAt     time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt     time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"updated_at"`

	// Relasi
	Guru           Guru             `gorm:"foreignKey:GuruID" json:"guru,omitempty"`
	AnggotaHalaqah []AnggotaHalaqah `gorm:"foreignKey:HalaqahID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"anggota,omitempty"`
}

type AnggotaHalaqah struct {
	ID        int       `gorm:"primaryKey;autoIncrement" json:"id"`
	HalaqahID int       `gorm:"not null;index;uniqueIndex:idx_anggota_halaqah" json:"halaqah_id"`
	SiswaID   int       `gorm:"not null;index;uniqueIndex:idx_anggota_halaqah" json:"siswa_id"`
	CreatedAt time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"updated_at"`

	Siswa Siswa `gorm:"foreignKey:SiswaID" json:"siswa,omitempty"`
}

type MapelDiniyah struct {
	ID            int       `gorm:"primaryKey;autoIncrement" json:"id"`
	SekolahID     int       `gorm:"not null;index;uniqueIndex:idx_mapel_diniyah_scope" json:"sekolah_id"`
	NamaMapel     string    `gorm:"type:varchar(150);not null;uniqueIndex:idx_mapel_diniyah_scope" json:"nama_mapel"`
	NamaMapelArab string    `gorm:"type:varchar(255)" json:"nama_mapel_arab"`
	Kategori      string    `gorm:"type:varchar(50);not null;default:'tes_tulis'" json:"kategori"`
	Urutan        int       `gorm:"type:int;not null;default:0" json:"urutan"`
	Aktif         bool      `gorm:"type:boolean;not null;default:true" json:"aktif"`
	CreatedAt     time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt     time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

type NilaiDiniyah struct {
	ID             int       `gorm:"primaryKey;autoIncrement" json:"id"`
	SiswaID        int       `gorm:"not null;index;uniqueIndex:idx_nilai_diniyah_scope" json:"siswa_id"`
	SekolahID      int       `gorm:"not null;index;uniqueIndex:idx_nilai_diniyah_scope" json:"sekolah_id"`
	TahunAjaranID  int       `gorm:"not null;index;uniqueIndex:idx_nilai_diniyah_scope" json:"tahun_ajaran_id"`
	SemesterID     int       `gorm:"not null;index;uniqueIndex:idx_nilai_diniyah_scope" json:"semester_id"`
	MapelDiniyahID int       `gorm:"not null;index;uniqueIndex:idx_nilai_diniyah_scope" json:"mapel_diniyah_id"`
	UhScores       string    `gorm:"type:json;not null" json:"-"`
	RataUh         float64   `gorm:"type:decimal(5,2);not null;default:0" json:"rata_uh"`
	MurniPts       int       `gorm:"type:int;not null;default:0" json:"murni_pts"`
	RemidiPts      *int      `gorm:"type:int" json:"remidi_pts"`
	NaPts          int       `gorm:"type:int;not null;default:0" json:"na_pts"`
	Pas            int       `gorm:"type:int;not null;default:0" json:"pas"`
	NilaiRapor     float64   `gorm:"type:decimal(5,2);not null;default:0" json:"nilai_rapor"`
	CreatedAt      time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt      time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"updated_at"`

	// Virtual fields for joins
	NamaSiswa string `gorm:"->" json:"nama_siswa,omitempty"`
	NISN      string `gorm:"->" json:"nisn,omitempty"`
}

// Rapor Keseharian Models
type IndikatorKeseharian struct {
	ID             int       `gorm:"primaryKey;autoIncrement" json:"id"`
	SekolahID      int       `gorm:"not null;index" json:"sekolah_id"`
	Nama           string    `gorm:"type:varchar(150);not null" json:"nama"`
	NilaiPengurang float64   `gorm:"type:decimal(5,2);default:1.00;not null" json:"nilai_pengurang"`
	Aktif          bool      `gorm:"type:boolean;not null;default:true" json:"aktif"`
	CreatedAt      time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt      time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

type SesiKeseharian struct {
	ID            int       `gorm:"primaryKey;autoIncrement" json:"id"`
	SekolahID     int       `gorm:"not null;index" json:"sekolah_id"`
	TahunAjaranID int       `gorm:"not null;index" json:"tahun_ajaran_id"`
	SemesterID    int       `gorm:"not null;index" json:"semester_id"`
	IndikatorID   int       `gorm:"not null;index;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"indikator_id"`
	GuruID        int       `gorm:"not null;index" json:"guru_id"`
	Tanggal       time.Time `gorm:"type:date;not null" json:"tanggal"`
	Catatan       string    `gorm:"type:text" json:"catatan"`
	CreatedAt     time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt     time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

type SesiKeseharianDetail struct {
	ID               int       `gorm:"primaryKey;autoIncrement" json:"id"`
	SesiKeseharianID int       `gorm:"not null;index;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"sesi_keseharian_id"`
	SiswaID          int       `gorm:"not null;index" json:"siswa_id"`
	Status           bool      `gorm:"type:boolean;not null;default:false" json:"status"`
	CreatedAt        time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"created_at"`
}

type NilaiAkhirKeseharian struct {
	ID              int       `gorm:"primaryKey;autoIncrement" json:"id"`
	SiswaID         int       `gorm:"not null;index;uniqueIndex:idx_nilai_akhir_keseharian" json:"siswa_id"`
	SekolahID       int       `gorm:"not null;index;uniqueIndex:idx_nilai_akhir_keseharian" json:"sekolah_id"`
	TahunAjaranID   int       `gorm:"not null;index;uniqueIndex:idx_nilai_akhir_keseharian" json:"tahun_ajaran_id"`
	SemesterID      int       `gorm:"not null;index;uniqueIndex:idx_nilai_akhir_keseharian" json:"semester_id"`
	IndikatorID     int       `gorm:"not null;index;uniqueIndex:idx_nilai_akhir_keseharian" json:"indikator_id"`
	TotalSesi       int       `gorm:"type:int;not null;default:0" json:"total_sesi"`
	HadirSesi       int       `gorm:"type:int;not null;default:0" json:"hadir_sesi"`
	PoinMurni       float64   `gorm:"type:decimal(5,2);not null;default:0" json:"poin_murni"`
	PoinPenyesuaian float64   `gorm:"type:decimal(5,2);not null;default:0" json:"poin_penyesuaian"`
	PoinAkhir       float64   `gorm:"type:decimal(5,2);not null;default:0" json:"poin_akhir"`
	CreatedAt       time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt       time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

type CatatanNgajiSiswa struct {
	ID        int       `gorm:"primaryKey;autoIncrement" json:"id"`
	SiswaID   int       `gorm:"not null;uniqueIndex:idx_siswa_sekolah" json:"siswa_id"`
	SekolahID int       `gorm:"not null;uniqueIndex:idx_siswa_sekolah" json:"sekolah_id"`
	Catatan   string    `gorm:"type:text" json:"catatan"`
	CreatedAt time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

type JenisPelanggaran struct {
	ID        int       `gorm:"primaryKey;autoIncrement" json:"id"`
	SekolahID int       `gorm:"not null;index" json:"sekolah_id"`
	Nama      string    `gorm:"type:varchar(255);not null" json:"nama"`
	Poin      float64   `gorm:"type:decimal(5,2);not null;default:0" json:"poin"`
	Iqob      string    `gorm:"type:text" json:"iqob"`
	Aktif     bool      `gorm:"type:tinyint(1);not null;default:1" json:"aktif"`
	CreatedAt time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

type CatatanPelanggaran struct {
	ID                 int       `gorm:"primaryKey;autoIncrement" json:"id"`
	SiswaID            int       `gorm:"not null;index" json:"siswa_id"`
	SekolahID          int       `gorm:"not null;index" json:"sekolah_id"`
	TahunAjaranID      int       `gorm:"not null;index" json:"tahun_ajaran_id"`
	SemesterID         int       `gorm:"not null;index" json:"semester_id"`
	GuruID             int       `gorm:"not null;index" json:"guru_id"`
	JenisPelanggaranID int       `gorm:"not null;index" json:"jenis_pelanggaran_id"`
	Tanggal            time.Time `gorm:"type:date;not null" json:"tanggal"`
	Poin               float64   `gorm:"type:decimal(5,2);not null" json:"poin"`
	Iqob               string    `gorm:"type:text" json:"iqob"`
	Pemutihan          float64   `gorm:"type:decimal(5,2);not null;default:0" json:"pemutihan"`
	CreatedAt          time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt          time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"updated_at"`

	Siswa            Siswa            `gorm:"foreignKey:SiswaID" json:"siswa,omitempty"`
	JenisPelanggaran JenisPelanggaran `gorm:"foreignKey:JenisPelanggaranID" json:"jenis_pelanggaran,omitempty"`
}

type JenisIqob struct {
	ID        int       `gorm:"primaryKey;autoIncrement" json:"id"`
	SekolahID int       `gorm:"not null;index" json:"sekolah_id"`
	Nama      string    `gorm:"type:varchar(255);not null" json:"nama"`
	Poin      float64   `gorm:"type:decimal(5,2);not null;default:0" json:"poin"`
	Keterangan string   `gorm:"type:text" json:"keterangan"`
	Aktif     bool      `gorm:"type:tinyint(1);not null;default:1" json:"aktif"`
	CreatedAt time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

type CatatanIqob struct {
	ID                 int       `gorm:"primaryKey;autoIncrement" json:"id"`
	SiswaID            int       `gorm:"not null;index" json:"siswa_id"`
	SekolahID          int       `gorm:"not null;index" json:"sekolah_id"`
	TahunAjaranID      int       `gorm:"not null;index" json:"tahun_ajaran_id"`
	SemesterID         int       `gorm:"not null;index" json:"semester_id"`
	GuruID             int       `gorm:"not null;index" json:"guru_id"`
	JenisIqobID        int       `gorm:"not null;index" json:"jenis_iqob_id"`
	Tanggal            time.Time `gorm:"type:date;not null" json:"tanggal"`
	Poin               float64   `gorm:"type:decimal(5,2);not null" json:"poin"`
	Keterangan         string    `gorm:"type:text" json:"keterangan"`
	CreatedAt          time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt          time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"updated_at"`

	Siswa     Siswa     `gorm:"foreignKey:SiswaID" json:"siswa,omitempty"`
	JenisIqob JenisIqob `gorm:"foreignKey:JenisIqobID" json:"jenis_iqob,omitempty"`
}

type CatatanRapor struct {
	ID            int       `gorm:"primaryKey;autoIncrement" json:"id"`
	SiswaID       int       `gorm:"not null;index;uniqueIndex:idx_catatan_rapor" json:"siswa_id"`
	SekolahID     int       `gorm:"not null;index;uniqueIndex:idx_catatan_rapor" json:"sekolah_id"`
	TahunAjaranID int       `gorm:"not null;index;uniqueIndex:idx_catatan_rapor" json:"tahun_ajaran_id"`
	SemesterID    int       `gorm:"not null;index;uniqueIndex:idx_catatan_rapor" json:"semester_id"`
	JenisRapor    string    `gorm:"type:varchar(50);not null;uniqueIndex:idx_catatan_rapor" json:"jenis_rapor"`
	Catatan       string    `gorm:"type:text" json:"catatan"`
	CreatedAt     time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt     time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"updated_at"`

	Siswa Siswa `gorm:"foreignKey:SiswaID" json:"siswa,omitempty"`
}
