#!/bin/bash
export PI="/media/federico/673b8ab6-6426-474b-87d3-71bff0fcebc3/"
export PIIP="192.168.8.200"

# installs influxdb
curl -sL https://repos.influxdata.com/influxdb.key | sudo apt-key add -
sudo apt install apt-transport-https
echo "deb https://repos.influxdata.com/debian jessie stable" | sudo tee /etc/apt/sources.list.d/influxdb.list
sudo apt update && sudo apt-get -y -qq install influxdb
systemctl daemon-reload
sudo systemctl enable influxdb.service
sudo service influxdb start

# installs grafana 
curl https://bintray.com/user/downloadSubjectPublicKey?username=bintray | sudo apt-key add -
echo "deb https://dl.bintray.com/fg2it/deb jessie main" | sudo tee /etc/apt/sources.list.d/grafana.list
sudo apt-get -y update
sudo apt-get -y install grafana

# enables the systemd service so that Grafana starts at boot.
systemctl daemon-reload
sudo systemctl enable grafana-server.service

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
