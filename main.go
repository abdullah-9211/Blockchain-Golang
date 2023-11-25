package main

func main() {
	scenario_number := 0
	switch scenario_number {
	case 0:
		scenario_connection_control()
	case 1:
		scenario_blockchain_observation()
	case 2:
		scenario_blockchain_bad_node()
	}
}
