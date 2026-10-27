package main

import (
	_ "embed"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"text/template"
)

//go:embed res/index.html
var indexHTML []byte

func getExternalIP() (string, error) {
	intfs, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	for _, intf := range intfs {
		addrs, err := intf.Addrs()
		if err != nil {
			return "", err
		}
		for _, addr := range addrs {
			ipnet, ok := addr.(*net.IPNet)
			if !ok {
				continue
			}
			ip4 := ipnet.IP.To4()
			if ip4 == nil {
				continue // не IPv4
			}
			if ip4.IsLoopback() || ip4.IsLinkLocalUnicast() {
				continue
			}
			if ip4[0] == 192 && ip4[1] == 168 {
				return ip4.String(), nil
			}
		}
	}
	return "", errors.New("no 192.168.x.x address found")
}

type pageData struct {
	Text  string
	Files []string
}

var tmpl = template.Must(template.New("index").Parse(string(indexHTML)))

func runServer(text string, filenames []string, port int) error {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		data := pageData{
			Text:  text,
			Files: filenames,
		}
		err := tmpl.Execute(w, data)
		if err != nil {
			panic(err)
		}
	})

	http.HandleFunc("/submit", func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseMultipartForm(32 << 20)
		if err != nil {
			panic(err)
		}
		inputText := r.MultipartForm.Value["data"]
		if len(inputText) > 0 && inputText[0] != "" {
			fmt.Printf("Text received:\n%s", inputText[0])
		}
		for _, fh := range r.MultipartForm.File["files"] {
			src, _ := fh.Open()
			err := os.MkdirAll("out", 0o700)
			if err != nil {
				panic(err)
			}
			path := filepath.Join("out", filepath.Base(fh.Filename))
			dst, err := os.Create(path)
			if err != nil {
				panic(err)
			}
			_, _ = io.Copy(dst, src)
			_ = src.Close()
			_ = dst.Close()
			fmt.Printf("saved file as %s", path)
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
	})
	localAddr := net.JoinHostPort("0.0.0.0", strconv.Itoa(port))
	localURL := fmt.Sprintf("http://%s", localAddr)

	externalIp, err := getExternalIP()
	if err != nil {
		return err
	}
	externalAddr := net.JoinHostPort(externalIp, strconv.Itoa(port))
	externalURL := fmt.Sprintf("http://%s", externalAddr)

	fmt.Println("Serving at ", string(localURL), "\t", string(externalURL))
	err = http.ListenAndServe(localAddr, nil)
	fmt.Println(err.Error())
	return nil
}

func main() {
	text := flag.String("text", "", "text to send to clients. Panics when used with --stdin")
	stdinMode := flag.Bool("stdin", false, "text to send to clients. Panics when used with --text")
	port := flag.Int("port", 8000, "port to listen on")
	flag.Parse()
	filenames := flag.Args()
	if *text != "" && *stdinMode {
		panic("cannot use --text and --stdin simultaneously")
	}
	if *stdinMode {
		stdinText, err := io.ReadAll(os.Stdin)
		if err != nil {
			panic(err)
		}
		*text = string(stdinText)
	}

	fmt.Println(filenames)
	err := runServer(*text, filenames, *port)
	if err != nil {
		panic(err)
	}
}
