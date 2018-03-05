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
			fmt.Printf("NY FLOOR EVENT \n")
			eventNewFloor(motorChan, doorLampChan, doorTimer, mapChangesChan, buttonLampChan, floor, idleTimer)
			
			
		case buttonPushed := <- newOrderChan:
			fmt.Printf("\n \n \n \n EVENT ACKORDER\n")
			eventNewAckOrder(buttonLampChan, motorChan, doorLampChan, doorTimer, mapChangesChan, buttonPushed, idleTimer)
			
		case <- doorTimer.C:
			eventDoorTimeout(doorLampChan, mapChangesChan, idleTimer)
			
		case <- idleTimer.C:
			fmt.Printf("\n \n \n \n IDLE TIMEOUT\n")
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
					fmt.Printf("Det finnes flere bestillinger\n\n")
					

					if shouldStop(currentMap){
						motorChan <- elevio.MD_Stop
						doorLampChan <- true
						doorTimer.Reset(time.Second * DOOR_TIME)

						currentMap[config.My_ID].Door = true
						orderCompleted(&currentMap, buttonLampChan, mapChangesChan)
						mapChangesChan <- currentMap
						fmt.Printf("STATE FRA MOVING = DOOR OPEN\n")
						state = DOOR_OPEN
					} else {
						motorDir := chooseDirection(&currentMap)
						fmt.Printf("eventIDLETIMEOUT motor dir %v\n", motorDir)
						mapChangesChan <- currentMap
						if motorDir != elevio.MD_Stop {
							motorChan <- motorDir
							fmt.Printf("STATE = MOVING \n")
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
				fmt.Printf("STATE FRA MOVING = DOOR OPEN\n")
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
			fmt.Printf("State = IDLE\n")
			
	}
	idleTimer.Reset(time.Second * IDLE_TIME)
}


