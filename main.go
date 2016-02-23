package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/moovweb/gokogiri"

	// "gopkg.in/xmlpath.v2"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/proxy", proxyHandler)
	mux.HandleFunc("/prompt", promptHandler)
	mux.HandleFunc("/imgsUrl", imgsUrlHandler)
	mux.HandleFunc("/ok", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(200)
		fmt.Fprintf(w, "OK")
	})
	mux.HandleFunc("/", promptHandler)

	log.Fatal(http.ListenAndServe(":14736", mux))

	/*
		srcs, _ := extractImageSrcs("http://bing.com/images/search?q=green")
		fmt.Printf("%#v\n", srcs)
		return
		// xmlpath.MustCompile("asd")
		// path := xmlpath.MustCompile("//a")

		// xpath := `//a[@class="yt-uix-tile-link"]`
		xpath := `//img`

		// if page, err := ioutil.ReadFile("yt-sample.html"); err != nil {
		if page, err := ioutil.ReadFile("search?q=china"); err != nil {
			log.Fatal("file not read")
			// } else if root, err := xmlpath.Parse(bytes.NewBuffer(file)); err != nil {
		} else if root, err := gokogiri.ParseHtml(page); err != nil {
			log.Fatal(err)
			// } else if value, ok := path.String(root); !ok {
		} else if value, err := root.Search(xpath); err != nil {
			log.Fatal("xpath not found: ")
		} else {
			// fmt.Printf("Found: %#v", value)
			for i, node := range value {
				// fmt.Printf("Found %d: %#v\n", i, node)
				_ = i
				fmt.Printf("Found %v: \n", node.Attributes()["src"])

				if _url, err := url.Parse(node.Attributes()["src"].String()); err != nil {

				} else {
					fmt.Printf("url: %#v\n", _url)
				}
				// url = url.Parse(node.Attributes())
			}
		}*/
}

func proxyHandler(w http.ResponseWriter, req *http.Request) {
	_url := req.URL.RawQuery
	if response, err := http.Get(_url); err != nil {
		http.Error(w, fmt.Sprintf("error fetching url: %s", _url), 400)
	} else {
		// w.Header().Set("Content-Type", "")
		w.WriteHeader(200)
		// bytes.NewBuffer(response).WriteTo(w)
		// response.Write(w) TODO write response directly?
		bodyReq, _ := ioutil.ReadAll(response.Body)
		bytes.NewBuffer(bodyReq).WriteTo(w)
		// contentType :=
	}
}

func promptHandler(w http.ResponseWriter, req *http.Request) {
	var html = `
<HTML>
<HEAD>
<TITLE>Image proxy</TITLE>
</HEAD>
<BODY>
<SCRIPT LANGUAGE="JAVASCRIPT" TYPE="TEXT/JAVASCRIPT">
<!--
query = window.prompt("Enter image search query", "san ramon CA");
// window.location = encodeURI("/lucky?"+query);
bingBase = "http://bing.com/images/search?q="
window.location = encodeURI("/imgsUrl?"+bingBase+query);//redirect
//-->
</SCRIPT>
</BODY>
</HTML>`
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(200)
	fmt.Fprintf(w, html)
}

func imgsUrlHandler(w http.ResponseWriter, req *http.Request) {
	_url := req.URL.RawQuery
	log.Printf("imsUrlHandler")
	if srcs, err := extractImageSrcs(_url); err != nil {
		log.Printf("imsUrlHandler err")
		http.Error(w, fmt.Sprintf("error fetching url: %s", _url), 400)
	} else {
		HTML_FMT := `
<HTML>
<HEAD>
<TITLE>%s</TITLE>
%s
</BODY>
</HTML>`
		// <img src="smiley.gif" alt="Smiley face" height="42" width="42">
		IMGTAG_FMT := "<img src=%s>"
		imgTags := make([]string, len(srcs))
		for i, src := range srcs {
			var newSrc string
			if !strings.HasPrefix(src, "data:") { //don't proxy literal image data
				newSrc = "/proxy?" + src
			} else {
				newSrc = src
			}
			tag := fmt.Sprintf(IMGTAG_FMT, newSrc)
			imgTags[i] = tag
		}
		html := fmt.Sprintf(HTML_FMT, _url, strings.Join(imgTags, "\n"))
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(200)
		fmt.Fprintf(w, html) //TODO use buffer
	}
}

func extractImageSrcs(_url string) ([]string, error) {
	imgXpath := `//img`

	if urlHost, err := url.Parse(_url); err != nil {
		return nil, fmt.Errorf("bad url %s: %s", _url, err)
	} else if response, err := http.Get(_url); err != nil {
		return nil, fmt.Errorf("open problem: %s", err)
	} else if html, err := ioutil.ReadAll(response.Body); err != nil {
		return nil, fmt.Errorf("read problem: %s", err)
	} else if doc, err := gokogiri.ParseHtml(html); err != nil {
		return nil, fmt.Errorf("parse problem: %s", err)
	} else if imgs, err := doc.Search(imgXpath); err != nil {
		log.Fatal("xpath not found: ")
		return nil, fmt.Errorf("xpath problem: %s", err)
	} else {
		srcs := make([]string, 0, 100)
		for i, node := range imgs {
			// fmt.Printf("Found %d: %#v\n", i, node)
			_ = i
			// TODO check src
			if srcUrl, err := url.Parse(node.Attributes()["src"].String()); err != nil {

			} else {
				if srcUrl.Host == "" {
					srcUrl.Host = urlHost.Host
				}
				srcs = append(srcs, srcUrl.String())
			}
			// url = url.Parse(node.Attributes())
		}
		return srcs, nil
	}
}

// Local Variables:
// compile-cmd: "go run main.go "
// End:
