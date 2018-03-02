package main

import (
	"./network/broadcast"
	"./network/localip"
	"./network/peers"
	"./elevio"
	"./config"
	"./elevStateMap"
	"./fsm"
	"flag"
	"os"
	"fmt"
)



func main() {

	var id string
	var port string	
	flag.StringVar(&id, "id", "", "id")
    flag.StringVar(&port, "port", "15657", "portnumber")
    
    flag.Parse()
    fmt.Println("id:", id)
    fmt.Println("port:", port)
	
	config.InitConfig(id)
	elevStateMap.InitElevStateMap()
	elevio.Init("localhost:" + port, config.NUM_FLOORS)
	

	
	if id == "" {
		localIP, err := localip.LocalIP()
		if err != nil {
			fmt.Println(err)
			localIP = "DISCONNECTED"
		}
		id = fmt.Sprintf("peer-%s-%d", localIP, os.Getpid())
	}

	motorChan := make(chan elevio.MotorDirection)
	doorLampChan := make(chan bool)
    buttonChan := make(chan elevio.ButtonEvent)
    floorChan  := make(chan int)  
    buttonLampChan  := make(chan elevio.ButtonLamp)
    peerUpdateCh := make(chan peers.PeerUpdate)
    peerTxEnable := make(chan bool)


    go fsm.Fsm(motorChan, doorLampChan, floorChan, buttonLampChan)
    go elevio.Elevio(motorChan, doorLampChan, buttonChan, floorChan, buttonLampChan)

    // We can disable/enable the transmitter after it has been started.
	// This could be used to signal that we are somehow "unavailable".

    go peers.Transmitter(15647, id, peerTxEnable)
	go peers.Receiver(15647, peerUpdateCh)
    
    

	// We make channels for sending and receiving our custom data types
	floorTx := make(chan FloorMsg)
	floorRx := make(chan FloorMsg)
	// ... and start the transmitter/receiver pair on some port
	// These functions can take any number of channels! It is also possible to
	//  start multiple transmitters/receivers on the same port.
	go broadcast.Transmitter(16569, floorTx)
	go broadcast.Receiver(16569, floorRx)

	
	
	
	for {
		select {
		case p := <-peerUpdateCh:
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", p.Peers)
			fmt.Printf("  New:      %q\n", p.New)
			fmt.Printf("  Lost:     %q\n", p.Lost)

		case a := <-floorRx:
			fmt.Printf("Received: %#v\n", a)
		
		}
	}
}
