package main

import (
	"bufio"
	"fmt"
	_ "github.com/influxdata/influxdb1-client" // this is important because of the bug in go mod
	client "github.com/influxdata/influxdb1-client/v2"
	"github.com/memcachier/mc"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	//sensorFileName         = os.Getenv("PI_SENSOR_FILENAME")
	//influxConnectionString = os.Getenv("INFLUX_DB")
	memcachedConn          = "mc5.dev.ec2.memcachier.com:11211"
	influxConnectionString = "http://localhost:8086"
	sensorFileName         = "/home/pi/w1_slave_test"
)

func main() {
	startRecording()
	defer rec()
}

func startRecording() {
	var wg0 sync.WaitGroup
	var wg1 sync.WaitGroup
	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr: influxConnectionString,
	})
	m := mc.NewMC(memcachedConn, "F59154", "25565BCD30B9353C04E5EAB794735F1E")
	defer m.Quit()
	if err != nil {
		fmt.Println("Error creating InfluxDB Client: ", err.Error())
	}
	defer c.Close()

	wg0.Add(1)
	// Registers the sensor data in InfluxDB
	go func() {
		for {
			wg1.Add(1)
			go func() {
				bp, err := client.NewBatchPoints(client.BatchPointsConfig{
					Database:  "vistanave",
					Precision: "ms",
				})
				if err != nil {
					fmt.Println("Error creating BatchPoints: ", err.Error())
				}
				file, err := os.Open(sensorFileName)
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
				tags := map[string]string{"nombre_fermentador": "oriente"}
				for i := 0; i < 10; i++ {
					fields := map[string]interface{}{
						"temperatura": floatValue,
					}
					pt, err := client.NewPoint("fermentadores", tags, fields, time.Now())
					if err != nil {
						fmt.Println("Error: ", err.Error())
					}
					bp.AddPoint(pt)
					time.Sleep(100 * time.Second)
				}
				err = c.Write(bp)
				if err != nil {
					fmt.Println(err)
				}
				wg1.Done()
			}()
			wg1.Wait()
		}
		// Unreachable code
		wg0.Done()
	}()

	// Creates a new datapoint for the current IP
	go func() {
		for {
			res, err := http.Get("https://api.ipify.org")
			if err != nil {
				fmt.Println("Couldn't get IP: ", err.Error())
			}
			ip, err := ioutil.ReadAll(res.Body)
			defer res.Body.Close()
			if err != nil {
				fmt.Println("failure to read response body: ", err.Error())
			}
			_, err = m.Set("ip", string(ip), 0, 0, 0)
			if err != nil {
				fmt.Println("failure to store IP in memcached: ", err.Error())
			}
			time.Sleep(3 * time.Hour)
		}
		// Unreachable code
		wg0.Done()
	}()

	wg0.Wait()
}

func rec() {
	if r := recover(); r != nil {
		fmt.Println("Panic: recovering. Restarting ", r)
	}
	time.Sleep(2 * time.Minute)
	startRecording()
}
