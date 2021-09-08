package main

import (
	"flag"
	"fmt"
	"net/http"
	"net/rpc"

	"github.com/sanrentai/shorturl"
)

var (
	listenAddr = flag.String("http", ":8074", "http listen address")
	dataFile   = flag.String("file", "store.json", "data store file name")
	hostname   = flag.String("host", "localhost:8074", "http host name")
	masterAddr = flag.String("master", "", "RPC master address")
	rpcEnabled = flag.Bool("rpc", false, "enable RPC server")
)

var store shorturl.Store

func main() {
	flag.Parse()
	if *masterAddr != "" {
		store = shorturl.NewProxyStore(*masterAddr)
	} else {
		store = shorturl.NewURLStore(*dataFile)
	}
	if *rpcEnabled { // the master is the rpc server:
		rpc.RegisterName("Store", store)
		rpc.HandleHTTP()
	}
	http.HandleFunc("/", Redirect)
	http.HandleFunc("/add", Add)
	http.ListenAndServe(*listenAddr, nil)

}

func Redirect(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Path[1:]
	if key == "" {
		http.NotFound(w, r)
		return
	}
	var url string
	if err := store.Get(&key, &url); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, url, http.StatusFound)
}

func Add(w http.ResponseWriter, r *http.Request) {
	url := r.FormValue("url")
	if url == "" {
		fmt.Fprint(w, AddForm)
		return
	}
	var key string
	if err := store.Put(&url, &key); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "http://%s/%s", *hostname, key)
}

const AddForm = `
<html><body>
<form method="POST" action="/add">
URL: <input type="text" name="url">
<input type="submit" value="Add">
</form>
</body></html>
`
