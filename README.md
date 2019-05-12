# go-max31855
This is a simple package for interfacing with the MAX31855 thermocouple amplifier over SPI. 

## Compatibility
Tested on a Raspberry Pi 3B using an  [Adafruit MAX31855 sensor](https://www.adafruit.com/product/269). It should be fine for other variants of Linux / RPi devices, but you might need to adjust the build settings (e.g. `GOARCH`, `GOARM`)

## Usage
Use this library in your project: `go get github.com/teebr/go-max31855`

There's not much to it: a `MAX31855` type with the file descriptor, internal temperature, thermocouple temperature and the timestamp when the measurements were read. There are three methods: `Open`, `Read` and `Close`. `Open` requires the name of the SPI device (e.g. `/dev/spidev0.0`). `Read()` then reads from the amplifier, checks for errors, and then populates the `Internal`, `Thermocouple` and `Timestamp` fields.

## Examples
There are two examples:
- `basic` which prints the temperatures every 100ms (about as fast as the sensor updates its readings),
- `streaming`, which creates a small webserver with a graph for the two temperatures on port 8080 of your device. This opens a websocket from the server to the browser to stream data, and uses [Plotly](https://github.com/plotly/plotly.js/) to view the data in real-time. There are a couple of basic widgets to control the data rate, time window and to pause streaming. [Gorilla Websockets](https://github.com/gorilla/websocket) are used for this example, so `go get github.com/gorilla/websocket` before building it.

### Build
I find it simpler to develop on a regular computer then cross-compile, rather than installing Go on the Raspberry Pi. To do this:

`GOOS=linux GOARCH=arm GOARM=7 go build src/github.com/teebr/go-max31855/examples/[NAME]/[NAME].go`
then `scp` the files to your Pi (for `streaming` make sure to copy the html file as well as the executable)
