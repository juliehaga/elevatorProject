package fsm

import(
	"../elevStateMap"
	"../elevio"
	"../config"
	"time"
	"math"
	"fmt"
)

var state ElevState

type ElevState int
const(
	IDLE            = 0
	MOVING          = 1
	DOOR_OPEN	    = 2
)

const DOOR_TIME 	= 2
const IDLE_TIME 	= 2


func Fsm(motorChan chan elevio.MotorDirection, doorLampChan chan bool, floorChan chan int, buttonLampChan chan elevio.ButtonLamp, mapChangesChan chan elevStateMap.ElevStateMap, newOrderChan chan elevio.ButtonEvent){
	doorTimer := time.NewTimer(time.Second * DOOR_TIME)
	doorTimer.Stop()

	idleTimer := time.NewTimer(time.Second * DOOR_TIME)
	idleTimer.Stop()



	for{
		//checkIfAcceptOrder(mapChangesChan, buttonLampChan, ackOrderChan)
		select{
		case  floor := <- floorChan:
			fmt.Printf("floor event\n")
			eventNewFloor(motorChan, doorLampChan, doorTimer, mapChangesChan, buttonLampChan, floor, idleTimer)
			
		case buttonPushed := <- newOrderChan:
			fmt.Printf("buttonpushed\n")
			eventNewAckOrder(buttonLampChan, motorChan, doorLampChan, doorTimer, mapChangesChan, buttonPushed, idleTimer)
			
		case <- doorTimer.C:
			fmt.Printf("door timeout\n")
			eventDoorTimeout(doorLampChan, mapChangesChan, idleTimer)
			
		//case <- idleTimer.C:
		//	eventIdleTimeout(motorChan, mapChangesChan, newOrderChan, doorLampChan, doorTimer, buttonLampChan)
		//	idleTimer.Reset(time.Second * IDLE_TIME)
			

		}
		

	}
}


func eventIdleTimeout(motorChan chan elevio.MotorDirection, mapChangesChan chan elevStateMap.ElevStateMap, newOrderChan chan elevio.ButtonEvent,  doorLampChan chan bool, doorTimer *time.Timer, buttonLampChan chan elevio.ButtonLamp) {
	currentMap := elevStateMap.GetLocalMap()
	for f:= 0; f < config.NUM_FLOORS; f++{
		for b:= elevio.BT_HallUp; b <= elevio.BT_Cab; b++{
				if currentMap[config.My_ID].Orders[f][b] == elevStateMap.OT_OrderPlaced {
					if shouldStop(currentMap){
						motorChan <- elevio.MD_Stop
						doorLampChan <- true
						doorTimer.Reset(time.Second * DOOR_TIME)

						currentMap[config.My_ID].Door = true
						orderCompleted(&currentMap, buttonLampChan, mapChangesChan)
						
						currentMap[config.My_ID].IDLE = false
						state = DOOR_OPEN
						mapChangesChan <- currentMap
					} else {
						motorDir := chooseDirection(&currentMap)
						
						if motorDir != elevio.MD_Stop {
							motorChan <- motorDir
							state = MOVING
							currentMap[config.My_ID].IDLE = false

						}
						mapChangesChan <- currentMap

					
					}
				}
			
	}


		}
	}



func eventNewFloor(motorChan chan elevio.MotorDirection, doorLampChan chan bool, doorTimer *time.Timer, mapChangesChan chan elevStateMap.ElevStateMap, buttonLampChan chan elevio.ButtonLamp, floor int, idleTimer *time.Timer){
	currentMap := elevStateMap.GetLocalMap()
	if floor != -1 {
		currentMap[config.My_ID].CurrentFloor = floor
	}
	idleTimer.Reset(time.Second * IDLE_TIME)

	switch(state){
		case MOVING:
			if(shouldStop(currentMap)){
				motorChan <- elevio.MD_Stop
				doorLampChan <- true
				doorTimer.Reset(time.Second * DOOR_TIME)

				currentMap[config.My_ID].Door = true
				orderCompleted(&currentMap, buttonLampChan, mapChangesChan)
				currentMap[config.My_ID].IDLE = false
				mapChangesChan <- currentMap
				state = DOOR_OPEN
			}
	}
}

