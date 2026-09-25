package tag

import (
	"fmt"
	"os"
	"time"

	"github.com/google/renameio/v2"

	"gpspicturetagd/internal/exifgps"
	"gpspicturetagd/internal/seen"
)

func File(path string, pos exifgps.Position, maxAge time.Duration, cache *seen.Cache) (tagged bool, reason string, err error) {
	reason = Check(path, maxAge, cache)
	if reason != "" {
		return false, reason, nil
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return false, "", err
	}
	out, err := exifgps.SetGPS(raw, pos)
	if err != nil {
		return false, "", err
	}
	st, err := os.Stat(path)
	if err != nil {
		return false, "", err
	}
	pf, err := renameio.NewPendingFile(path, renameio.WithExistingPermissions())
	if err != nil {
		return false, "", err
	}
	defer pf.Cleanup()
	if _, err := pf.Write(out); err != nil {
		return false, "", err
	}
	if err := os.Chtimes(pf.Name(), st.ModTime(), st.ModTime()); err != nil {
		return false, "", err
	}
	if err := pf.CloseAtomicallyReplace(); err != nil {
		return false, "", fmt.Errorf("replace: %w", err)
	}
	if cache != nil {
		if st2, err := os.Stat(path); err == nil {
			cache.Remember(path, st2.ModTime().UnixNano(), st2.Size())
		}
	}
	return true, "", nil
}

func FromFix(lat, lon, alt float64, hasAlt bool, t time.Time) exifgps.Position {
	return exifgps.Position{Lat: lat, Lon: lon, Alt: alt, HasAlt: hasAlt, Time: t}
}

func HasGPS(path string) bool {
	raw, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	return exifgps.HasGPS(raw)
}

// Check returns a skip reason, or "" if the file should be tagged.
func Check(path string, maxAge time.Duration, cache *seen.Cache) string {
	st, err := os.Stat(path)
	if err != nil {
		return err.Error()
	}
	if maxAge > 0 && time.Since(st.ModTime()) > maxAge {
		return "too old"
	}
	if cache.Tagged(path, st.ModTime().UnixNano(), st.Size()) {
		return "already tagged"
	}
	if HasGPS(path) {
		cache.Remember(path, st.ModTime().UnixNano(), st.Size())
		return "already tagged"
	}
	return ""
}
