TTK 4145 - REAL TIME PROGRAMMING

This project is a elevator system consisting of n elevators serving orders in m floors. The purpose is to synchronize all orders and orderlights between the elevators and make sure that the best suited elevator serves the order. Multiple elevators should be more efficient than one. 

Our solution is a peer-to-peer network of elevators where all peers has access to the same information stored in a struct named ElevStateMap. The map keeps track of the direction, current floor and orders in all elevators. If the state is updated in one elevator it sends the new map using UDP messages. 

We used the network module handed out by the faculty. The same applies to the hardware functions in the elevio module.


The system consist of the following modules: 

config:
Keeps all data types and constants. 

elevio:
Reads and writes to hardware. 

elevStateMap: 
Keeps track of the status of all elevators. Updates the local map from changes made locally and of other elevators. 

Makes sure that all elevators has received the order before placing it. 


fsm: 
Controls the behaviour of the elevator. 

network: 
Keeps track of all connected peers and the communication between them. 
