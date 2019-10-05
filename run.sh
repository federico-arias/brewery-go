#!/bin/bash
docker run -d --name=grafana -p 3000:3000 grafana/grafana 

docker run -p 8086:8086 \
	      -d -v /tmp:/var/lib/influxdb \
		        influxdb

go run main.go
