package timezones

import (
	"bytes"
	_ "embed"

	"github.com/interline-io/transitland-admins/enc"
	"github.com/twpayne/go-geom/encoding/geojson"
)

//go:embed timezones.polyline
var timezonesData []byte

func Load() (geojson.FeatureCollection, error) {
	return enc.PolylineToGeojson(bytes.NewBuffer(timezonesData))
}
