TTK 4145 - REAL TIME PROGRAMMING

This project is a elevator system consisting of n elevators serving orders in m floors. The purpose is to synchronize all orders and orderlights between the elevators and make sure that the best suited elevator serves the order. Multiple elevators should be more efficient than one. 

Our solution is a peer-to-peer network of elevators where all peers has access to the same information stored in a struct named ElevStateMap. The map keeps track of the direction, current floor and orders in all elevators. If the state is updated in one elevator it sends the new map using UDP messages. 

We used the network module handed out by the faculty. 