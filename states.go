package main

import (
	"./dm";
	"reflect";
	"fmt";
)

type State struct {
	Id	int;
	Name	string;
	Abbv	string;
	// *dm.Model;
}

func main() {
	dm.Init("states.db");
	dm.RegisterModel("State", reflect.Typeof(State{}));
	state := dm.Find("State", 12).(State);
	fmt.Printf("FIND BY ID: %s\n", state.Name);

	state = dm.FindFirst("State").(State);
	fmt.Printf("FIRST: %s\n", state.Name);

	state = dm.FindLast("State").(State);
	fmt.Printf("LAST: %s\n", state.Name);

	count := dm.Count("State");
	fmt.Printf("COUNT: %d\n", count);

	println("First 5 states");
	states := dm.FindAll("State", dm.Opts{"limit": 5});
	for s := range states.Results.Iter() {
		state = s.(State);
		fmt.Printf("id: %d  name: %s abbv: %s\n", state.Id, state.Name, state.Abbv);
	}
	println();

	println("States");
	states = dm.FindAll("State");
	for s := range states.Results.Iter() {
		state = s.(State);
		fmt.Printf("id: %d  name: %s abbv: %s\n", state.Id, state.Name, state.Abbv);
	}
}
