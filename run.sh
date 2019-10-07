#!/bin/bash
# docker run -d --name=grafana -p 3000:3000 grafana/grafana 

# docker run -p 8086:8086  -d -v /tmp:/var/lib/influxdb influxdb

export PI_SENSOR_FILENAME="/home/federico/w1_slave"
#export PI_SENSOR_FILENAME="/home/pi/w1_slave_test"
export INFLUX_DB="http://localhost:8086"
export MEMCACHED="mc5.dev.ec2.memcachier.com:11211"

go run main.go
#vistanave
