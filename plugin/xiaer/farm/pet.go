package farm

import "encoding/json"

var (
	petList []Pet
	petMap  = map[int]Pet{}
)

func init() {
	json.Unmarshal([]byte(petJson), &petList)
	for index := 0; index < len(petList); index++ {
		pet := petList[index]
		petMap[pet.Level] = pet
	}
}

type Pet struct {
	Level       int    `json:"level"`
	Name        string `json:"name"`
	Price       int    `json:"price"`
	Probability int    `json:"probability"`
}

const petJson = `
[
{
"level": 1,
"name": "斗牛犬",
"price": 1000,
"probability": 30
},{
"level": 2,
"name": "哈士奇",
"price": 1500,
"probability": 35
},{
"level": 3,
"name": "柴犬",
"price": 1500,
"probability": 30
},{
"level": 4,
"name": "萨摩耶",
"price": 2000,
"probability": 40
},{
"level": 5,
"name": "阿拉斯加",
"price": 3000,
"probability": 50
},{
"level": 6,
"name": "牧羊犬",
"price": 4000,
"probability": 60
},{
"level": 7,
"name": "藏獒",
"price": 5000,
"probability": 70
},{
"level": 8,
"name": "霸王龙",
"price": 8000,
"probability": 80
},{
"level": 9,
"name": "奥特曼",
"price": 10000,
"probability": 80
}
]
`