func eventDoorTimeout(doorLampChan chan bool, mapChangesChan chan elevStateMap.ElevStateMap, idleTimer *time.Timer){
	currentMap := elevStateMap.GetLocalMap()
	switch(state){
		case DOOR_OPEN:
			doorLampChan <- false
			currentMap[config.My_ID].Door = false
			mapChangesChan <- currentMap
			currentMap[config.My_ID].IDLE = true
			state = IDLE	
	}
	idleTimer.Reset(time.Second * IDLE_TIME)
}


func eventNewAckOrder(buttonLampChan chan elevio.ButtonLamp, motorChan chan elevio.MotorDirection, doorLampChan chan bool, doorTimer *time.Timer, mapChangesChan chan elevStateMap.ElevStateMap, buttonPushed elevio.ButtonEvent, idleTimer *time.Timer){
	currentMap := elevStateMap.GetLocalMap()
	buttonLampChan <- elevio.ButtonLamp{buttonPushed.Floor, buttonPushed.Button, true}
	currentMap[config.My_ID].Orders[buttonPushed.Floor][buttonPushed.Button] = elevStateMap.OT_OrderPlaced

	switch(state){
		case IDLE:

			if shouldStop(currentMap) {
				doorLampChan <- true	
				currentMap[config.My_ID].Door = true
				orderCompleted(&currentMap, buttonLampChan, mapChangesChan)
				mapChangesChan <- currentMap
				doorTimer.Reset(time.Second * DOOR_TIME)
				currentMap[config.My_ID].IDLE = false
				state = DOOR_OPEN

				
			}else{
				motorDir := chooseDirection(&currentMap)
				mapChangesChan <- currentMap
				if motorDir != elevio.MD_Stop {
					motorChan <- motorDir
					currentMap[config.My_ID].IDLE = false
					state = MOVING
				}
			}			
	}
	idleTimer.Reset(time.Second * IDLE_TIME)
}

func shouldStop(elevMap elevStateMap.ElevStateMap) bool{
	if !orderInThisFloor(elevMap[config.My_ID].CurrentFloor, elevMap){
		return false
	}
	if elevMap[config.My_ID].Orders[elevMap[config.My_ID].CurrentFloor][elevio.BT_Cab]==elevStateMap.OT_OrderPlaced{
		return true
	}


	switch elevMap[config.My_ID].CurrentDir{
		case elevStateMap.ED_Up:
			//order on current floor and no orders above
			if elevMap[config.My_ID].Orders[elevMap[config.My_ID].CurrentFloor][elevio.BT_HallUp]==elevStateMap.OT_OrderPlaced{
				return true
			} else if !ordersAbove(elevMap) && elevMap[config.My_ID].Orders[elevMap[config.My_ID].CurrentFloor][elevio.BT_HallDown]==elevStateMap.OT_OrderPlaced{
				return true }

		case elevStateMap.ED_Down:
			//order on current floor and no orders below
		 	if elevMap[config.My_ID].Orders[elevMap[config.My_ID].CurrentFloor][elevio.BT_HallDown]==elevStateMap.OT_OrderPlaced{
		 		return true
		 	} else if !ordersBelow(elevMap) && elevMap[config.My_ID].Orders[elevMap[config.My_ID].CurrentFloor][elevio.BT_HallUp]==elevStateMap.OT_OrderPlaced {
			 	return true 
			}
	}
	return false
}

func ordersAbove(elevMap elevStateMap.ElevStateMap) bool{
	for f := elevMap[config.My_ID].CurrentFloor + 1; f<config.NUM_FLOORS; f++{
		for b := elevio.BT_HallUp; b<= elevio.BT_Cab; b++{ 
			if elevMap[config.My_ID].Orders[f][b] == elevStateMap.OT_OrderPlaced{
				return true
			}
		}
	}

	return false
}

