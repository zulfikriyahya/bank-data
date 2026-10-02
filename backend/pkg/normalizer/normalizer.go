package normalizer

import (
	"strconv"
	"strings"
	"time"
)

var bulanIndonesia = map[string]string{
	"januari": "01", "februari": "02", "maret": "03", "april": "04",
	"mei": "05", "juni": "06", "juli": "07", "agustus": "08",
	"september": "09", "oktober": "10", "november": "11", "desember": "12",
}

// NullableString - ubah "-", "", "N/A" jadi nil (NULL di DB)
func NullableString(val string) *string {
	trimmed := strings.TrimSpace(val)
	if trimmed == "" || trimmed == "-" || strings.EqualFold(trimmed, "n/a") {
		return nil
	}
	return &trimmed
}

// NullableInt - parse string ke *int, return nil kalau tidak valid/kosong
func NullableInt(val string) *int {
	s := NullableString(val)
	if s == nil {
		return nil
	}
	n, err := strconv.Atoi(strings.TrimSpace(*s))
	if err != nil {
		return nil
	}
	return &n
}

// ParseTanggalIndonesia - ubah "21 April 2014" jadi *time.Time
func ParseTanggalIndonesia(val string) *time.Time {
	s := NullableString(val)
	if s == nil {
		return nil
	}

	parts := strings.Fields(strings.TrimSpace(*s))
	if len(parts) != 3 {
		return nil
	}

	day := parts[0]
	if len(day) == 1 {
		day = "0" + day
	}

	monthName := strings.ToLower(parts[1])
	month, ok := bulanIndonesia[monthName]
	if !ok {
		return nil
	}

	year := parts[2]

	dateStr := year + "-" + month + "-" + day
	parsed, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return nil
	}

	return &parsed
}

// NormalizeJenisKelamin - pastikan hanya "Laki-laki" atau "Perempuan" yang masuk DB
func NormalizeJenisKelamin(val string) *string {
	s := NullableString(val)
	if s == nil {
		return nil
	}
	v := strings.TrimSpace(*s)
	if v == "Laki-laki" || v == "Perempuan" {
		return &v
	}
	return nil
}
