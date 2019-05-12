package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/teebr/go-max31855"
)

func main() {
	interrupt := make(chan os.Signal, 1) // for catching ctrl+c
	stop := make(chan bool, 1)
	signal.Notify(interrupt, os.Interrupt)
	var sensor max31855.MAX31855

	err := sensor.Open("/dev/spidev0.0")
	if err != nil {
		log.Fatal(err)
	}

	ticker := time.NewTicker(100 * time.Millisecond)
	go func() {
		for {
			select {
			case <-stop:
				log.Println("exiting loop")
				return
			case <-ticker.C:
				if sensor.Read() != nil {
					log.Fatal(err)
				}
				fmt.Println(sensor.Timestamp, sensor.Thermocouple, sensor.Internal)
			}
		}
	}()
	// more stuff could go here, e.g. push data to database
	<-interrupt
	stop <- true
	sensor.Close()
}