func eventNewAckOrder(buttonLampChan chan elevio.ButtonLamp, motorChan chan elevio.MotorDirection, doorLampChan chan bool, doorTimer *time.Timer, mapChangesChan chan elevStateMap.ElevStateMap, buttonPushed elevio.ButtonEvent, idleTimer *time.Timer){
	currentMap := elevStateMap.GetLocalMap()


	buttonLampChan <- elevio.ButtonLamp{buttonPushed.Floor, buttonPushed.Button, true}
	currentMap[config.My_ID].Orders[buttonPushed.Floor][buttonPushed.Button] = elevStateMap.OT_OrderPlaced
	elevStateMap.PrintMap(currentMap)
	mapChangesChan <- currentMap

	switch(state){
		case IDLE:



			if shouldStop(currentMap) {
				fmt.Printf("Bestilt etasje der vi står\n")
				doorLampChan <- true
				
				currentMap[config.My_ID].Door = true
				orderCompleted(&currentMap, buttonLampChan, mapChangesChan)
				mapChangesChan <- currentMap
				
				doorTimer.Reset(time.Second * DOOR_TIME)
				state = DOOR_OPEN
				fmt.Printf("STATE = DOOR OPEN\n")
			}else{
				motorDir := chooseDirection(&currentMap)
				fmt.Printf("Skal sette motorDir %v \n", motorDir)
				mapChangesChan <- currentMap
				if motorDir != elevio.MD_Stop {
					motorChan <- motorDir
					fmt.Printf("STATE = MOVING \n")
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
	fmt.Printf("Map for orders above\n")
	elevStateMap.PrintMap(elevMap)
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
	fmt.Printf("ordre completed\n")

	if elevMap[config.My_ID].Orders[elevMap[config.My_ID].CurrentFloor][elevio.BT_Cab] == elevStateMap.OT_OrderPlaced{
		elevMap[config.My_ID].Orders[elevMap[config.My_ID].CurrentFloor][elevio.BT_Cab] = elevStateMap.OT_NoOrder
		buttonLampChan <- elevio.ButtonLamp{elevMap[config.My_ID].CurrentFloor, elevio.BT_Cab, false}
	}


	switch(elevMap[config.My_ID].CurrentDir){
		case elevStateMap.ED_Up: 
			if elevMap[config.My_ID].Orders[elevMap[config.My_ID].CurrentFloor][elevio.BT_HallUp] == elevStateMap.OT_OrderPlaced{
				elevMap[config.My_ID].Orders[elevMap[config.My_ID].CurrentFloor][elevio.BT_HallUp] = elevStateMap.OT_NoOrder
				buttonLampChan <-  elevio.ButtonLamp{elevMap[config.My_ID].CurrentFloor, elevio.BT_HallUp, false}
				fmt.Printf("Retning opp skal slukke hallup\n")
			} else if elevMap[config.My_ID].Orders[elevMap[config.My_ID].CurrentFloor][elevio.BT_HallDown] == elevStateMap.OT_OrderPlaced{
				elevMap[config.My_ID].Orders[elevMap[config.My_ID].CurrentFloor][elevio.BT_HallDown] = elevStateMap.OT_NoOrder
				fmt.Printf("Retning opp skal slukke halldown\n")
				buttonLampChan <-  elevio.ButtonLamp{elevMap[config.My_ID].CurrentFloor, elevio.BT_HallDown, false}
			}
			
		case elevStateMap.ED_Down:
			if elevMap[config.My_ID].Orders[elevMap[config.My_ID].CurrentFloor][elevio.BT_HallDown] == elevStateMap.OT_OrderPlaced{
				elevMap[config.My_ID].Orders[elevMap[config.My_ID].CurrentFloor][elevio.BT_HallDown] = elevStateMap.OT_NoOrder
				buttonLampChan <-  elevio.ButtonLamp{elevMap[config.My_ID].CurrentFloor, elevio.BT_HallDown, false}
				fmt.Printf("Retning ned skal slukke halldown\n")
			} else if elevMap[config.My_ID].Orders[elevMap[config.My_ID].CurrentFloor][elevio.BT_HallUp] == elevStateMap.OT_OrderPlaced{
				elevMap[config.My_ID].Orders[elevMap[config.My_ID].CurrentFloor][elevio.BT_HallUp] = elevStateMap.OT_NoOrder
				buttonLampChan <-  elevio.ButtonLamp{elevMap[config.My_ID].CurrentFloor, elevio.BT_HallUp, false}
				fmt.Printf("Retning ned skal slukke hallup\n")
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
	elevStateMap.PrintMap(*elevMap)
	switch elevMap[config.My_ID].CurrentDir{
		case elevStateMap.ED_Up: 
			fmt.Print("Min retning er opp fra før \n")
			if ordersAbove(*elevMap){
				fmt.Printf("Finner ordre over\n")

				for f:= elevMap[config.My_ID].CurrentFloor + 1; f < config.NUM_FLOORS; f++{
					if nearestElevator(*elevMap, f){
						fmt.Printf("nærmeste heis \n")
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
			fmt.Printf("Min retning er ned fra før")
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
 	fmt.Printf("Distanse regnes ut til %v\n", dist)

 	
 	for e := 0; e<config.NUM_ELEVS; e++ {
 		dist_e := int(math.Abs(float64(elevMap[config.My_ID].CurrentFloor - floor)));

 		if elevMap[e].CurrentFloor < floor && elevMap[e].CurrentDir == elevStateMap.ED_Up {
 			
 			if dist_e < dist {
 				fmt.Printf("Jeg er %v etasjer unna, retning UP\n", dist)
 				return false
 			}

 		}else if elevMap[e].CurrentFloor > floor && elevMap[e].CurrentDir == elevStateMap.ED_Down {
 			if dist_e < dist {
 				fmt.Printf("Jeg er %v etasjer unna, retning NED", dist)
 				return false
 			}
 		}
 	}
 	for e := 0; e<config.NUM_ELEVS; e++ {
	 	if elevMap[e].CurrentFloor - floor == dist {
	 			fmt.Printf("vinner på lavest ID\n")
	 			return config.My_ID <= e 

	 		}
	}
 	return true
}
