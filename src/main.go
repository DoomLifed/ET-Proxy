package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

func copyResponseHeader(dst, src http.Header)  {
	for key, valueSrc := range src {
		for _, value := range valueSrc {
			dst.Add(key, value)
		}
	}
}

func createHeader(header http.Header)  (http.Header, string){
	var addr string
	for key := range header{
		if key == "Addr"{
			addr = header.Get(key)
			header.Del(key)
			continue
		}else if key == "Proxy"{
			header.Del(key)
			continue
		}
	}
	return header, addr
}

func doForwardRequestToAnotherProxy(request *http.Request, response http.ResponseWriter,proto , proxy string)  {
	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		http.Error(response, err.Error(), http.StatusInternalServerError)

	}else {
		proxyReq , err := http.NewRequest(request.Method, proto+"://"+proxy, bytes.NewReader(body))
		proxyReq.Header = make(http.Header)
		for h, val := range request.Header {
			proxyReq.Header[h] = val
		}
		client := http.Client{}
		res, err := client.Do(proxyReq)
		if err != nil {
			http.Error(response, err.Error(), http.StatusBadGateway)
		}else {
			defer res.Body.Close()
			copyResponseHeader(response.Header(), res.Header)
			response.WriteHeader(res.StatusCode)
			_, _ = io.Copy(response, res.Body)
		}
	}

}
func inputRequest(response http.ResponseWriter, request *http.Request){
	proto := strings.Split(request.Proto, "/")[0]
	if proto != "HTTP" && proto != "HTTPS"{
		http.Error(response, "Bad protocol " + request.URL.Scheme, http.StatusBadRequest)
	}else {
		if request.Header.Get("proxy") == "none"{
			request.Header,  request.RemoteAddr= createHeader(request.Header)
			client := http.Client{}
			res , err := client.Do(request)
			if err!=nil{
				http.Error(response, err.Error(), res.StatusCode)
			}else {
				defer res.Body.Close()
				copyResponseHeader(response.Header(), res.Header)
				response.WriteHeader(res.StatusCode)
				_, _ = io.Copy(response, res.Body)
			}
		}else {
			proxy := strings.Split(request.Header.Get("proxy"), ",")
			newValueOfProxy := ""
			if len(proxy) > 1{
				for i:=1 ; i<len(proxy) ; i++{
					newValueOfProxy += proxy[i]
					if i != len(proxy)-1{
						newValueOfProxy += ","
					}
				}
				request.Header.Set("proxy", newValueOfProxy)
				doForwardRequestToAnotherProxy(request, response, proto, proxy[0])
			}else {
				fmt.Println(request.Header.Get("proxy"))
				request.Header.Set("proxy", "none")
				fmt.Println(request.Header.Get("proxy"))
				doForwardRequestToAnotherProxy(request, response, proto, proxy[0])
			}
		}
	}
}

func main() {
	http.HandleFunc("/", inputRequest)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
