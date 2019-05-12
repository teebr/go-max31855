package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/teebr/go-max31855"
)

type wsData struct {
	X  time.Time `json:"x"`
	Y1 float64   `json:"y1"`
	Y2 float64   `json:"y2"`
}

var upgrader = websocket.Upgrader{} // use default options
var sensor max31855.MAX31855
var mutex = &sync.Mutex{}

// var stop chan bool

func pushData(w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true } //TODO: check r.RemoteAddr
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	// plot state is set by the JS client over WebSocket:
	// -1: close stream.
	// 0: pause stream
	// >0 period in ms to stream at.
	plotState := make(chan int)

	//routine to read from WS and update plotState channel
	go wsRead(conn, plotState)

	// wait for JS client to say things are good.
	reqestedSampleTime := <-plotState
	if reqestedSampleTime <= 0 {
		log.Printf("invalid start time: %d\n", reqestedSampleTime)
		return
	}
	log.Printf("sample time: %d\n", reqestedSampleTime)

	// start sending data
	var data wsData
	ticker := time.NewTicker(time.Duration(reqestedSampleTime) * time.Millisecond)

	for {
		select {
		case reqestedSampleTime = <-plotState:
			log.Printf("streamData: %d\n", reqestedSampleTime)
			if reqestedSampleTime < 0 {
				log.Println("requested stop")
			} else if reqestedSampleTime >= 100 {
				ticker = time.NewTicker(time.Duration(reqestedSampleTime) * time.Millisecond)
			}

		case <-ticker.C:
			if reqestedSampleTime > 0 {
				mutex.Lock()
				data.X = sensor.Timestamp
				data.Y1 = sensor.Thermocouple
				data.Y2 = sensor.Internal
				mutex.Unlock()

				log.Println(data)
				if err := conn.WriteJSON(data); err != nil {
					log.Println(err)
				}
			}
		}
	}
}

func wsRead(conn *websocket.Conn, control chan int) {
	for {
		mt, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("fatal in wsRead...")
			log.Println(err)
		}
		log.Println("read from ws:", string(message))
		if mt == 1 {
			// text data
			var ms int
			_, err := fmt.Sscanf(string(message), "start: %d", &ms)
			if err != nil {
				control <- 0
			} else {
				control <- ms
			}
		} else {
			control <- -1
			return
		}
	}
}

func runSensor() {
	ticker := time.NewTicker(50 * time.Millisecond) // faster than actual update.
	log.Println("starting loop")

	for {
		select {
		case <-ticker.C:
			mutex.Lock()
			if err := sensor.Read(); err != nil {
				log.Fatal(err)
			}
			mutex.Unlock()
		}
	}
}

func main() {
	err := sensor.Open("/dev/spidev0.0")
	if err != nil {
		log.Fatal(err)
	}
	defer sensor.Close()
	log.Println("sensor initialised")

	//generate sensor data periodically
	go runSensor()

	// websocket server
	addr := "0.0.0.0:8081"
	http.HandleFunc("/sensor", pushData)
	go func() {
		log.Fatal(http.ListenAndServe(addr, nil))
	}()

	// plot server:
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "max31855.html")
	})

	log.Fatal(http.ListenAndServe("0.0.0.0:8080", nil))

}
