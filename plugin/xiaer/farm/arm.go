package farm

import "encoding/json"

var (
	armList []Arm
	armMap  = map[int]Arm{}
)

func init() {
	json.Unmarshal([]byte(armJson), &armList)
	for index := 0; index < len(armList); index++ {
		arm := armList[index]
		armMap[arm.Level] = arm
	}
}

type Arm struct {
	Level       int    `json:"level"`
	Name        string `json:"name"`
	Price       int    `json:"price"`
	Probability int    `json:"probability"`
	Noise       int    `json:"noise"`
}

const armJson = `
[
{
"level": 1,
"name": "P1911",
"price": 1000,
"probability": 30,
"noise": 50
},{
"level": 2,
"name": "沙漠之鹰",
"price": 1500,
"probability": 35,
"noise": 60
},{
"level": 3,
"name": "UZI",
"price": 1500,
"probability": 30,
"noise": 45
},{
"level": 4,
"name": "M416",
"price": 2000,
"probability": 40,
"noise": 50
},{
"level": 5,
"name": "AK47",
"price": 3000,
"probability": 50,
"noise": 55
},{
"level": 6,
"name": "98k",
"price": 4000,
"probability": 60,
"noise": 40
},{
"level": 7,
"name": "AWM",
"price": 5000,
"probability": 70,
"noise": 30
},{
"level": 8,
"name": "弩",
"price": 8000,
"probability": 80,
"noise": 20
},{
"level": 9,
"name": "麻醉针",
"price": 10000,
"probability": 80,
"noise": 10
}
]
`
