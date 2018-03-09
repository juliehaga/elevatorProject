package fsm

import(
	"../elevStateMap"
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

const DOOR_TIME 	    = 2
const IDLE_TIME 	    = 4
const MOTOR_DEAD_TIME 	= 5


func Fsm(motorChan chan config.MotorDirection, doorLampChan chan bool, floorChan chan int, buttonLampChan chan config.ButtonLamp, orderChangesChan chan config.ElevStateMap, newOrderChan chan config.ButtonEvent, statusChangesChan chan config.ElevStateMap){
	doorTimer := time.NewTimer(time.Second * DOOR_TIME)
	doorTimer.Stop()

	idleTimer := time.NewTimer(time.Second * IDLE_TIME)
	idleTimer.Stop()

	motorTimer := time.NewTimer(time.Second * MOTOR_DEAD_TIME)
	motorTimer.Stop()



	for{
		//checkIfAcceptOrder(mapChangesChan, buttonLampChan, ackOrderChan)
		select{
		case  floor := <- floorChan:
			fmt.Printf("floor event\n")
			eventNewFloor(motorChan, doorLampChan, doorTimer,orderChangesChan, buttonLampChan, floor, idleTimer, statusChangesChan, motorTimer)
			idleTimer.Reset(time.Second * IDLE_TIME)

		case buttonPushed := <- newOrderChan:
			fmt.Printf("buttonpushed\n")
			eventNewAckOrder(buttonLampChan, motorChan, doorLampChan, doorTimer, orderChangesChan, buttonPushed, idleTimer, motorTimer)
			idleTimer.Reset(time.Second * IDLE_TIME)

		case <- doorTimer.C:
			fmt.Printf("door timeout\n")
			eventDoorTimeout(doorLampChan, statusChangesChan, idleTimer, motorChan, motorTimer)
			idleTimer.Reset(time.Second * IDLE_TIME)
			
		/*case <- idleTimer.C:
			eventIdleTimeout(motorChan, statusChangesChan, orderChangesChan, doorLampChan, doorTimer, buttonLampChan, motorTimer)
			idleTimer.Reset(time.Second * IDLE_TIME)
*/
		case <- motorTimer.C:
			currentMap := elevStateMap.GetLocalMap()
			if (currentMap[config.My_ID].IDLE == false){
				currentMap[config.My_ID].OutOfOrder = true
				statusChangesChan <- currentMap
				idleTimer.Reset(time.Second * MOTOR_DEAD_TIME)
			}
		}
	}
}


func eventIdleTimeout(motorChan chan config.MotorDirection, statusChangesChan chan config.ElevStateMap, orderChangesChan chan config.ElevStateMap, doorLampChan chan bool, doorTimer *time.Timer, buttonLampChan chan config.ButtonLamp,  motorTimer *time.Timer) {
	fmt.Printf("In idle timeout")
	currentMap := elevStateMap.GetLocalMap()
	motorDir := forceChooseDirection(&currentMap, motorTimer)
	if motorDir != config.MD_Stop {
		motorChan <- motorDir
		currentMap[config.My_ID].IDLE = false
		statusChangesChan <- currentMap	
		state = MOVING

	} else if orderInThisFloor(currentMap[config.My_ID].CurrentFloor, currentMap){
		doorLampChan <- true
		doorTimer.Reset(time.Second * DOOR_TIME)

		currentMap[config.My_ID].Door = true
		orderCompleted(&currentMap, buttonLampChan)
		currentMap[config.My_ID].IDLE = false
		orderChangesChan <- currentMap
		state = DOOR_OPEN
	}
}



func eventNewFloor(motorChan chan config.MotorDirection, doorLampChan chan bool, doorTimer *time.Timer, orderChangesChan chan config.ElevStateMap, buttonLampChan chan config.ButtonLamp, floor int, idleTimer *time.Timer, statusChangesChan chan config.ElevStateMap, motorTimer *time.Timer){
	currentMap := elevStateMap.GetLocalMap()
	if floor != -1 {
		currentMap[config.My_ID].CurrentFloor = floor
		statusChangesChan <- currentMap
	}

	switch(state){
		case MOVING:
			motorTimer.Reset(time.Second * MOTOR_DEAD_TIME)
			if shouldStop(currentMap) {
				motorChan <- config.MD_Stop
					if  orderInThisFloor(currentMap[config.My_ID].CurrentFloor, currentMap){
						doorLampChan <- true
						doorTimer.Reset(time.Second * DOOR_TIME)

						currentMap[config.My_ID].Door = true
						orderCompleted(&currentMap, buttonLampChan)
						currentMap[config.My_ID].IDLE = false
						orderChangesChan <- currentMap
						state = DOOR_OPEN
					}	
			}
	}
}

func eventDoorTimeout(doorLampChan chan bool, statusChangesChan chan config.ElevStateMap, idleTimer *time.Timer, motorChan chan config.MotorDirection, motorTimer *time.Timer){
	currentMap := elevStateMap.GetLocalMap()
	switch(state){
		case DOOR_OPEN:
			doorLampChan <- false
			currentMap[config.My_ID].Door = false
			

			motorDir := chooseDirection(&currentMap, motorTimer)
			
			if motorDir != config.MD_Stop {
				motorChan <- motorDir
				currentMap[config.My_ID].IDLE = false
				state = MOVING
			} else {
				currentMap[config.My_ID].IDLE = true
				state = IDLE
			}

			statusChangesChan <- currentMap	
		
	}

	
}


func eventNewAckOrder(buttonLampChan chan config.ButtonLamp, motorChan chan config.MotorDirection, doorLampChan chan bool, doorTimer *time.Timer, orderChangesChan chan config.ElevStateMap, buttonPushed config.ButtonEvent, idleTimer *time.Timer, motorTimer *time.Timer){

	currentMap := elevStateMap.GetLocalMap()
	fmt.Printf("\n \n CURRENT FLOOR %v \n \n", currentMap[config.My_ID].CurrentFloor)
	buttonLampChan <- config.ButtonLamp{buttonPushed.Floor, buttonPushed.Button, true}

	if buttonPushed.Button != config.BT_Cab{
		for elev := 0; elev < config.NUM_ELEVS; elev++{		
			if currentMap[elev].Connected == true{		
				currentMap[elev].Orders[buttonPushed.Floor][buttonPushed.Button] = config.OT_OrderPlaced
			}
		}	
	}else {
		currentMap[config.My_ID].Orders[buttonPushed.Floor][buttonPushed.Button] = config.OT_OrderPlaced
	}	

	switch(state){
		case IDLE:

			if shouldStop(currentMap) && orderInThisFloor(currentMap[config.My_ID].CurrentFloor, currentMap){
				doorLampChan <- true	
				currentMap[config.My_ID].Door = true
				orderCompleted(&currentMap, buttonLampChan)
				doorTimer.Reset(time.Second * DOOR_TIME)
				currentMap[config.My_ID].IDLE = false
				state = DOOR_OPEN

				
			}else{
				motorDir := chooseDirection(&currentMap, motorTimer)
				if motorDir != config.MD_Stop {
					motorChan <- motorDir
					currentMap[config.My_ID].IDLE = false
					
					state = MOVING
				}

				
			}	

	}
	orderChangesChan <- currentMap



	//elevStateMap.PrintMap(currentMap)
}

func shouldStop(elevMap config.ElevStateMap) bool{
	switch(state){
		case MOVING: 
			if elevMap[config.My_ID].CurrentFloor == config.NUM_FLOORS-1 && elevMap[config.My_ID].CurrentDir == config.ED_Up{
				return true
			} else if elevMap[config.My_ID].CurrentFloor == 0 && elevMap[config.My_ID].CurrentDir == config.ED_Down {
				return true
			}
	}
	
	//elevStateMap.PrintMap(elevMap)

	if elevMap[config.My_ID].Orders[elevMap[config.My_ID].CurrentFloor][config.BT_Cab]==config.OT_OrderPlaced{
		return true
	}

	if !orderInThisFloor(elevMap[config.My_ID].CurrentFloor, elevMap){
		return false
	}

	switch elevMap[config.My_ID].CurrentDir{
		case config.ED_Up:
			//order on current floor and no orders above
			if elevMap[config.My_ID].Orders[elevMap[config.My_ID].CurrentFloor][config.BT_HallUp]==config.OT_OrderPlaced{
				return true
			} else if !ordersAbove(elevMap) && elevMap[config.My_ID].Orders[elevMap[config.My_ID].CurrentFloor][config.BT_HallDown]==config.OT_OrderPlaced{
				return true }

		case config.ED_Down:
			//order on current floor and no orders below
		 	if elevMap[config.My_ID].Orders[elevMap[config.My_ID].CurrentFloor][config.BT_HallDown]==config.OT_OrderPlaced{
		 		return true
		 	} else if !ordersBelow(elevMap) && elevMap[config.My_ID].Orders[elevMap[config.My_ID].CurrentFloor][config.BT_HallUp]==config.OT_OrderPlaced {
			 	return true 
			}
	}
	return false
}

func ordersAbove(elevMap config.ElevStateMap) bool{
	for f := elevMap[config.My_ID].CurrentFloor + 1; f<config.NUM_FLOORS; f++{
		for b := config.BT_HallUp; b<= config.BT_Cab; b++{ 
			if elevMap[config.My_ID].Orders[f][b] == config.OT_OrderPlaced{


				return true
			}
		}
	}

	return false
}

func ordersBelow(elevMap config.ElevStateMap) bool{
	for f := elevMap[config.My_ID].CurrentFloor - 1; f>=0; f--{
		for b := config.BT_HallUp; b<= config.BT_Cab; b++{ 
			if elevMap[config.My_ID].Orders[f][b] == config.OT_OrderPlaced {
				return true
			}
		}
	}
	return false
}




func orderCompleted(elevMap *config.ElevStateMap, buttonLampChan chan config.ButtonLamp){

	if elevMap[config.My_ID].Orders[elevMap[config.My_ID].CurrentFloor][config.BT_Cab] == config.OT_OrderPlaced{
		elevMap[config.My_ID].Orders[elevMap[config.My_ID].CurrentFloor][config.BT_Cab] = config.OT_NoOrder
		buttonLampChan <- config.ButtonLamp{elevMap[config.My_ID].CurrentFloor, config.BT_Cab, false}
	}

	switch(elevMap[config.My_ID].CurrentDir){
		case config.ED_Up: 
			if elevMap[config.My_ID].Orders[elevMap[config.My_ID].CurrentFloor][config.BT_HallUp] == config.OT_OrderPlaced{
				
				for elev := 0; elev < config.NUM_ELEVS; elev++{	
					if elevMap[elev].Connected == true{				
						elevMap[elev].Orders[elevMap[config.My_ID].CurrentFloor][config.BT_HallUp] = config.OT_NoOrder
					}
				}	
				buttonLampChan <-  config.ButtonLamp{elevMap[config.My_ID].CurrentFloor, config.BT_HallUp, false}

			} else if elevMap[config.My_ID].Orders[elevMap[config.My_ID].CurrentFloor][config.BT_HallDown] == config.OT_OrderPlaced{

				for elev := 0; elev < config.NUM_ELEVS; elev++{			
					if elevMap[elev].Connected == true{		
						elevMap[elev].Orders[elevMap[config.My_ID].CurrentFloor][config.BT_HallDown] = config.OT_NoOrder
					}
				}
				
				buttonLampChan <-  config.ButtonLamp{elevMap[config.My_ID].CurrentFloor, config.BT_HallDown, false}
			}
			
		case config.ED_Down:
			if elevMap[config.My_ID].Orders[elevMap[config.My_ID].CurrentFloor][config.BT_HallDown] == config.OT_OrderPlaced{

				for elev := 0; elev < config.NUM_ELEVS; elev++{		
					if elevMap[elev].Connected == true{			
						elevMap[elev].Orders[elevMap[config.My_ID].CurrentFloor][config.BT_HallDown] = config.OT_NoOrder
					}
				}


				buttonLampChan <-  config.ButtonLamp{elevMap[config.My_ID].CurrentFloor, config.BT_HallDown, false}
			} else if elevMap[config.My_ID].Orders[elevMap[config.My_ID].CurrentFloor][config.BT_HallUp] == config.OT_OrderPlaced{

				for elev := 0; elev < config.NUM_ELEVS; elev++{	
					if elevMap[elev].Connected == true{				
						elevMap[elev].Orders[elevMap[config.My_ID].CurrentFloor][config.BT_HallUp] = config.OT_NoOrder
					}
				}					
				buttonLampChan <-  config.ButtonLamp{elevMap[config.My_ID].CurrentFloor, config.BT_HallUp, false}


			}
			
	}
}



func orderInThisFloor( floor int, elevMap config.ElevStateMap) bool{
	for b := config.BT_HallUp; b <= config.BT_Cab; b++ {
		if elevMap[config.My_ID].Orders[floor][b] == config.OT_OrderPlaced {
			return true
		
		}
	}
	return false
}


func chooseDirection(elevMap *config.ElevStateMap, motorTimer *time.Timer) config.MotorDirection{
	motorTimer.Reset(time.Second * MOTOR_DEAD_TIME)
	switch elevMap[config.My_ID].CurrentDir{
		case config.ED_Up: 
			if ordersAbove(*elevMap){
				for f:= elevMap[config.My_ID].CurrentFloor + 1; f < config.NUM_FLOORS; f++{
					if  orderInThisFloor(f, *elevMap) && (nearestElevator(*elevMap, f) || elevMap[config.My_ID].Orders[f][config.BT_Cab] == config.OT_OrderPlaced){
						elevMap[config.My_ID].CurrentDir = config.ED_Up
						fmt.Printf("velger opp\n")
						return config.MD_Up
					}
				}
				
			} else if ordersBelow(*elevMap){
				for f:= elevMap[config.My_ID].CurrentFloor - 1; f >= 0; f--{
					if orderInThisFloor(f, *elevMap) && (nearestElevator(*elevMap, f) || elevMap[config.My_ID].Orders[f][config.BT_Cab] == config.OT_OrderPlaced){
						elevMap[config.My_ID].CurrentDir = config.ED_Down
						fmt.Printf("Velger ned \n")
						return config.MD_Down
					}
				}
			} else {
				return config.MD_Stop
			}
		case config.ED_Down:
			if ordersBelow(*elevMap){
				for f:= elevMap[config.My_ID].CurrentFloor - 1; f >= 0; f--{
					if orderInThisFloor(f, *elevMap) && (nearestElevator(*elevMap, f) || elevMap[config.My_ID].Orders[f][config.BT_Cab] == config.OT_OrderPlaced){
						elevMap[config.My_ID].CurrentDir = config.ED_Down
						fmt.Printf("Velger ned \n")
						return config.MD_Down
					}
				}

			} else if ordersAbove(*elevMap){
				for f:= elevMap[config.My_ID].CurrentFloor + 1; f < config.NUM_FLOORS; f++{
					if orderInThisFloor(f, *elevMap) && (nearestElevator(*elevMap, f) || elevMap[config.My_ID].Orders[f][config.BT_Cab] == config.OT_OrderPlaced){
						elevMap[config.My_ID].CurrentDir = config.ED_Up
						fmt.Printf("velger opp\n")
						return config.MD_Up
					}
				}
			} else {
				return config.MD_Stop
			}
	}
	fmt.Printf("STOP\n")
	return config.MD_Stop
}



func nearestElevator(elevMap config.ElevStateMap, floor int) bool{

 	myDist := int(math.Abs(float64(elevMap[config.My_ID].CurrentFloor - floor)))



 	if elevMap[config.My_ID].CurrentFloor < floor { 
	 	for e := 0; e<config.NUM_ELEVS; e++ {
	 		if elevMap[e].Connected == true{	
			 	if e != config.My_ID{	
			 		distElev := int(math.Abs(float64(elevMap[e].CurrentFloor - floor)))
			 		if distElev < myDist{
			 			if elevMap[e].CurrentFloor <= floor && (elevMap[e].CurrentDir == config.ED_Up || elevMap[e].IDLE ){
			 				return false
			 			} else if elevMap[e].CurrentFloor >= floor && (elevMap[e].CurrentDir == config.ED_Down || elevMap[e].IDLE ) {
			 				return false
			 			}
			 		} else if myDist == distElev && (elevMap[e].CurrentDir == config.ED_Up || elevMap[e].IDLE){
			 			if e > config.My_ID{
			 				return false
			 			}
			 		}
			 	}
			} 	
		 }
 	} else if elevMap[config.My_ID].CurrentFloor > floor {
		 	for e := 0; e<config.NUM_ELEVS; e++ {
		 		if elevMap[e].Connected == true{	
				 	if e != config.My_ID{	
				 		distElev := int(math.Abs(float64(elevMap[e].CurrentFloor - floor)))
				 		if distElev < myDist{
				 			if elevMap[e].CurrentFloor >= floor && (elevMap[e].CurrentDir ==config.ED_Down || elevMap[e].IDLE ){
				 				return false
				 			} else if elevMap[e].CurrentFloor <= floor && (elevMap[e].CurrentDir == config.ED_Up || elevMap[e].IDLE ) {
			 					return false
			 				}
				 		}  else if myDist == distElev && (elevMap[e].CurrentDir == config.ED_Down || elevMap[e].IDLE){
			 			if e > config.My_ID{
			 				return false
			 			}
			 		}
			 	}
			}
		}	
	}
	fmt.Printf("Jeg er nærmest\n")
 	return true
}



func forceChooseDirection(elevMap *config.ElevStateMap, motorTimer *time.Timer) config.MotorDirection{
	motorTimer.Reset(time.Second * MOTOR_DEAD_TIME)
	if orderInThisFloor(elevMap[config.My_ID].CurrentFloor, *elevMap){
		return config.MD_Stop
	
	} else if elevMap[config.My_ID].CurrentFloor >= config.NUM_FLOORS/2 {
		if ordersAbove(*elevMap){
			elevMap[config.My_ID].CurrentDir = config.ED_Up
			return config.MD_Up
		} else if ordersBelow(*elevMap){
			elevMap[config.My_ID].CurrentDir = config.ED_Down
			return config.MD_Down
		}
	} else if elevMap[config.My_ID].CurrentFloor < config.NUM_FLOORS/2{
		if ordersBelow(*elevMap) {
			elevMap[config.My_ID].CurrentDir = config.ED_Down
			return config.MD_Down
		} else if ordersAbove(*elevMap){
			elevMap[config.My_ID].CurrentDir = config.ED_Up
			return config.MD_Up
		}
	}
	fmt.Printf("Burde stoppe\n")
	return config.MD_Stop
} 
