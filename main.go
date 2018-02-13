package main

import "./elevio"
import "fmt"

func main(){

	fmt.Println("Running main")

    numFloors := 4

    elevio.Init("localhost:15657", numFloors)
	elevio.ClearAllButtonLamps();
    
    var d elevio.MotorDirection = elevio.MD_Up

    elevio.SetMotorDirection(d)
	elevio.SetFloorIndicator(3)
    //elevio.SetFloorIndicator(2)
	//elevio.SetFloorIndicator(3)
	//elevio.SetFloorIndicator(1)
    drv_buttons := make(chan elevio.ButtonEvent)
    drv_floors  := make(chan int)  
    
    go elevio.PollButtons(drv_buttons)
    go elevio.PollFloorSensor(drv_floors)
	//go elevio.SetFloorLamp()
    //go elevio.PollObstructionSwitch(drv_obstr)
    //go elevio.PollStopButton(drv_stop)
    
    
    for {
        select {
        case a := <- drv_buttons:
            fmt.Printf("%+v\n", a)
            elevio.SetButtonLamp(a.Button, a.Floor, true)
            
        case a := <- drv_floors:
            fmt.Printf("%+v\n", a)
            if a == numFloors-1 {
                d = elevio.MD_Down
            } else if a == 0 {
                d = elevio.MD_Up
            }
            elevio.SetMotorDirection(d)
            
     /*       
        case a := <- drv_obstr:
            fmt.Printf("%+v\n", a)
            if a {
                elevio.SetMotorDirection(elevio.MD_Stop)
            } else {
                elevio.SetMotorDirection(d)
            }
            
        case a := <- drv_stop:
            fmt.Printf("%+v\n", a)
            for f := 0; f < numFloors; f++ {
                for b := elevio.ButtonType(0); b < 3; b++ {
                    elevio.SetButtonLamp(b, f, false)
                }
            }*/
        }
    }    
}
