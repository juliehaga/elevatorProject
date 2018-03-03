package main

import (
	"./elevio"
	"./config"
	"./elevStateMap"
	"./fsm"
	"flag"
	"./network"
	"fmt"
)



func main() {

//command line arguments for port and id
	var port string	
	var id string
	flag.StringVar(&id, "id", "", "id")
    flag.StringVar(&port, "port", "15657", "portnumber")
    
    flag.Parse()	
    if id == "" {
    	//dafault ID
		id = "0"
	}
	fmt.Println("id:", id)
    fmt.Println("port:", port)


	

//init fuctions 
	config.InitConfig(id)
	elevStateMap.InitElevStateMap()
	elevio.Init("localhost:" + port, config.NUM_FLOORS)
	
//channels for communication between modules

	//hardware channels
	motorChan := make(chan elevio.MotorDirection)
	doorLampChan := make(chan bool)
    buttonChan := make(chan elevio.ButtonEvent)
    floorChan  := make(chan int)  
    buttonLampChan  := make(chan elevio.ButtonLamp)
    mapChangesChan := make(chan elevStateMap.ElevStateMap)


    // We make a channel for receiving updates on the id's of the peers that are
	//  alive on the network
    peerUpdateCh := make(chan network.PeerUpdate)
    // This could be used to signal that we are somehow "unavailable".
    peerTxEnable := make(chan bool)


    
    // We make channels for sending and receiving our custom data types
	elevMapTx := make(chan network.ElevMapMsg)
	elevMapRx := make(chan network.ElevMapMsg)
	// ... and start the transmitter/receiver pair on some port
	// These functions can take any number of channels! It is also possible to
	//  start multiple transmitters/receivers on the same port.
	


    go fsm.Fsm(motorChan, doorLampChan, floorChan, buttonLampChan, mapChangesChan, buttonChan)
    go elevio.Elevio(motorChan, doorLampChan, buttonChan, floorChan, buttonLampChan)
	go network.Transmitter(16569, elevMapTx)
	go network.Receiver(16569, elevMapRx)
    go network.PeerTransmitter(15647, id, peerTxEnable)
	go network.PeerReceiver(15647, peerUpdateCh)

    
   

	
	
	
	for {
		select {
		case p := <-peerUpdateCh:
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", p.Peers)
			fmt.Printf("  New:      %q\n", p.New)
			fmt.Printf("  Lost:     %q\n", p.Lost)

		case a := <- elevMapRx:
			fmt.Printf("Received: %#v\n", a)


		case elevMap:= <-mapChangesChan:
	
			elevStateMap.UpdateLocalMap(elevMap)
			
			elevStateMap.PrintMap()
			network.SendElevMap(elevMapTx, elevMap)
			//seeeend


			//en bestilling har skjedd
			//registrer at det har skjedd 

			//send beskjed til nettverket
		}

	}
}
