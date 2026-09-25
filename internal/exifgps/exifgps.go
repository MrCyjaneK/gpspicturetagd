package exifgps

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"slices"
	"sort"
	"time"

	gometadata "github.com/FlavioCFOliveira/GoMetadata"
	"github.com/FlavioCFOliveira/GoMetadata/exif"
)

type Position struct {
	Lat, Lon, Alt float64
	HasAlt        bool
	Time          time.Time
}

// HasGPS reports whether the JPEG already has GPS latitude and longitude.
func HasGPS(jpeg []byte) bool {
	_, ok := ReadGPS(jpeg)
	return ok
}

// SetGPS writes WGS-84 GPS tags into the JPEG without recompressing.
func SetGPS(jpeg []byte, pos Position) ([]byte, error) {
	r := bytes.NewReader(jpeg)
	m, err := gometadata.Read(r)
	if err != nil {
		return nil, err
	}
	if m.EXIF == nil {
		m.EXIF = &exif.EXIF{ByteOrder: binary.LittleEndian, IFD0: &exif.IFD{}}
	}
	m.EXIF.SetGPS(pos.Lat, pos.Lon)
	if m.EXIF.GPSIFD == nil {
		return nil, fmt.Errorf("invalid GPS coordinates")
	}
	decorateGPS(m.EXIF, pos)

	if _, err := r.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := gometadata.Write(r, &buf, m); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// ReadGPS returns lat/lon (and altitude when present) from EXIF.
func ReadGPS(jpeg []byte) (Position, bool) {
	m, err := gometadata.Read(bytes.NewReader(jpeg),
		gometadata.WithoutMakerNote(),
		gometadata.WithoutIPTC(),
		gometadata.WithoutXMP(),
	)
	if err != nil {
		return Position{}, false
	}
	lat, lon, ok := m.GPS()
	if !ok {
		return Position{}, false
	}
	pos := Position{Lat: lat, Lon: lon}
	if alt, ok := m.Altitude(); ok {
		pos.HasAlt = true
		pos.Alt = alt
	}
	return pos, true
}

func decorateGPS(e *exif.EXIF, pos Position) {
	gps := e.GPSIFD
	if gps == nil || len(gps.Entries) == 0 {
		return
	}
	bo := e.ByteOrder
	if bo == nil {
		bo = binary.LittleEndian
	}
	putGPSTag(gps, exif.TagGPSMapDatum, exif.TypeASCII, asciiBytes("WGS-84"))
	if pos.HasAlt && !math.IsNaN(pos.Alt) {
		ref := byte(0)
		alt := pos.Alt
		if alt < 0 {
			ref = 1
			alt = -alt
		}
		putGPSTag(gps, exif.TagGPSAltitudeRef, exif.TypeByte, []byte{ref})
		putGPSTag(gps, exif.TagGPSAltitude, exif.TypeRational, rational(bo, alt))
	}
	t := pos.Time.UTC()
	if t.IsZero() {
		t = time.Now().UTC()
	}
	putGPSTag(gps, exif.TagGPSTimeStamp, exif.TypeRational, rationals(bo,
		float64(t.Hour()),
		float64(t.Minute()),
		float64(t.Second())+float64(t.Nanosecond())/1e9,
	))
	putGPSTag(gps, exif.TagGPSDateStamp, exif.TypeASCII, asciiBytes(t.Format("2006:01:02")))
}

func putGPSTag(ifd *exif.IFD, tag exif.TagID, typ exif.DataType, value []byte) {
	count := uint32(len(value))
	if typ == exif.TypeRational {
		count = uint32(len(value) / 8)
	}
	// Clone an existing GPS entry so the unexported byte-order flag matches the file.
	e := ifd.Entries[0]
	e.Tag = tag
	e.Type = typ
	e.Count = count
	e.Value = value
	i := sort.Search(len(ifd.Entries), func(i int) bool { return ifd.Entries[i].Tag >= tag })
	if i < len(ifd.Entries) && ifd.Entries[i].Tag == tag {
		ifd.Entries[i] = e
		return
	}
	ifd.Entries = slices.Insert(ifd.Entries, i, e)
}

func asciiBytes(s string) []byte {
	return append([]byte(s), 0)
}

func rational(bo binary.ByteOrder, v float64) []byte {
	const den = 10000
	n := uint32(math.Round(v * den))
	b := make([]byte, 8)
	bo.PutUint32(b[0:4], n)
	bo.PutUint32(b[4:8], den)
	return b
}

func rationals(bo binary.ByteOrder, vs ...float64) []byte {
	out := make([]byte, 0, 8*len(vs))
	for _, v := range vs {
		out = append(out, rational(bo, v)...)
	}
	return out
}
