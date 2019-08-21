package gui

import (
	"github.com/gobuffalo/packr"
	"github.com/noah-blockchain/noah-go-node/log"
	"net/http"
)

func Run(addr string) {
	box := packr.NewBox("./html")

	http.Handle("/", http.FileServer(box))
	log.Error(http.ListenAndServe(addr, nil).Error())
}
