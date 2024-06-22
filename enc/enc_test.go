package enc

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"

	"github.com/twpayne/go-geom"
	"github.com/twpayne/go-geom/encoding/geojson"
)

func testDecode(fn string) (geojson.FeatureCollection, error) {
	fc := geojson.FeatureCollection{}
	r, err := os.Open(fn)
	if err != nil {
		return fc, err
	}
	fc, err = PolylineToGeojson(r)
	if err != nil {
		return fc, err
	}
	return fc, nil
}

func TestCreateFromGeojson(t *testing.T) {
	outf, _ := os.CreateTemp("", "")
	outf.Close()
	outfn := outf.Name()
	defer os.Remove(outfn)
	err := CreateFromGeojson("../testdata/test.geojson", outfn, "tzid")
	if err != nil {
		t.Fatal(err)
	}
	//
	fc, err := testDecode(outfn)
	if err != nil {
		t.Fatal(err)
	}
	checkTestGeojson(t, fc)
}

func TestCreateFromZipGeojson(t *testing.T) {
	outf, _ := os.CreateTemp("", "")
	outf.Close()
	outfn := outf.Name()
	defer os.Remove(outfn)
	err := CreateFromZipGeojson("../testdata/test.geojson.zip", outfn, "tzid")
	if err != nil {
		t.Fatal(err)
	}
	//
	fc, err := testDecode(outfn)
	if err != nil {
		t.Fatal(err)
	}
	checkTestGeojson(t, fc)
}

func TestDecode(t *testing.T) {
	outfn, err := testCreate()
	if err != nil {
		t.Fatal(err)
	}
	fc, err := testDecode(outfn)
	if err != nil {
		t.Fatal(err)
	}
	checkTestGeojson(t, fc)
}

func BenchmarkDecodeTimezones(b *testing.B) {
	outf, _ := os.CreateTemp("", "")
	outf.Close()
	outfn := outf.Name()
	defer os.Remove(outfn)
	err := CreateFromZipGeojson("../timezones/timezones-now.geojson.zip", outfn, "tzid")
	if err != nil {
		b.Fatal(err)
	}
	r, err := os.Open(outfn)
	if err != nil {
		b.Skip(err)
	}
	data, err := ioutil.ReadAll(r)
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		rdata := bytes.NewReader(data)
		fc, err := PolylineToGeojson(rdata)
		if err != nil {
			b.Fatal(err)
		}
		_ = fc
	}
}

func testCreate() (string, error) {
	outf, _ := os.CreateTemp("", "")
	outf.Close()
	outfn := outf.Name()
	err := CreateFromGeojson("../testdata/test.geojson", outfn, "tzid")
	if err != nil {
		return "", err
	}
	return outfn, err
}

func checkTestGeojson(t *testing.T, fc geojson.FeatureCollection) {
	featCount := map[string]int{"sf": 2, "eb": 1}
	ringCount := map[string]int{"sf": 3, "eb": 1}
	gotRingCount := map[string]int{}
	gotFeatCount := map[string]int{}
	for _, feature := range fc.Features {
		g, ok := feature.Geometry.(*geom.Polygon)
		if !ok {
			t.Errorf("got %T, expected *geom.Polygon", feature.Geometry)
		}
		gotFeatCount[feature.ID] += 1
		gotRingCount[feature.ID] += g.NumLinearRings()
	}
	for k, v := range ringCount {
		if a := gotRingCount[k]; a != v {
			t.Errorf("got %d total rings for features with name %s, expected %d", a, k, v)
		}
	}
	for k, v := range featCount {
		if a := gotFeatCount[k]; a != v {
			t.Errorf("got %d features with name %s, expected %d", a, k, v)
		}
	}
}
