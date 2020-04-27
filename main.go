package main

import (
	"flag"
	"io"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/net/proxy"
)

func init() {
	flag.Parse()
}

func main() {
	args := flag.Args()

	addr := args[0]

	dialer := proxy.FromEnvironment()
	ctxDialer := dialer.(proxy.ContextDialer)
	client := &http.Client{
		Transport: &http.Transport{
			DialContext: ctxDialer.DialContext,
			Dial:        dialer.Dial,
		},
		Timeout: time.Second * 32,
	}

	resp, err := client.Get(addr)
	ce(err, "get %s", addr)
	defer resp.Body.Close()

	var filename string
	dispositionHeader := resp.Header.Get("Content-Disposition")
	disposition, params, err := mime.ParseMediaType(dispositionHeader)
	ce(err, "parse Content-Disposition header")
	if disposition == "attachment" {
		if name, ok := params["filename"]; ok {
			filename = name
		}
	}
	if filename == "" {
		u, err := url.Parse(addr)
		ce(err, "parse %s", addr)
		filename = filepath.Base(u.Path)
	}

	if _, err := os.Stat(filename); err == nil {
		panic(me(nil, "file exists: %s", filename))
	}

	f, err := os.Create(filename)
	ce(err, "create %s", filename)
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	ce(err)

}