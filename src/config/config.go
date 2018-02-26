package config

import "strconv"

const(
	NUM_ELEVS		= 3
	NUM_FLOORS   	= 4
	NUM_BUTTONS		= 3
	
)

var My_ID int 

func InitConfig(id string){
	My_ID, _ = strconv.Atoi(id)
}
