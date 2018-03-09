package elevio


import( 
	"../config"
	"../elevStateMap"
	"time"
	"sync"
	"net"
	"fmt"
)



const _pollRate = 20 * time.Millisecond

var _initialized bool = false
var _numFloors int = 4
var _mtx sync.Mutex
var _conn net.Conn





func Elevio(motorChan chan config.MotorDirection, doorLampChan chan bool, newOrderChan chan config.ButtonEvent, floorChan chan int, buttonLampChan chan config.ButtonLamp) {
	go PollButtons(newOrderChan)
    go PollFloorSensor(floorChan)
    //update map?

	for {
		select {
		case dir := <- motorChan:
			SetMotorDirection(dir)
		case light := <-doorLampChan:
			SetDoorOpenLamp(light)
		case lamp := <- buttonLampChan:
			fmt.Printf("Slukker lys %v", lamp)
			SetButtonLamp(lamp)

		}
	


	}
	
}



func Init(addr string, numFloors int) {
	if _initialized {
		fmt.Println("Driver already initialized!")
		return
	}
	_numFloors = numFloors
	_mtx = sync.Mutex{}
	var err error
	_conn, err = net.Dial("tcp", addr)
	if err != nil {
		panic(err.Error())
	}
	_initialized = true
	SetMotorDirection(config.MD_Up)
	for GetFloor() == -1{}
	SetMotorDirection(config.MD_Stop)



	currentMap := elevStateMap.GetLocalMap()
	for f := 0; f < config.NUM_FLOORS; f++{
		for b:= config.BT_HallUp; b <= config.BT_Cab; b++{
			if currentMap[config.My_ID].Orders[f][b] == config.OT_OrderPlaced{
				lamp := config.ButtonLamp{f, b, true}	
				SetButtonLamp(lamp)
			} else {
				lamp := config.ButtonLamp{f, b, false}	
				SetButtonLamp(lamp)
			}
		}
	}
	
	SetDoorOpenLamp(false)
}


func ClearAllButtonLamps(){
	for f:= 0; f < _numFloors; f++ {
		for b:= config.ButtonType(0); b < 3; b++ {
			SetButtonLamp(config.ButtonLamp{f, b, false})	
		}
	}
}


func SetMotorDirection(dir config.MotorDirection) {
	_mtx.Lock()
	defer _mtx.Unlock()
	_conn.Write([]byte{1, byte(dir), 0, 0})
}

func SetButtonLamp(lamp config.ButtonLamp) {
	_mtx.Lock()
	defer _mtx.Unlock()
	_conn.Write([]byte{2, byte(lamp.Button), byte(lamp.Floor), toByte(lamp.Set)})
}

func SetFloorIndicator(floor int) {
	_mtx.Lock()
	defer _mtx.Unlock()
	_conn.Write([]byte{3, byte(floor), 0, 0})
}

func SetDoorOpenLamp(value bool) {
	_mtx.Lock()
	defer _mtx.Unlock()
	_conn.Write([]byte{4, toByte(value), 0, 0})
}

func SetStopLamp(value bool) {
	_mtx.Lock()
	defer _mtx.Unlock()
	_conn.Write([]byte{5, toByte(value), 0, 0})
}



func PollButtons(receiver chan<- config.ButtonEvent) {
	prev := make([][3]bool, _numFloors)
	for {
		time.Sleep(_pollRate)
		for f := 0; f < _numFloors; f++ {
			for b := config.ButtonType(0); b < 3; b++ {
				v := getButton(b, f)
				if v != prev[f][b] && v != false {
					receiver <- config.ButtonEvent{f, config.ButtonType(b)}
				}
				prev[f][b] = v
			}
		}
	}
}

func PollFloorSensor(receiver chan<- int) {	
	prev := -1
	for {
		time.Sleep(_pollRate)
		v := GetFloor()
		if v < _numFloors && v >= 0 {
			SetFloorIndicator(v)

		}


		if v != prev && v != -1 {
			receiver <- v
		}
		prev = v
	}
}


func PollStopButton(receiver chan<- bool) {
	prev := false
	for {
		time.Sleep(_pollRate)
		v := getStop()
		if v != prev {
			receiver <- v
		}
		prev = v
	}
}

func PollObstructionSwitch(receiver chan<- bool) {
	prev := false
	for {
		time.Sleep(_pollRate)
		v := getObstruction()
		if v != prev {
			receiver <- v
		}
		prev = v
	}
}



func getButton(button config.ButtonType, floor int) bool {
	_mtx.Lock()
	defer _mtx.Unlock()
	_conn.Write([]byte{6, byte(button), byte(floor), 0})
	var buf [4]byte
	_conn.Read(buf[:])
	return toBool(buf[1])
}


func GetFloor() int {
	_mtx.Lock()
	defer _mtx.Unlock()
	_conn.Write([]byte{7, 0, 0, 0})
	var buf [4]byte
	_conn.Read(buf[:])
	if buf[1] != 0 {
		return int(buf[2])
	} else {
		return -1
	}
}


func getStop() bool {
	_mtx.Lock()
	defer _mtx.Unlock()
	_conn.Write([]byte{8, 0, 0, 0})
	var buf [4]byte
	_conn.Read(buf[:])
	return toBool(buf[1])
}

func getObstruction() bool {
	_mtx.Lock()
	defer _mtx.Unlock()
	_conn.Write([]byte{9, 0, 0, 0})
	var buf [4]byte
	_conn.Read(buf[:])
	return toBool(buf[1])
}

func toByte(a bool) byte {
	var b byte = 0
	if a {
		b = 1
	}
	return b
}

func toBool(a byte) bool {
	var b bool = false
	if a != 0 {
		b = true
	}
	return b
}
