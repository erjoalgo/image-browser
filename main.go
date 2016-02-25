package main

import (
	//"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/erjoalgo/image-browser/Godeps/_workspace/src/github.com/moovweb/gokogiri"
)


var client http.Client
func main() {
     var internalProxyURL string
     var port string
	flag.StringVar(&internalProxyURL, "proxy", "http://proxy-src.research.ge.com:8080", "proxy for the server")
	//flag.StringVar(&internalProxyURL, "proxy", "", "proxy for the server")  
	flag.StringVar(&port, "port", os.Getenv("PORT"), "port server")
	flag.Parse()
	
	if port == "" {
		port = "14736"
	}
	if internalProxyURL != "" {
		log.Printf("using proxy: %s", internalProxyURL)
		client.Transport = &http.Transport{
			Proxy: func(*http.Request) (*url.URL, error) {
			return url.Parse(internalProxyURL)
			},
		}
	}
	log.Printf("using port: %s", port)
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
	log.Fatal(http.ListenAndServe(":"+port, mux))
}

func proxyHandler(w http.ResponseWriter, req *http.Request) {
	_url := req.URL.RawQuery
	if response, err := client.Get(_url); err != nil {
		http.Error(w, fmt.Sprintf("error fetching url %s:\n%s", _url, err), 400)
	} else {
		w.WriteHeader(200)

		//bodyReq, _ := ioutil.ReadAll(response.Body)
		//bytes.NewBuffer(bodyReq).WriteTo(w)

		//this is actually slower
		io.Copy(w, response.Body)
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
	if srcs, err := extractImageSrcs(_url); err != nil {
		http.Error(w, fmt.Sprintf("error fetching url: %s", _url), 400)
	} else {
		HTML_FMT_PRE := `<HTML><HEAD><TITLE>%s</TITLE><BODY>`	
		HTML_FMT_POST := `</BODY></HTML>`
		// <img src="smiley.gif" alt="Smiley face" height="42" width="42">
		//IMGTAG_FMT := "<img src=%s>"
		imgTags := make([]string, len(srcs))
		for i, src := range srcs {
			var newSrc string
			if !strings.HasPrefix(src, "data:") { //don't proxy literal image data
				newSrc = "/proxy?" + src
			} else {
				newSrc = src
			}
			//tag := fmt.Sprintf(IMGTAG_FMT, newSrc)
			tag := "<img src="+newSrc+">"
			imgTags[i] = tag
		}
		//html := fmt.Sprintf(HTML_FMT, _url, strings.Join(imgTags, "\n"))
		html := fmt.Sprintf(HTML_FMT_PRE, _url)+ strings.Join(imgTags, "\n") + HTML_FMT_POST
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(200)
		fmt.Fprint(w, html) //TODO use buffer
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
		for _, node := range imgs {
			if srcUrl, err := url.Parse(node.Attributes()["src"].String()); err != nil {

			} else {
				if srcUrl.Host == "" {
					srcUrl.Host = urlHost.Host
				}
				if srcUrl.Scheme == "" {
					srcUrl.Scheme = urlHost.Scheme
				}
				srcs = append(srcs, srcUrl.String())
			}
		}
		return srcs, nil
	}
}

// Local Variables:
// compile-cmd: "go run main.go "
// End:
