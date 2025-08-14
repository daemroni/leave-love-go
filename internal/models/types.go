package models

type PlantPreferences struct {
	LightCondition string `json:"lightCondition"` // full-sun | partial-shade | low-light
	CareLevel      string `json:"careLevel"`      // low | medium | high
	PlantType      string `json:"plantType"`      // flowering | foliage | succulent | any
	Location       string `json:"location"`       // indoor | outdoor | both
	Size           string `json:"size"`           // small | medium | large | any
}

type CareInstructions struct {
	Watering    string `json:"watering"`
	Light       string `json:"light"`
	Temperature string `json:"temperature"`
	Humidity    string `json:"humidity"`
}

type Plant struct {
	ID             string           `json:"id"`
	Name           string           `json:"name"`
	ScientificName string           `json:"scientificName"`
	Description    string           `json:"description"`
	Image          string           `json:"image"`
	LightCondition []string         `json:"lightCondition"`
	CareLevel      string           `json:"careLevel"`
	PlantType      string           `json:"plantType"`
	Location       string           `json:"location"`
	Size           string           `json:"size"`
	Features       []string         `json:"features"`
	Care           CareInstructions `json:"careInstructions"`
}
