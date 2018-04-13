package config

import "strconv"

const(
	NUM_ELEVS		= 3
	NUM_FLOORS   	= 4
	NUM_BUTTONS		= 3
	
)

var My_ID int
var My_PORT int 

func Init(id string, port string){
	My_ID, _ = strconv.Atoi(id)
	My_PORT, _ = strconv.Atoi(port)
}



const (
	MD_Up   MotorDirection = 1
	MD_Down                = -1
	MD_Stop                = 0
)

type ButtonType int

const (
	BT_HallUp   ButtonType = 0
	BT_HallDown            = 1
	BT_Cab                 = 2
)

type ButtonEvent struct {
	Floor  int
	Button ButtonType
}

type ButtonLamp struct {
	Floor  int
	Button ButtonType
	Set bool
}

type OrderType int

const (
	OT_NoOrder                = 0
	OT_OrderPlaced            = 1
)

type ElevDir int

const (
	ED_Up   		ElevDir = 1
	ED_Down                 = -1
)


type ElevInfo struct{
	CurrentFloor int
	CurrentDir ElevDir
	Connected bool
	Door bool
	IDLE bool
	OutOfOrder bool 
	Orders [NUM_FLOORS][NUM_BUTTONS] OrderType
}


type ElevStateMap [NUM_ELEVS]ElevInfo

type Message struct {
	ID int
	MsgType int
	ElevMap   ElevStateMap
	Reciever_ID int 
	Button ButtonEvent
}


type StatusMsg struct {
	ID int
	CurrentFloor int
	CurrentDir ElevDir
	Door bool
	OutOfOrder bool
	IDLE bool
}

type OrderMsg struct{
	ID int
	ElevMap ElevStateMap
}

type ActiveOrderMsg struct{
	ID int
	Order ButtonEvent
}

type OrderCompleteMsg struct{
	button ButtonType
}

const (
	ElevStatus     = 0
	Orders         = 1
	Ack 	       = 2
	ActiveOrder    = 3
)

type PeerUpdate struct {
	Peers []string
	New   string
	Lost  []string
}

type ActiveOrders struct {
	Button ButtonEvent
	ID int
	ActiveOrder bool
	ElevMap ElevStateMap
}

type NewButtons struct {
	PushesMatrix [NUM_FLOORS][NUM_BUTTONS]bool
	ElevMap ElevStateMap
}

type MotorDirection int