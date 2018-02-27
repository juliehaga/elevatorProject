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

// We define some custom struct to send over the network.
// Note that all members we want to transmit must be public. Any private members
//  will be received as zero-values.

type FloorMsg struct {
	Message string
	Floor	int
}

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
	

	
	if id == "" {
		localIP, err := localip.LocalIP()
		if err != nil {
			fmt.Println(err)
			localIP = "DISCONNECTED"
		}
		id = fmt.Sprintf("peer-%s-%d", localIP, os.Getpid())
	}

	numFloors := 4
    
    
    elevio.Init("localhost:" + port, numFloors)
    elevio.ClearAllButtonLamps();
    
    var d elevio.MotorDirection = elevio.MD_Down
    
    elevio.SetMotorDirection(d)
	elevio.SetFloorIndicator(3)



    


   
	motorChan := make(chan elevio.MotorDirection)
	doorLampChan := make(chan bool)
    buttonChan := make(chan elevio.ButtonEvent)
    floorChan  := make(chan int)  
    buttonLampChan  := make(chan elevio.ButtonLamp)


    fsm.Fsm(motorChan, doorLampChan, floorChan, buttonLampChan)


    go elevio.PollButtons(buttonChan)
    go elevio.PollFloorSensor(floorChan)
    
    peerUpdateCh := make(chan peers.PeerUpdate)
	// We can disable/enable the transmitter after it has been started.
	// This could be used to signal that we are somehow "unavailable".
	peerTxEnable := make(chan bool)
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
			
        case a := <- floorChan:
    		fmt.Printf("Passed floor%+v\n", a)
    		floorMsg := FloorMsg{"Hello from " + id, a}
    		floorTx <- floorMsg
    		
            if a == numFloors-1 {
                d = elevio.MD_Down
            } else if a == 0 {
                d = elevio.MD_Up
            }
            elevio.SetMotorDirection(d)  
		}
		
	}
}
