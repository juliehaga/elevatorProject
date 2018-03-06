package fsm

import(
	"../elevStateMap"
	"../elevio"
	"../config"
	"time"
	"math"
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
			eventNewFloor(motorChan, doorLampChan, doorTimer, mapChangesChan, buttonLampChan, floor, idleTimer)
			
		case buttonPushed := <- newOrderChan:
			eventNewAckOrder(buttonLampChan, motorChan, doorLampChan, doorTimer, mapChangesChan, buttonPushed, idleTimer)
			
		case <- doorTimer.C:
			eventDoorTimeout(doorLampChan, mapChangesChan, idleTimer)
			
		case <- idleTimer.C:
			eventIdleTimeout(motorChan, mapChangesChan, newOrderChan, doorLampChan, doorTimer, buttonLampChan)
			idleTimer.Reset(time.Second * IDLE_TIME)
			

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
						mapChangesChan <- currentMap
						state = DOOR_OPEN
					} else {
						motorDir := chooseDirection(&currentMap)
						mapChangesChan <- currentMap
						if motorDir != elevio.MD_Stop {
							motorChan <- motorDir
							state = MOVING

					}

					
					}
				}
			
	}


		}
	}



func eventNewFloor(motorChan chan elevio.MotorDirection, doorLampChan chan bool, doorTimer *time.Timer, mapChangesChan chan elevStateMap.ElevStateMap, buttonLampChan chan elevio.ButtonLamp, floor int, idleTimer *time.Timer){
	currentMap := elevStateMap.GetLocalMap()
	if floor != -1 {
		currentMap[config.My_ID].CurrentFloor = floor
		mapChangesChan <- currentMap

	}
	switch(state){
		case MOVING:
			if(shouldStop(currentMap)){
				motorChan <- elevio.MD_Stop
				doorLampChan <- true
				doorTimer.Reset(time.Second * DOOR_TIME)

				currentMap[config.My_ID].Door = true
				orderCompleted(&currentMap, buttonLampChan, mapChangesChan)
				mapChangesChan <- currentMap
				state = DOOR_OPEN
			}
		case DOOR_OPEN:
			
	}
	idleTimer.Reset(time.Second * IDLE_TIME)
}

func eventDoorTimeout(doorLampChan chan bool, mapChangesChan chan elevStateMap.ElevStateMap, idleTimer *time.Timer){
	currentMap := elevStateMap.GetLocalMap()
	switch(state){
		case DOOR_OPEN:
			doorLampChan <- false
			currentMap[config.My_ID].Door = false
			mapChangesChan <- currentMap
			state = IDLE	
	}
	idleTimer.Reset(time.Second * IDLE_TIME)
}


func eventNewAckOrder(buttonLampChan chan elevio.ButtonLamp, motorChan chan elevio.MotorDirection, doorLampChan chan bool, doorTimer *time.Timer, mapChangesChan chan elevStateMap.ElevStateMap, buttonPushed elevio.ButtonEvent, idleTimer *time.Timer){
	currentMap := elevStateMap.GetLocalMap()
	buttonLampChan <- elevio.ButtonLamp{buttonPushed.Floor, buttonPushed.Button, true}
	currentMap[config.My_ID].Orders[buttonPushed.Floor][buttonPushed.Button] = elevStateMap.OT_OrderPlaced
	mapChangesChan <- currentMap

	switch(state){
		case IDLE:

			if shouldStop(currentMap) {
				doorLampChan <- true	
				currentMap[config.My_ID].Door = true
				orderCompleted(&currentMap, buttonLampChan, mapChangesChan)
				mapChangesChan <- currentMap
				
				doorTimer.Reset(time.Second * DOOR_TIME)
				state = DOOR_OPEN
			}else{
				motorDir := chooseDirection(&currentMap)
				mapChangesChan <- currentMap
				if motorDir != elevio.MD_Stop {
					motorChan <- motorDir
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
	mapChangesChan <- *elevMap
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
					if nearestElevator(*elevMap, f){
						elevMap[config.My_ID].CurrentDir = elevStateMap.ED_Up
						return elevio.MD_Up
					}
				}
				
			} else if ordersBelow(*elevMap){
				for f:= elevMap[config.My_ID].CurrentFloor - 1; f >= 0; f--{
					if nearestElevator(*elevMap, f){
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
					if nearestElevator(*elevMap, f){
						elevMap[config.My_ID].CurrentDir = elevStateMap.ED_Down
						return elevio.MD_Down
					}
				}

			} else if ordersAbove(*elevMap){
				for f:= elevMap[config.My_ID].CurrentFloor + 1; f < config.NUM_FLOORS; f++{
					if nearestElevator(*elevMap, f){
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
 	dist := int(math.Abs(float64(elevMap[config.My_ID].CurrentFloor - floor)));
 	for e := 0; e<config.NUM_ELEVS; e++ {
 		dist_e := int(math.Abs(float64(elevMap[config.My_ID].CurrentFloor - floor)));
 		if elevMap[e].CurrentFloor < floor && elevMap[e].CurrentDir == elevStateMap.ED_Up {
 			
 			if dist_e < dist {
 				fmt.Printf("Jeg er ikke nærmest, retning opp\n")
 				return false
 			}
 		}else if elevMap[e].CurrentFloor > floor && elevMap[e].CurrentDir == elevStateMap.ED_Down {
 			if dist_e < dist {
 				fmt.Printf("Jeg er ikke nærmest, retning ned\n")
 				return false
 			}
 		}
 	}
 	for e := 0; e<config.NUM_ELEVS; e++ {
	 	if elevMap[e].CurrentFloor - floor == dist {
	 			fmt.Printf("Samme distanse, skal prioritere på ID\n")
	 			return config.My_ID <= e 

	 		}
	}
	fmt.Printf("Skal ikke komme hit\n")
 	return true
}
