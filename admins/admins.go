package admins

import (
	"bytes"
	_ "embed"

	"github.com/interline-io/transitland-admins/enc"
	"github.com/twpayne/go-geom/encoding/geojson"
)

//go:embed admins.polyline
var adminData []byte

func Load() (geojson.FeatureCollection, error) {
	return enc.PolylineToGeojson(bytes.NewBuffer(adminData))
}
