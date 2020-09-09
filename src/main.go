package main

import (
	"fmt"
	"io"
	"net/http"
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
	for key, _:= range header{
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

func inputRequest(response http.ResponseWriter, request *http.Request){
	fmt.Println(request.RemoteAddr , request.RequestURI)
	if request.URL.Scheme == "http" && request.URL.Scheme == "https"{
		http.Error(response, "Bad protocol " + request.URL.Scheme, http.StatusBadRequest)
	}else {
		if request.Header.Get("proxy") == "none"{
			request.Header,  request.RemoteAddr= createHeader(request.Header)
			client := http.Client{}
			res, err := client.Do(request)
			if err!=nil{
				http.Error(response, err.Error(), res.StatusCode)
			}else {
				defer res.Body.Close()
				copyResponseHeader(response.Header(), res.Header)
				response.WriteHeader(res.StatusCode)
				io.Copy(response, res.Body)
			}
		}
	}
}

func main() {
	http.HandleFunc("/", inputRequest)
	http.ListenAndServe(":8080", nil)
}