func ordersBelow(elevMap elevStateMap.ElevStateMap) bool{
	for f := elevMap[config.My_ID].CurrentFloor - 1; f>=0; f--{
		for b := elevio.BT_HallUp; b<= elevio.BT_Cab; b++{ 
			if elevMap[config.My_ID].Orders[f][b] == elevStateMap.OT_OrderPlaced {
				return true
			}
		}
	}
	return false
}




func orderCompleted(elevMap *elevStateMap.ElevStateMap, buttonLampChan chan elevio.ButtonLamp, mapChangesChan chan elevStateMap.ElevStateMap){

	if elevMap[config.My_ID].Orders[elevMap[config.My_ID].CurrentFloor][elevio.BT_Cab] == elevStateMap.OT_OrderPlaced{
		elevMap[config.My_ID].Orders[elevMap[config.My_ID].CurrentFloor][elevio.BT_Cab] = elevStateMap.OT_NoOrder
		buttonLampChan <- elevio.ButtonLamp{elevMap[config.My_ID].CurrentFloor, elevio.BT_Cab, false}
	}


	switch(elevMap[config.My_ID].CurrentDir){
		case elevStateMap.ED_Up: 
			if elevMap[config.My_ID].Orders[elevMap[config.My_ID].CurrentFloor][elevio.BT_HallUp] == elevStateMap.OT_OrderPlaced{
				elevMap[config.My_ID].Orders[elevMap[config.My_ID].CurrentFloor][elevio.BT_HallUp] = elevStateMap.OT_NoOrder
				buttonLampChan <-  elevio.ButtonLamp{elevMap[config.My_ID].CurrentFloor, elevio.BT_HallUp, false}
			} else if elevMap[config.My_ID].Orders[elevMap[config.My_ID].CurrentFloor][elevio.BT_HallDown] == elevStateMap.OT_OrderPlaced{
				elevMap[config.My_ID].Orders[elevMap[config.My_ID].CurrentFloor][elevio.BT_HallDown] = elevStateMap.OT_NoOrder
				buttonLampChan <-  elevio.ButtonLamp{elevMap[config.My_ID].CurrentFloor, elevio.BT_HallDown, false}
			}
			
		case elevStateMap.ED_Down:
			if elevMap[config.My_ID].Orders[elevMap[config.My_ID].CurrentFloor][elevio.BT_HallDown] == elevStateMap.OT_OrderPlaced{
				elevMap[config.My_ID].Orders[elevMap[config.My_ID].CurrentFloor][elevio.BT_HallDown] = elevStateMap.OT_NoOrder
				buttonLampChan <-  elevio.ButtonLamp{elevMap[config.My_ID].CurrentFloor, elevio.BT_HallDown, false}
			} else if elevMap[config.My_ID].Orders[elevMap[config.My_ID].CurrentFloor][elevio.BT_HallUp] == elevStateMap.OT_OrderPlaced{
				elevMap[config.My_ID].Orders[elevMap[config.My_ID].CurrentFloor][elevio.BT_HallUp] = elevStateMap.OT_NoOrder
				buttonLampChan <-  elevio.ButtonLamp{elevMap[config.My_ID].CurrentFloor, elevio.BT_HallUp, false}
			}
			
	}
}



func orderInThisFloor( floor int, elevMap elevStateMap.ElevStateMap) bool{
	for b := elevio.BT_HallUp; b <= elevio.BT_Cab; b++ {
		if elevMap[config.My_ID].Orders[floor][b] == elevStateMap.OT_OrderPlaced {
			return true
		
		}
	}
	return false
}


