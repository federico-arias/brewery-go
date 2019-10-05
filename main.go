package main

import (
	"bufio"
	"fmt"
	_ "github.com/influxdata/influxdb1-client" // this is important because of the bug in go mod
	client "github.com/influxdata/influxdb1-client/v2"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	//sensorFileName         = "/sys/bus/w1/devices/28-011313a1d9aa/w1_slave"
	sensorFileName         = "/home/federico/w1_slave"
	influxConnectionString = "http://localhost:8086"
)

func main() {
	var wg0 sync.WaitGroup
	var wg1 sync.WaitGroup
	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr: influxConnectionString,
	})
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
					time.Sleep(100 * time.Millisecond)
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
			bp, _ := client.NewBatchPoints(client.BatchPointsConfig{
				Database:  "vistanave",
				Precision: "h",
			})
			res, err := http.Get("https://api.ipify.org")
			if err != nil {
				fmt.Println("Couldn't get IP: ", err.Error())
			}
			ip, err := ioutil.ReadAll(res.Body)
			if err != nil {
				fmt.Println("failure to read response body: ", err.Error())
			}
			tags := map[string]string{"artefacto": "raspberry_pi"}
			fields := map[string]interface{}{"ip": ip}
			pt, err := client.NewPoint("ips", tags, fields, time.Now())
			if err != nil {
				fmt.Println("Error: ", err.Error())
			}
			bp.AddPoint(pt)
			err = c.Write(bp)
			if err != nil {
				fmt.Println(err)
			}
			defer res.Body.Close()
			time.Sleep(3 * time.Hour)
		}
		// Unreachable code
		wg0.Done()
	}()

	wg0.Wait()
}
