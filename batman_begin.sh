#!/bin/sh
sudo modprobe batman-adv
sudo ip link set wlan0 down
sudo ip link set wlan0 mtu 1476
sudo iw wlan0 set type ibss
#sleep 1
sudo ip link set wlan0 up
#sleep 1
#iw dev wlan0 ibss leave
sudo iw wlan0 ibss join 131331 2437 fixed-freq 02:AA:BB:CC:DD:EE
#sleep 1
sudo batctl if add wlan0
#sleep 1
sudo ip link set bat0 mtu 1476
sudo ip link set up dev bat0
#sleep 1
#sudo ip addr add 172.27.0.2/16 dev bat0
#sudo ip addr add 10.9.6.3/16 dev bat0
