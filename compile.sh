#!/bin/bash

gnome-terminal -e "./Simulator/SimElevatorServer --port 15000" 
gnome-terminal -e "./Simulator/SimElevatorServer --port 15020" 
gnome-terminal -e "./Simulator/SimElevatorServer --port 15030" 


gccgo src/elevio/elevator_io.go src/network/conn/conn.go src/network/broadcast/broadcast.go src/network/peers/peers.go src/network/localip/localip.go -o main.go


