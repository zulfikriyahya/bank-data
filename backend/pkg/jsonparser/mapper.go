package jsonparser

import (
	"encoding/json"

	"bank-data/backend/internal/domain"
	"bank-data/backend/pkg/normalizer"
)

// MapToSiswa — ubah map JSON mentah (field "siswa") jadi domain.Siswa
func MapToSiswa(raw map[string]string, source string, page, row int, isFailed bool, rawPayload interface{}) *domain.Siswa {
	rawJSONBytes, _ := json.Marshal(rawPayload)
	rawJSONStr := string(rawJSONBytes)

	return &domain.Siswa{
		NIK:                  normalizer.NullableString(raw["NIK"]),
		NISN:                 normalizer.NullableString(raw["NISN"]),
		KIP:                  normalizer.NullableString(raw["KIP"]),
		Nama:                 raw["NAMA"], // wajib ada, tidak boleh NULL
		TempatLahir:          normalizer.NullableString(raw["TEMPAT LAHIR"]),
		TanggalLahir:         normalizer.ParseTanggalIndonesia(raw["TANGGAL LAHIR"]),
		JenisKelamin:         normalizer.NormalizeJenisKelamin(raw["JENIS KELAMIN"]),
		Agama:                normalizer.NullableString(raw["AGAMA"]),
		JumlahSaudara:        normalizer.NullableInt(raw["JUMLAH SAUDARA"]),
		AnakKe:               normalizer.NullableInt(raw["ANAK KE"]),
		Hobi:                 normalizer.NullableString(raw["HOBI"]),
		CitaCita:             normalizer.NullableString(raw["CITA-CITA"]),
		NoHandphone:          normalizer.NullableString(raw["NO. HANDPHONE"]),
		Email:                normalizer.NullableString(raw["ALAMAT EMAIL SISWA"]),
		YangMembiayai:        normalizer.NullableString(raw["YANG MEMBIAYAI SEKOLAH"]),
		KebutuhanDisabilitas: normalizer.NullableString(raw["KEBUTUHAN DISABILITAS"]),
		KebutuhanKhusus:      normalizer.NullableString(raw["KEBUTUHAN KHUSUS"]),
		Alamat:               normalizer.NullableString(raw["ALAMAT"]),
		StatusTempatTinggal:  normalizer.NullableString(raw["STATUS TEMPAT TINGGAL"]),
		JarakRumahMadrasah:   normalizer.NullableString(raw["JARAK TEMPAT TINGGAL - MADRASAH"]),
		WaktuTempuh:          normalizer.NullableString(raw["WAKTU TEMPUH"]),
		Transportasi:         normalizer.NullableString(raw["TRANSPORTASI KE SEKOLAH"]),
		Fingerprint:          normalizer.NullableString(raw["FINGERPRINT"]),
		RFID:                 normalizer.NullableString(raw["RFID"]),
		Source:               source,
		SourcePage:           &page,
		SourceRow:            &row,
		IsFailed:             isFailed,
		RawJSON:              &rawJSONStr,
	}
}

// MapToWali — dipakai untuk ayah/ibu/wali (struktur field sama, beda role)
func MapToWali(raw map[string]string, role string) domain.WaliSiswa {
	return domain.WaliSiswa{
		Role:                role,
		NIK:                 normalizer.NullableString(raw["NIK"]),
		NamaLengkap:         normalizer.NullableString(raw["NAMA LENGKAP"]),
		TempatLahir:         normalizer.NullableString(raw["TEMPAT LAHIR"]),
		TanggalLahir:        normalizer.ParseTanggalIndonesia(raw["TANGGAL LAHIR"]),
		Status:              normalizer.NullableString(raw["STATUS"]),
		PendidikanTerakhir:  normalizer.NullableString(raw["PENDIDIKAN TERAKHIR"]),
		PekerjaanUtama:      normalizer.NullableString(raw["PEKERJAAN UTAMA"]),
		Domisili:            normalizer.NullableString(raw["DOMISILI"]),
		NoHandphone:         normalizer.NullableString(raw["NO HANDPHONE"]),
		Penghasilan:         normalizer.NullableString(raw["PENGHASILAN RATA-RATA PER BULAN (Rp)"]),
		Alamat:              normalizer.NullableString(raw["ALAMAT"]),
		StatusTempatTinggal: normalizer.NullableString(raw["STATUS TEMPAT TINGGAL"]),
	}
}

// MapToAktivitasBelajar — list karena satu siswa bisa punya banyak riwayat
func MapToAktivitasBelajar(rawList []map[string]string) []domain.AktivitasBelajar {
	result := make([]domain.AktivitasBelajar, 0, len(rawList))

	for _, raw := range rawList {
		result = append(result, domain.AktivitasBelajar{
			TahunAjaranSemester: normalizer.NullableString(raw["TAHUN AJARAN - SEMESTER"]),
			TanggalMulaiMasuk:   normalizer.ParseTanggalIndonesia(raw["TANGGAL MULAI MASUK"]),
			NSMNamaLembaga:      normalizer.NullableString(raw["NSM - NAMA LEMBAGA"]),
			TingkatKelompok:     normalizer.NullableString(raw["TINGKAT / KELOMPOK"]),
			Jurusan:             normalizer.NullableString(raw["JURUSAN"]),
			StatusKeaktifan:     normalizer.NullableString(raw["STATUS KEAKTIFAN"]),
			Keterangan:          normalizer.NullableString(raw["KETERANGAN"]),
		})
	}

	return result
}

func MapToBeasiswa(rawList []map[string]string) []domain.Beasiswa {
	result := make([]domain.Beasiswa, 0, len(rawList))
	for _, raw := range rawList {
		result = append(result, domain.Beasiswa{
			NamaBeasiswa: raw["NAMA BEASISWA"],
			Tahun:        raw["TAHUN"],
			Keterangan:   raw["KETERANGAN"],
		})
	}
	return result
}

func MapToPrestasi(rawList []map[string]string) []domain.Prestasi {
	result := make([]domain.Prestasi, 0, len(rawList))
	for _, raw := range rawList {
		result = append(result, domain.Prestasi{
			NamaPrestasi: raw["NAMA PRESTASI"],
			Tingkat:      raw["TINGKAT"],
			Tahun:        raw["TAHUN"],
			Keterangan:   raw["KETERANGAN"],
		})
	}
	return result
}
