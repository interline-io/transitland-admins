package main

import (
	"archive/zip"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/twpayne/go-geom"
	"github.com/twpayne/go-geom/encoding/geojson"
	"github.com/twpayne/go-polyline"
	"github.com/twpayne/go-shapefile"
)

func main() {
	idKey := ""
	extraKeys := ""
	flag.StringVar(&idKey, "idkey", "", "")
	flag.StringVar(&extraKeys, "include", "", "Include this set of properties in output; comma separated")
	flag.Parse()
	fntype := flag.Arg(0)
	infn := flag.Arg(1)
	outfn := flag.Arg(2)
	ek := strings.Split(extraKeys, ",")
	if err := run(fntype, infn, outfn, idKey, ek); err != nil {
		fmt.Println("failed:", err)
		os.Exit(1)
	}
}

func run(fntype, infn, outfn, idKey string, ek []string) error {
	if strings.HasPrefix(infn, "http") {
		tmpf, err := os.CreateTemp("", "")
		if err != nil {
			return err
		}
		tmpf.Close()
		tname := tmpf.Name()
		defer os.Remove(tname)
		if err := downloadFile(infn, tname); err != nil {
			return err
		}
		infn = tname
	}

	fmt.Printf("Reading %s from %s, output: %s\n", fntype, infn, outfn)
	var err error
	if fntype == "geojson" {
		err = CreateFromGeojson(infn, outfn, idKey, ek...)
	} else if fntype == "zipgeojson" {
		err = CreateFromZipGeojson(infn, outfn, idKey, ek...)
	} else if fntype == "shapefile" {
		err = CreateFromShapefile(infn, outfn, idKey, ek...)
	} else {
		return fmt.Errorf("unknown format: %s", fntype)
	}
	return err
}

func downloadFile(url string, filepath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

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
	if err := GeojsonToPolylines(fc, w, idKey, keys); err != nil {
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
	return GeojsonToPolylines(fc, w, idKey, keys)
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
		if err := GeojsonToPolylines(fc, w, idKey, keys); err != nil {
			return err
		}
	}
	return nil
}

// Copied frmo tlxy
const polylineScale = 1e6

func GeojsonToPolylines(fc geojson.FeatureCollection, w io.Writer, idKey string, keys []string) error {
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
