package config

import "strconv"

const(
	NUM_ELEVS		= 2
	NUM_FLOORS   	= 4
	NUM_BUTTONS		= 3
	
)

var My_ID int
var My_PORT int 

func Init(id string, port string){
	My_ID, _ = strconv.Atoi(id)
	My_PORT, _ = strconv.Atoi(port)
}