func chooseDirection(elevMap *elevStateMap.ElevStateMap) elevio.MotorDirection{
	switch elevMap[config.My_ID].CurrentDir{
		case elevStateMap.ED_Up: 
			if ordersAbove(*elevMap){
				for f:= elevMap[config.My_ID].CurrentFloor + 1; f < config.NUM_FLOORS; f++{
					if  orderInThisFloor(f, *elevMap) && nearestElevator(*elevMap, f){
						elevMap[config.My_ID].CurrentDir = elevStateMap.ED_Up
						return elevio.MD_Up
					}
				}
				
			} else if ordersBelow(*elevMap){
				for f:= elevMap[config.My_ID].CurrentFloor - 1; f >= 0; f--{
					if orderInThisFloor(f, *elevMap) && nearestElevator(*elevMap, f){
						elevMap[config.My_ID].CurrentDir = elevStateMap.ED_Down
						return elevio.MD_Down
					}
				}
			} else {
				return elevio.MD_Stop
			}
		case elevStateMap.ED_Down:
			if ordersBelow(*elevMap){
				for f:= elevMap[config.My_ID].CurrentFloor - 1; f >= 0; f--{
					if orderInThisFloor(f, *elevMap) && nearestElevator(*elevMap, f){
						elevMap[config.My_ID].CurrentDir = elevStateMap.ED_Down
						return elevio.MD_Down
					}
				}

			} else if ordersAbove(*elevMap){
				for f:= elevMap[config.My_ID].CurrentFloor + 1; f < config.NUM_FLOORS; f++{
					if orderInThisFloor(f, *elevMap) && nearestElevator(*elevMap, f){
						elevMap[config.My_ID].CurrentDir = elevStateMap.ED_Up
						return elevio.MD_Up
					}
				}
			} else {
				return elevio.MD_Stop
			}
	}
	return elevio.MD_Stop
}



func nearestElevator(elevMap elevStateMap.ElevStateMap, floor int) bool{

 	myDist := int(math.Abs(float64(elevMap[config.My_ID].CurrentFloor - floor)))
 	fmt.Printf("jeg er %v etasjer unna %v \n", myDist, floor)

 	elevStateMap.PrintMap(elevMap)



 	if elevMap[config.My_ID].CurrentFloor < floor { 
 		fmt.Printf("jeg står under \n")
	 	for e := 0; e<config.NUM_ELEVS; e++ {
		 	if e != config.My_ID{	
		 		distElev := int(math.Abs(float64(elevMap[e].CurrentFloor - floor)))
		 		fmt.Printf("Den andre er %v etasjer unna\n", distElev)
		 		if distElev < myDist{
		 			fmt.Printf("den har kortere\n")
		 			if elevMap[e].CurrentFloor < floor && (elevMap[e].CurrentDir == elevStateMap.ED_Up || elevMap[e].IDLE ){
		 				fmt.Printf("den andre tar den\n")
		 				return false
		 			} else if elevMap[e].CurrentFloor > floor && (elevMap[e].CurrentDir == elevStateMap.ED_Down || elevMap[e].IDLE ) {
		 				fmt.Printf("den andre tar den\n")
		 				return false
		 			}
		 		} else if myDist == distElev && (elevMap[e].CurrentDir == elevStateMap.ED_Up || elevMap[e].IDLE){
		 			if e > config.My_ID{
		 				return false
		 			}
		 		}
		 	}
		 }
 	} else if elevMap[config.My_ID].CurrentFloor > floor {
	 		fmt.Printf("jeg står over \n")
		 	for e := 0; e<config.NUM_ELEVS; e++ {
			 	if e != config.My_ID{	
			 		distElev := int(math.Abs(float64(elevMap[e].CurrentFloor - floor)))
			 		if distElev < myDist{
			 			fmt.Printf("den har kortere\n")
			 			if elevMap[e].CurrentFloor > floor && (elevMap[e].CurrentDir == elevStateMap.ED_Down || elevMap[e].IDLE ){
			 				fmt.Printf("den andre tar den\n")
			 				return false
			 			} else if elevMap[e].CurrentFloor < floor && (elevMap[e].CurrentDir == elevStateMap.ED_Up || elevMap[e].IDLE ) {
		 					fmt.Printf("den andre tar den\n")
		 					return false
		 				}
			 		}  else if myDist == distElev && (elevMap[e].CurrentDir == elevStateMap.ED_Down || elevMap[e].IDLE){
		 			if e > config.My_ID{
		 				return false
		 			}
		 		}
			 	}
			}
	}

	fmt.Printf("jeg var nærmest\n")
 	return true
}
