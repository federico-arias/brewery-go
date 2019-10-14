#!/bin/bash
export PI="/media/federico/673b8ab6-6426-474b-87d3-71bff0fcebc3/"
export PIIP="192.168.8.200"

# installs fluentd to forward systemd logs to Elasticsearch
curl -L https://toolbelt.treasuredata.com/sh/install-debian-stretch-td-agent3.sh | sh

# installs fluentd plugin to read from journal logs
sudo /usr/sbin/td-agent-gem install fluent-plugin-systemd -v 1.0.1
sudo usermod -a -G systemd-journal td-agent # plugin needs access to journal
# installs fluentd plugin to write to elasticsearch 
sudo /usr/sbin/td-agent-gem install fluent-plugin-elasticsearch

# installs fluentbit (https://fluentbit.io/documentation/0.13/installation/raspberry_pi.html)
wget -qO - https://packages.fluentbit.io/fluentbit.key | sudo apt-key add -
echo "deb https://packages.fluentbit.io/raspbian jessie main" | sudo tee -a /etc/apt/sources.list
sudo apt-get update
sudo apt-get install td-agent-bit
sudo systemctl enable td-agent-bit 

# creates log to store systemd journal logs
sudo mkdir -p /var/log/journal
sudo mkdir -p /var/log/federico

# runs elasticsearch+kibana for local testing
sudo sysctl -w vm.max_map_count=262144
docker pull docker.elastic.co/elasticsearch/elasticsearch:7.4.0
docker pull docker.elastic.co/kibana/kibana:7.4.0
docker run -p 9200:9200 -p 9300:9300 \
	--detach \
	-e "discovery.type=single-node" \
	--name "elasticsearch_container" \
	docker.elastic.co/elasticsearch/elasticsearch:7.4.0
docker run --link elasticsearch_container:elasticsearch -p 5601:5601 \
	--name "kibana_container" \
	--detach \
	docker.elastic.co/kibana/kibana:7.4.0

# fix locale
sudo cp /usr/share/zoneinfo/America/Santiago /etc/localtime
sudo apt-get install ntp
date

# installs influxdb
curl -sL https://repos.influxdata.com/influxdb.key | sudo apt-key add -
sudo apt install apt-transport-https
echo "deb https://repos.influxdata.com/debian jessie stable" | sudo tee /etc/apt/sources.list.d/influxdb.list
sudo apt update && sudo apt-get -y -qq install influxdb
systemctl daemon-reload
sudo systemctl enable influxdb.service
sudo service influxdb start

# installs grafana 
# curl https://bintray.com/user/downloadSubjectPublicKey?username=bintray | sudo apt-key add -
# echo "deb https://dl.bintray.com/fg2it/deb jessie main" | sudo tee /etc/apt/sources.list.d/grafana.list
# sudo apt-get -y update
# sudo apt-get -y install grafana

# installs grafana for armv6 architecture
wget https://dl.grafana.com/oss/release/grafana-rpi_6.4.1_armhf.deb 
sudo dpkg -i grafana-rpi_6.4.1_armhf.deb 

# enables the systemd service so that Grafana starts at boot.
systemctl daemon-reload
sudo systemctl enable grafana-server.service
sudo systemctl start grafana-server.service

# creates vistanave database
influx -execute 'create database vistanave'

# copies network configuration
sudo cp conf/wpa_supplicant.conf ${PI}/etc/wpa_supplicant/wpa_supplicant.conf
sudo cp conf/interfaces ${PI}/etc/network/interfaces

# copies the systemd unit file to start app on startup
sudo scp brewery.service pi@${PIIP}:/lib/systemd/system/brewery.service

# copies test file
scp w1_slave_test pi@${PIIP}:/home/pi/w1_slave_test

# compiles to bytecode and copies it to the PI
env GOARCH=arm GOOS=linux go build -o "$HOME/.local/bin/vistanave"
scp "$HOME/.local/bin/vistanave" pi@${PIIP}:/usr/bin/vistanave

# lets systemd know of this new configuration
sudo systemctl daemon-reload

sudo systemctl enable brewery.service

sudo reboot
