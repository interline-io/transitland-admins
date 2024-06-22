package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/interline-io/transitland-admins/enc"
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
		err = enc.CreateFromGeojson(infn, outfn, idKey, ek...)
	} else if fntype == "zipgeojson" {
		err = enc.CreateFromZipGeojson(infn, outfn, idKey, ek...)
	} else if fntype == "shapefile" {
		err = enc.CreateFromShapefile(infn, outfn, idKey, ek...)
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
