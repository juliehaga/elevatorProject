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
	config.Init(id, port)
	elevio.Init("localhost:" + port, config.NUM_FLOORS)
	elevStateMap.InitElevStateMap()
	
	
//channels for communication between modules

	//hardware channels
	motorChan := make(chan elevio.MotorDirection)
	doorLampChan := make(chan bool)

    floorChan  := make(chan int)  
    buttonLampChan  := make(chan elevio.ButtonLamp)
    mapChangesChan := make(chan elevStateMap.ElevStateMap)
  
    newOrderChan := make(chan elevio.ButtonEvent)


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
	


    go fsm.Fsm(motorChan, doorLampChan, floorChan, buttonLampChan, mapChangesChan, newOrderChan)
    go elevio.Elevio(motorChan, doorLampChan, newOrderChan, floorChan, buttonLampChan)
	go network.Transmitter(16502, elevMapTx)
	go network.Receiver(16502, elevMapRx)
    go network.PeerTransmitter(15600, id, peerTxEnable)
	go network.PeerReceiver(15600, peerUpdateCh, mapChangesChan)

    
   

	
	
	
	for {
		select {
		case p := <-peerUpdateCh:

			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", p.Peers)
			fmt.Printf("  New:      %q\n", p.New)
			fmt.Printf("  Lost:     %q\n", p.Lost)
			//currentMap := elevStateMap.GetLocalMap()
			//elevStateMap.PrintMap(currentMap)





		case networkMapMsg := <- elevMapRx:

			//Når vi mottar melding bør vi sjekke at hardware er oppdatert?


			if networkMapMsg.ID != config.My_ID {
		
				elevStateMap.UpdateMapFromNetwork(networkMapMsg.ElevMap, newOrderChan, buttonLampChan)
				//currentMap := elevStateMap.GetLocalMap()
				//elevStateMap.PrintMap(currentMap)
				//elevStateMap.PrintMap(networkMapMsg.ElevMap)
			}
			
		case elevMap:= <-mapChangesChan:
			fmt.Printf("gjør forandring\n")	
	
			elevStateMap.UpdateLocalMap(elevMap)
			//fmt.Printf("MAP OPPDATERES lokalt\n \n\n")

			//currentMap := elevStateMap.GetLocalMap()

			//elevStateMap.PrintMap(currentMap)
			network.SendElevMap(elevMapTx, elevMap)

		}
	}
}