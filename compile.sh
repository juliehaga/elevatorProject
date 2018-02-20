#!/bin/bash

#gnome-terminal -e "./Simulator/SimElevatorServer --port 15000" 
#gnome-terminal -e "./Simulator/SimElevatorServer --port 15020" 
#gnome-terminal -e "./Simulator/SimElevatorServer --port 15030" 


#gccgo src/elevio/elevator_io.go src/network/conn/conn.go src/network/broadcast/broadcast.go src/network/peers/peers.go src/network/localip/localip.go -o main.go



#go build -o src/elevio/elevator_io.go
#go build -o src/network/conn/conn.go 
#go build -o src/network/broadcast/broadcast.go 
#go build -o src/network/peers/peers.go 
#go build -o src/network/localip/localip.go

#go build -o src/main.go

#gccgo src/main.go
go run src/main.go
