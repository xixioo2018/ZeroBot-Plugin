package farm

import (
	"context"
	log "github.com/sirupsen/logrus"
	"math"
	"time"

	"github.com/FloatTech/ZeroBot-Plugin/database/mongo"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// 用户库存

func stock(groupCode int64, uin int64) Stock {
	var stock Stock
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	stockCollection := mongo.Collection("game.farm.stock")
	cur, _ := stockCollection.Find(ctx, bson.M{"groupCode": groupCode, "uin": uin})
	defer cur.Close(ctx)
	if cur.Next(ctx) {
		err := cur.Decode(&stock)
		if err != nil {
			log.Error(err)
		}
	} else {
		stock.GroupCode = groupCode
		stock.Uin = uin
		stock.CropCount = map[string]int{}
		stockUpdate(stock)
	}
	return stock
}

func stockUpdate(stock Stock) {
	cropCount := bson.M{}
	for k, v := range stock.CropCount {
		cropCount[k] = v
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	user := bson.M{"uin": stock.Uin, "groupCode": stock.GroupCode}
	update := bson.M{"$set": bson.M{"cropCount": cropCount}}
	stockCollection := mongo.Collection("game.farm.stock")
	_, err := stockCollection.UpdateOne(ctx, user, update, options.Update().SetUpsert(true))
	if err != nil {
		log.Error(err)
	}
}

type Stock struct {
	Uin       int64
	GroupCode int64
	CropCount map[string]int
}

// 用户资产

func assets(groupCode int64, uin int64) Assets {
	var assets Assets
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	assetsCollection := mongo.Collection("game.farm.assets")
	cur, _ := assetsCollection.Find(ctx, bson.M{"groupCode": groupCode, "uin": uin})
	defer cur.Close(ctx)
	if cur.Next(ctx) {
		err := cur.Decode(&assets)
		if err != nil {
			log.Error(err)
		}
	} else {
		assets.GroupCode = groupCode
		assets.Uin = uin
		assets.Exp = 0
		assets.Coins = 3000
		assets.Fields = 1
		assets.Ponds = 1
		assetsCollection.InsertOne(ctx, bson.M{
			"uin":       assets.Uin,
			"groupCode": assets.GroupCode,
			"exp":       assets.Exp,
			"coins":     assets.Coins,
			"fields":    assets.Fields,
			"ponds":     assets.Ponds,
		})
	}
	return assets
}

func assetsGroup(groupCode int64) []Assets {
	groupAssets := make([]Assets, 0)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	assetsCollection := mongo.Collection("game.farm.assets")

	findOptions := options.Find()
	// Sort by `price` field descending
	findOptions.SetSort(bson.D{{"exp", -1}})
	cur, _ := assetsCollection.Find(ctx, bson.D{{"groupCode", groupCode}}, findOptions)
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		var asset Assets
		err := cur.Decode(&asset)
		if err != nil {
			log.Error(err)
		}
		groupAssets = append(groupAssets, asset)
	}
	return groupAssets
}

func assetsCoinsInc(code int64, uin int64, inc int64) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	assetsCollection := mongo.Collection("game.farm.assets")
	_, err := assetsCollection.UpdateOne(ctx, bson.M{"uin": uin, "groupCode": code}, bson.M{"$inc": bson.M{"coins": inc}})
	if err != nil {
		log.Error(err)
	}
}

func assetsFieldInc(code int64, uin int64, inc int) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	assetsCollection := mongo.Collection("game.farm.assets")
	_, err := assetsCollection.UpdateOne(ctx, bson.M{"uin": uin, "groupCode": code}, bson.M{"$inc": bson.M{"fields": inc}})
	if err != nil {
		log.Error(err)
	}
}

func assetsExpInc(code int64, uin int64, up int64) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	assetsCollection := mongo.Collection("game.farm.assets")
	_, err := assetsCollection.UpdateOne(ctx, bson.M{"uin": uin, "groupCode": code}, bson.M{"$inc": bson.M{"exp": up}})
	if err != nil {
		log.Error(err)
	}
}

type Assets struct {
	GroupCode int64
	Uin       int64
	Exp       int64
	Coins     int64
	Fields    int
	Ponds     int
}

// 场地

func land(groupCode int64, uin int64) Land {
	var land Land
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	landCollection := mongo.Collection("game.farm.land")
	cur, _ := landCollection.Find(ctx, bson.M{"groupCode": groupCode, "uin": uin})
	defer cur.Close(ctx)
	if cur.Next(ctx) {
		err := cur.Decode(&land)
		if err != nil {
			log.Error(err)
		}
	} else {
		land.GroupCode = groupCode
		land.Uin = uin
		land.Fields = map[string]Field{}
		landUpdate(land)
	}
	return land
}

