package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/kikuchi-02/file-buffer-server/libs"
)

type Apple struct {
	Kind   string
	Price  int
	Origin string
}

func main() {
	source := libs.BufferSetup()
	settings := libs.LoadSettings()
	handler := func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			fmt.Println(err)
			return
		}
		var b Apple
		err = json.Unmarshal(body, &b)
		if err != nil {
			fmt.Println(err)
			return
		}
		str := fmt.Sprintf("kind: %s, price: %d, origin: %s", b.Kind, b.Price, b.Origin)
		fmt.Println(str)
		source <- str
		io.WriteString(w, "ok!\n")
	}

	http.HandleFunc(settings.Endpoint, handler)

	fmt.Printf("serve on Port:%d\n", settings.Port)
	http.ListenAndServe(fmt.Sprintf(":%d", settings.Port), nil)

}
