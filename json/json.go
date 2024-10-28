package main

import (
	"encoding/json"
	"fmt"
	"log"
)

type Person struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	HairColor string `json:"hair_color"`
	HasDog    bool   `json:"has_dog"`
}

func jsonTest() {
	myJson := `
[
    {
        "first_name": "Clark",
        "last_name": "Kent",
        "hair_color": "black",
        "has_dog": true
    },
    {
        "first_name": "Bruce",
        "last_name": "Wayne",
        "hair_color": "black",
        "has_dog": false
    }
]`

	var unmarshelled []Person

	err := json.Unmarshal([]byte(myJson), &unmarshelled)
	if err != nil {
		log.Println("Error")
	}

	log.Printf("unmarshelled: %v", unmarshelled)

	var mySlice []Person

	m1 := Person{
		FirstName: "Wally",
		LastName:  "West",
		HairColor: "red",
		HasDog:    false,
	}

	mySlice = append(mySlice, m1)

	m2 := Person{
		FirstName: "Diana",
		LastName:  "West",
		HairColor: "black",
		HasDog:    true,
	}

	mySlice = append(mySlice, m2)

	newJson, err := json.MarshalIndent(mySlice, "", "   ")
	if err != nil {
		log.Println("Error")
	}

	fmt.Println(string(newJson))
}
