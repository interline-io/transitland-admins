package enc

import (
	"archive/zip"
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/twpayne/go-geom"
	"github.com/twpayne/go-geom/encoding/geojson"
	"github.com/twpayne/go-polyline"
	"github.com/twpayne/go-shapefile"
)

// Space efficient encoding using polylines

const polylineScale = 1e6

func CreateFromGeojson(fn string, outfn string, idKey string, keys ...string) error {
	w, _ := os.Create(outfn)
	defer w.Close()
	r, err := os.Open(fn)
	if err != nil {
		return err
	}
	fcData, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	fc := geojson.FeatureCollection{}
	if err := fc.UnmarshalJSON(fcData); err != nil {
		return err
	}
	if err := geojsonToPolyline(fc, w, idKey, keys); err != nil {
		return err
	}
	return nil
}

func CreateFromShapefile(fn string, outfn string, idKey string, keys ...string) error {
	r, err := shapefile.ReadZipFile(fn, nil)
	if err != nil {
		return err
	}
	var features []*geojson.Feature
	for i := 0; i < r.NumRecords(); i++ {
		rec, recGeom := r.Record(i)
		features = append(features, &geojson.Feature{
			Properties: rec,
			Geometry:   recGeom,
		})
	}

	w, _ := os.Create(outfn)
	defer w.Close()
	fc := geojson.FeatureCollection{
		Features: features,
	}
	return geojsonToPolyline(fc, w, idKey, keys)
}

func CreateFromZipGeojson(fn string, outfn string, idKey string, keys ...string) error {
	w, _ := os.Create(outfn)
	defer w.Close()
	zf, err := zip.OpenReader(fn)
	if err != nil {
		return err
	}
	for _, f := range zf.File {
		if !(strings.HasSuffix(f.Name, ".json") || strings.HasSuffix(f.Name, ".geojson")) {
			continue
		}
		r, err := f.Open()
		if err != nil {
			return err
		}
		fcData, err := ioutil.ReadAll(r)
		if err != nil {
			return err
		}
		fc := geojson.FeatureCollection{}
		if err := fc.UnmarshalJSON(fcData); err != nil {
			return err
		}
		if err := geojsonToPolyline(fc, w, idKey, keys); err != nil {
			return err
		}
	}
	return nil
}

func geojsonToPolyline(fc geojson.FeatureCollection, w io.Writer, idKey string, keys []string) error {
	codec := polyline.Codec{Dim: 2, Scale: polylineScale}
	for i, feature := range fc.Features {
		if i == 0 {
			var recKeys []string
			for k := range feature.Properties {
				recKeys = append(recKeys, k)
			}
			fmt.Printf("first record has keys: %v\n", recKeys)
			fmt.Printf("selecting keys: %v\n", keys)
			fmt.Printf("first record has geom: %T\n", feature.Geometry)
		}
		fmt.Printf("processing record: %d\n", i)
		// jj, _ := json.Marshal(feature.Properties)
		// fmt.Println(string(jj))

		var polys []*geom.Polygon
		if v, ok := feature.Geometry.(*geom.Polygon); ok {
			polys = append(polys, v)
		} else if v, ok := feature.Geometry.(*geom.MultiPolygon); ok {
			for i := 0; i < v.NumPolygons(); i++ {
				polys = append(polys, v.Polygon(i))
			}
		}
		for _, g := range polys {
			tzName := feature.ID
			if a, ok := feature.Properties[idKey].(string); idKey != "" && ok {
				tzName = a
			}
			var jj []byte
			if len(keys) > 0 {
				props := map[string]any{}
				for _, key := range keys {
					props[key] = feature.Properties[key]
				}
				jj, _ = json.Marshal(props)
			}
			row := []string{tzName, string(jj)}
			// Encode coordinates
			for p := 0; p < g.NumLinearRings(); p++ {
				pring := g.LinearRing(p)
				var pc [][]float64
				for _, p2 := range pring.Coords() {
					pc = append(pc, []float64{p2[0], p2[1]})
				}
				var buf []byte
				buf = codec.EncodeCoords(buf, pc)
				row = append(row, string(buf))
			}
			w.Write([]byte(strings.Join(row, "\t") + "\n"))
		}
	}
	return nil
}

func polylineToGeojson(r io.Reader) (geojson.FeatureCollection, error) {
	codec := polyline.Codec{Dim: 2, Scale: polylineScale}
	data, _ := ioutil.ReadAll(r)
	var features []*geojson.Feature
	scanner := bufio.NewScanner(bytes.NewReader(data))
	scanner.Buffer(nil, 1024*1024)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		sp := bytes.Split(scanner.Bytes(), []byte("\t"))
		if len(sp) < 2 {
			continue
		}
		g := geom.NewPolygon(geom.XY)
		tzName := string(sp[0])
		var props map[string]any

		if spi := sp[1]; len(spi) > 0 {
			if err := json.Unmarshal(spi, &props); err != nil {
				panic(err)
			}
		}
		for i := 2; i < len(sp); i++ {
			spi := sp[i]
			if len(spi) == 0 {
				continue
			}
			var dec []float64
			dec, _, err := codec.DecodeFlatCoords(dec, spi)
			if err != nil {
				panic(err)
			}
			g.Push(geom.NewLinearRingFlat(geom.XY, dec))
		}
		features = append(features, &geojson.Feature{
			ID:         tzName,
			Properties: props,
			Geometry:   g,
		})
	}
	return geojson.FeatureCollection{Features: features}, nil
}
