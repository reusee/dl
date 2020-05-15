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
			DialContext:           ctxDialer.DialContext,
			Dial:                  dialer.Dial,
			ResponseHeaderTimeout: time.Second * 32,
		},
	}

	retry := 10
do:

	resp, err := client.Get(addr)
	if err != nil {
		if retry > 0 {
			retry--
			goto do
		}
		ce(err, "get %s", addr)
	}
	defer resp.Body.Close()

	var filename string
	dispositionHeader := resp.Header.Get("Content-Disposition")
	disposition, params, err := mime.ParseMediaType(dispositionHeader)
	if err == nil {
		if disposition == "attachment" {
			if name, ok := params["filename"]; ok {
				filename = name
			}
		}
	}
	if filename == "" {
		u, err := url.Parse(addr)
		ce(err, "parse %s", addr)
		filename = filepath.Base(u.Path)
	}

	//if _, err := os.Stat(filename); err == nil {
	//	panic(me(nil, "file exists: %s", filename))
	//}

	tmpFilename := filename + ".tmp"
	f, err := os.Create(tmpFilename)
	ce(err, "create %s", tmpFilename)

	buf := make([]byte, 8192)
	c := 0
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			_, err := f.Write(buf[:n])
			ce(err)
		}
		if err == io.EOF {
			break
		} else if err != nil {
			ce(err)
		}
		c += n
		pt("%s %d\n", addr, c)
	}

	ce(f.Close())
	ce(os.Rename(tmpFilename, filename))

}
