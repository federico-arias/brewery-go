package main

import (
	"bufio"
	"fmt"
	_ "github.com/influxdata/influxdb1-client" // this is important because of the bug in go mod
	client "github.com/influxdata/influxdb1-client/v2"
	"github.com/memcachier/mc"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	//sensorFileName         = os.Getenv("PI_SENSOR_FILENAME")
	//influxConnectionString = os.Getenv("INFLUX_DB")
	//sensorFileName         = "/home/pi/w1_slave_test"
	memcachedConn          = "mc5.dev.ec2.memcachier.com:11211"
	influxConnectionString = "http://localhost:8086"
	sensorFileName         = "/sys/devices/w1_bus_master1/28-011313a1d9aa/w1_slave"
)

func main() {
	startRecording()
	defer rec()
}

func startRecording() {
	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr: influxConnectionString,
	})
	defer c.Close()

	// Creates a group of points
	tags := map[string]string{"nombre_fermentador": "oriente"}
	for {
		bp, err := client.NewBatchPoints(client.BatchPointsConfig{
			Database:  "vistanave",
			Precision: "ms",
		})
		if err != nil {
			fmt.Println("Error creating BatchPoints: ", err.Error())
		}
		for i := 0; i < 5; i++ {
			fields := map[string]interface{}{
				"temperatura": floatValue,
			}
			pt, err := client.NewPoint("fermentadores", tags, fields, time.Now())
			if err != nil {
				fmt.Println("Error: ", err.Error())
			}
			bp.AddPoint(pt)
			time.Sleep(time.Second)
		}
		err = c.Write(bp)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func readFromSensor() float64 {
	file, err := os.Open(sensorFileName)
	defer file.Close()
	if err != nil {
		fmt.Println("failure opening sensor file", err.Error())
	}
	fileScanner := bufio.NewScanner(file) //.Scan().Scan()
	fileScanner.Scan()
	fileScanner.Scan()
	line := fileScanner.Text()
	err = fileScanner.Err()
	if err != nil {
		fmt.Println("Error in reading from sensor: ", err.Error())
	}
	value := strings.Split(line, "=")[1]
	floatValue, err := strconv.ParseFloat(value, 64)
	if err != nil {
		fmt.Println("failure to parse float", err.Error())
	}
	if floatValue != 0.0 {
		floatValue = floatValue / 1000
	}
	return floatValue
}

func rec() {
	r := recover()
	fmt.Println("Panic: recovering. Restarting ", r)
	time.Sleep(2 * time.Minute)
	startRecording()
}