func landUpdate(land Land) {
	fields := bson.M{}
	for k, v := range land.Fields {
		watered := bson.M{}
		for state, uin := range v.Watered {
			watered[state] = uin
		}
		fields[k] = bson.M{
			"level":     v.Level,
			"plantTime": v.PlantTime,
			"watered":   watered,
			"stealer":   v.Stealer,
			"alerted":   v.Alerted,
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	user := bson.M{"groupCode": land.GroupCode, "uin": land.Uin}
	update := bson.M{"$set": bson.M{"fields": fields}}
	landCollection := mongo.Collection("game.farm.land")
	_, err := landCollection.UpdateOne(ctx, user, update, options.Update().SetUpsert(true))
	if err != nil {
		log.Error(err)
	}
}

type Field struct {
	Level     int              // 种的什么果实
	PlantTime int64            // 种植的时间
	Watered   map[string]int64 // 浇水
	Stealer   []int64          // 偷菜的人
	Alerted   []int64          // 被狗咬的人
}

type Land struct {
	GroupCode int64
	Uin       int64
	Fields    map[string]Field
}

// pets

func pets(groupCode int64, uin int64) Pets {
	var pets Pets
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	landCollection := mongo.Collection("game.farm.pets")
	cur, _ := landCollection.Find(ctx, bson.M{"groupCode": groupCode, "uin": uin})
	defer cur.Close(ctx)
	if cur.Next(ctx) {
		err := cur.Decode(&pets)
		if err != nil {
			log.Error(err)
		}
	} else {
		pets.GroupCode = groupCode
		pets.Uin = uin
		pets.Pets = []int{}
		petsUpdate(pets)
	}
	return pets
}

// arms

func arms(groupCode int64, uin int64) Arms {
	var arms Arms
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	landCollection := mongo.Collection("game.farm.arms")
	cur, _ := landCollection.Find(ctx, bson.M{"groupCode": groupCode, "uin": uin})
	defer cur.Close(ctx)
	if cur.Next(ctx) {
		err := cur.Decode(&arms)
		if err != nil {
			log.Error(err)
		}
	} else {
		arms.GroupCode = groupCode
		arms.Uin = uin
		arms.Arms = []int{}
		armsUpdate(arms)
	}
	return arms
}

func petsUpdate(pets Pets) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	user := bson.M{"groupCode": pets.GroupCode, "uin": pets.Uin}
	update := bson.M{"$set": bson.M{"pets": pets.Pets}}
	landCollection := mongo.Collection("game.farm.pets")
	_, err := landCollection.UpdateOne(ctx, user, update, options.Update().SetUpsert(true))
	if err != nil {
		log.Error(err)
	}
}

func armsUpdate(arms Arms) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	user := bson.M{"groupCode": arms.GroupCode, "uin": arms.Uin}
	update := bson.M{"$set": bson.M{"arms": arms.Arms}}
	landCollection := mongo.Collection("game.farm.arms")
	_, err := landCollection.UpdateOne(ctx, user, update, options.Update().SetUpsert(true))
	if err != nil {
		log.Error(err)
	}
}

type Pets struct {
	GroupCode int64
	Uin       int64
	Pets      []int
}

type Arms struct {
	GroupCode int64
	Uin       int64
	Arms      []int
}

// 等级/金额计算公式

func level(exp int64) int {
	for i := 401; i > 0; i-- {
		if exp >= (int64(math.Pow(float64(i), float64(4)))-1)/5 {
			return i
		}
	}
	return 0
}

func fieldPrice(currentFieldCount int) int64 {
	baseNumber := float64(currentFieldCount)
	return int64(math.Pow(2.5, .75*baseNumber)*baseNumber) * 1000
}

func cropState(crop Crop, plantTime int64, now int64) (state int, emoji string) {
	width := now - plantTime
	for i := 0; i < len(crop.StepHours); i++ {
		state = i
		emoji = crop.StepEmojis[i]
		band := 3600 * int64(crop.StepHours[i])
		if width > band {
			if i == len(crop.StepHours)-1 {
				state = MATURE
				emoji = crop.FruitEmoji
				break
			} else {
				width -= band
				continue
			}
		} else {
			break
		}
	}
	return
}

func now() int64 {
	return time.Now().Unix()
}

const MATURE = -1
