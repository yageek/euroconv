package main

import (
	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"github.com/unrolled/render"
	"github.com/yageek/euroconv/cache"
	"github.com/yageek/euroconv/eurobank"
	"log"
	"net/http"
	"os"
)

var router *mux.Router
var ren *render.Render

func init() {
	ren = render.New(render.Options{})

	router = mux.NewRouter()

	router.HandleFunc("/dayrate", func(w http.ResponseWriter, req *http.Request) {

		dayRate := cache.GetDayRate()

		if dayRate == nil {
			log.Println("Not present in cache...")
			cacheRate, err := eurobank.GetDayRate()

			if err != nil {
				log.Println("The eurobank query failed:", err)
				http.Error(w, "Err", http.StatusInternalServerError)
				return
			}

			dayRate = cacheRate

			err = cache.SetDayRate(dayRate)
			if err != nil {
				log.Println("Could not save data into the cache:", err)
			}
		}

		ren.JSON(w, http.StatusOK, dayRate)

	})
}

func main() {

	n := negroni.Classic()

	n.UseHandler(router)

	n.Run(":" + os.Getenv("PORT"))
}
