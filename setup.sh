#!/bin/bash
sudo cp conf/wpa_supplicant.conf /etc/wpa_supplicant/wpa_supplicant.conf

# copies the systemd unit file to start app on startup
sudo cp brewery.service /lib/systemd/system/brewery.service

# lets systemd know of this new configuration
sudo systemctl daemon-reload

sudo systemctl enable brewery.service

sudo reboot


