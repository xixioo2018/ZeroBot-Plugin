package farm

import (
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/FloatTech/ZeroBot-Plugin/database/redis"
	"github.com/FloatTech/ZeroBot-Plugin/plugin/xiaer"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	logger "github.com/sirupsen/logrus"
)

func init() {
	engine := control.Register("FARM", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Help: "农场\n" +
			"- 农场商店 守卫商店\n" +
			"- 购买种子 查询种子\n" +
			"- 购买守卫 查询守卫\n" +
			"- 武器商店 查询守卫\n" +
			"- 种植 收菜 偷菜 浇水\n" +
			"- 我的农场 农场等级\n" +
			"- 购买土地 农场排行\n",
	})
	engine.OnFullMatch("农场").SetBlock(true).Limit(ctxext.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
			printMenu(ctx)
		})
	engine.OnFullMatch("农场商店").SetBlock(true).Limit(ctxext.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
			printCrops(ctx)
		})
	engine.OnFullMatch("守卫商店").SetBlock(true).Limit(ctxext.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
			printPets(ctx)
		})
	engine.OnFullMatch("武器商店").SetBlock(true).Limit(ctxext.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
			printArms(ctx)
		})
	engine.OnPrefix("购买种子").SetBlock(true).Limit(ctxext.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
			printHelpBuy(ctx)
		})
	engine.OnPrefix("查询种子").SetBlock(true).Limit(ctxext.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
			printHelpSearch(ctx)
		})
	engine.OnPrefix("购买守卫").SetBlock(true).Limit(ctxext.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
			printHelpBuy(ctx)
		})
	engine.OnFullMatch("农场排行").SetBlock(true).Limit(ctxext.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
			printRank(ctx)
		})
	engine.OnPrefix("种植").SetBlock(true).Limit(ctxext.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
			printHelpPlant(ctx)
		})
	engine.OnFullMatch("我的农场").SetBlock(true).Limit(ctxext.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
			printSelf(ctx)
		})
	engine.OnFullMatch("农场等级").SetBlock(true).Limit(ctxext.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
			printLevels(ctx)
		})
	engine.OnFullMatch("收菜").SetBlock(true).Limit(ctxext.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
			collect(ctx)
		})
	engine.OnPrefix("偷菜").SetBlock(true).Limit(ctxext.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
			steal(ctx)
		})
	engine.OnPrefix("浇水").SetBlock(true).Limit(ctxext.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
			water(ctx)
		})
	engine.OnFullMatch("购买土地").SetBlock(true).Limit(ctxext.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
			buyField(ctx)
		})
	engine.OnRegex("^查询([\\s]+)?(\\p{Han}+)([\\s]+)?$").SetBlock(true).Limit(ctxext.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
			txt := ctx.State["args"].(string)
			if txt != "" {
				search(ctx, txt)
			}
		})
	engine.OnRegex("^购?买([\\s]+)?(\\p{Han}+)([\\s]+)?(\\d{1,5})?([\\s]+)?$").SetBlock(true).Limit(ctxext.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
			regex_matched := ctx.State["regex_matched"].([]string)
			txt := regex_matched[2]
			if strings.EqualFold("土地", txt) {
				buyField(ctx)
				return
			}
			count, err := strconv.Atoi(regex_matched[4])
			if err != nil {
				ctx.SendChain(message.Text("ERROR:", err))
				return
			}
			buy(ctx, txt, count)
		})
	engine.OnRegex("^播?种植?([\\s]+)?(\\p{Han}+)([\\s]+)?([\\s]+)?$").SetBlock(true).Limit(ctxext.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
			regex_matched := ctx.State["regex_matched"].([]string)
			fmt.Println(regex_matched)
			fmt.Println(len(regex_matched), regex_matched[0], regex_matched[2], regex_matched[4])
			txt := regex_matched[2]
			plant(ctx, txt)
		})
}

func printMenu(ctx *zero.Ctx) {
	ctx.SendChain(message.Text(" === 农场菜单 === \n\n" +
		"农场帮助\n" +
		"农场商店 守卫商店\n" +
		"购买种子 查询种子\n" +
		"购买守卫 查询守卫\n" +
		"武器商店 查询守卫\n" +
		"种植 收菜 偷菜 浇水\n" +
		"我的农场 农场等级\n" +
		"购买土地 农场排行"))
}

func printHelp(ctx *zero.Ctx) {
	ctx.SendChain(message.Text(
		"　　农场: 机器人主人无聊开发的小游戏\n\n" +
			"货币系统: " + emojiSun + "(阳光)是农场中的基本货币\n\n" +
			"升级系统: " + emojiExp + "(经验值)可以提高农场等级\n\n" +
			"　　作物: 种植种子, 经过一段时间可以 收获" + emojiSun + "(阳光)和" + emojiExp + "(经验值)\n\n" +
			"　　土地: 土地越多, 可以同时种的种子个数\n\n" +
			"　　偷菜: 赚点小外快?\n\n" +
			"　　查询: 查询种子或者其他物品的功能 例如'查询土豆'\n\n" +
			"    守卫: 特效宠物, 防止被偷, 打盹时触发减半\n" +
			"    浇水: 获得经验值和金币, 并且增加产量, 一株植物在成熟之前每个阶段可以浇水一次"))
}

func printHelpBuy(ctx *zero.Ctx) {
	ctx.SendChain(message.Text(
		"发送 \"购买+种子名称\" 购买相应种子, 例如 \"购买土豆\".\n\n" +
			"发送 \"购买+种子名称+数量\" 购买多个种子, 例如 \"购买土豆15\".\n\n" +
			"发送 \"购买+守卫名称\" 购买相应守卫, 例如 \"购买" + petList[0].Name + "\".\n\n" +
			"使用\"农场商店\"或者\"守卫商店\"查看列表"))
}

func printHelpSearch(ctx *zero.Ctx) {
	ctx.SendChain(message.Text(
		"发送 \"查询+种子名称\" 查询预计收益, 例如 \"查询土豆\".\n\n" +
			"发送 \"查询+守卫名称\" 查询预计收益, 例如 \"查询" + petList[0].Name + "\"."))
}

func printRank(ctx *zero.Ctx) {
	groupAssets := assetsGroup(ctx.Event.GroupID)
	result := "农夫|阳光|土地|经验|等级\n"
	for _, asset := range groupAssets {
		nickName := ctx.Event.Sender.Name()
		if len(nickName) > 6 {
			nickName = nickName[:6]
		}
		result += fmt.Sprintf(
			"%5s %5d %5d %5d %5d\n",
			nickName,
			asset.Coins,
			asset.Fields,
			asset.Exp, level(asset.Exp),
		)
	}
	ctx.SendChain(message.Text(result))
}

func printHelpPlant(ctx *zero.Ctx) {
	ctx.SendChain(message.Text(
		"发送 \"种+种子名称\" 种植作物, 例如 \"种土豆\"."))
}

func printHelpSteal(ctx *zero.Ctx) {
	ctx.SendChain(message.Text(
		"发送 \"偷菜+@一个人\" 可以偷菜, 例如 \"偷菜@张三\"."))
}

func printSelf(ctx *zero.Ctx) {
	assets := assets(sendUser(ctx))
	ctx.SendChain(message.Text(fmt.Sprintf(
		"阳光　%s　%d\n"+
			"土地　%s️　%d\n"+
			"经验　%s　%d\n"+
			"等级　%s️　%d\n",
		emojiSun, assets.Coins,
		emojiField, assets.Fields,
		emojiExp, assets.Exp, emojiLevel, level(assets.Exp),
	)))
}

func printLevels(ctx *zero.Ctx) {
	assets := assets(sendUser(ctx))
	level := level(assets.Exp)
	builder := strings.Builder{}
	builder.WriteString(fmt.Sprintf("当前农场等级为%d级(%s%d), ", level, emojiExp, assets.Exp))
	if level >= 400 {
		builder.WriteString("您已满级.")
	} else {
		builder.WriteString(fmt.Sprintf("距离升级还需要%s%d", emojiExp, ((int64(math.Pow(float64(level+1), float64(4)))-1)/5)-assets.Exp))
	}
	ctx.SendChain(message.Text(builder.String()))
}

func buyField(ctx *zero.Ctx) {
	// 加锁
	lock, err := lockUnit(sendUser(ctx))
	if err != nil {
		panic(err)
	}
	defer lock.Unlock()
	//
	assets := assets(sendUser(ctx))
	fieldPrice := fieldPrice(assets.Fields)
	if assets.Coins >= fieldPrice {
		assetsCoinsInc(ctx.Event.GroupID, ctx.Event.Sender.ID, -fieldPrice)
		assetsFieldInc(ctx.Event.GroupID, ctx.Event.Sender.ID, 1)
		ctx.SendChain(message.Text("购买成功 土地+1\n" +
			fmt.Sprintf("%s ↓ %d => %d", emojiSun, fieldPrice, assets.Coins-fieldPrice)))
	} else {
		ctx.SendChain(message.Text(fmt.Sprintf("购买第%d块土地需要%s%d", assets.Fields+1, emojiSun, fieldPrice)))
	}
}

func search(ctx *zero.Ctx, name string) bool {
	for _, crop := range cropList {
		if strings.EqualFold(crop.Name, name) {
			searchCrop(ctx, crop)
			return true
		}
	}
	for _, pet := range petList {
		if strings.EqualFold(pet.Name, name) {
			searchPet(ctx, pet)
			return true
		}
	}

	for _, arm := range armList {
		if strings.EqualFold(arm.Name, name) {
			searchArm(ctx, arm)
			return true
		}
	}

	return false
}

func searchCrop(ctx *zero.Ctx, crop Crop) {
	ctx.SendChain(message.Text(
		fmt.Sprintf("%s　%s, %d级别作物, 种子售价%s%d, 成熟时间%d小时.", crop.FruitEmoji, crop.Name, crop.Level, emojiSun, crop.SeedPrice, SumInts(crop.StepHours)) +
			fmt.Sprintf(" 每株结出果实%d到%d枚, 预计最少收益%s%d+%s%d。",
				crop.FruitsMin, crop.FruitsMax,
				emojiSun, crop.FruitsMin*crop.FruitPrice,
				emojiExp, crop.FruitsMin*crop.FruitExp)))
}

func searchPet(ctx *zero.Ctx, pet Pet) {
	ctx.SendChain(message.Text(
		fmt.Sprintf("%s　%s, %d级别守卫, 守卫售价%d, 防御能力%d%% \n", getEmojiDogByType(pet.Name), pet.Name, pet.Level,
			pet.Price, pet.Probability)))
}

func searchArm(ctx *zero.Ctx, arm Arm) {
	ctx.SendChain(message.Text(
		fmt.Sprintf("%s　%s, %d级别枪支, 枪支售价%d, 攻击能力%d%%, 噪音%d%% \n", getEmojiArmByType(arm.Name), arm.Name, arm.Level,
			arm.Price, arm.Probability, arm.Noise)))
}

func printCrops(ctx *zero.Ctx) {
	// 取得数据
	assets := assets(sendUser(ctx))
	level := level(assets.Exp)
	stock := stock(sendUser(ctx))
	var builder strings.Builder
	builder.WriteString(emojiLevel + " 　　　　　　" + emojiSun + "　 " + emojiStock + "\n")
	for _, crop := range cropList {
		if crop.Level > level || crop.Level < level-20 {
			continue
		}
		builder.WriteString(fmt.Sprintf("%02d　%s　%s　%d　", crop.Level, crop.FruitEmoji, crop.Name, crop.SeedPrice))
		if crop.SeedPrice < 10 {
			builder.WriteString("     ")
		} else if crop.SeedPrice < 100 {
			builder.WriteString("   ")
		} else if crop.SeedPrice < 1000 {
			builder.WriteString(" ")
		}
		if count, ok := stock.CropCount[strconv.Itoa(crop.Level)]; ok {
			builder.WriteString(fmt.Sprintf("%d", count))
		} else {
			builder.WriteString("0")
		}
		builder.WriteString("\n")
	}
	builder.WriteString(fmt.Sprintf("\n%s　%d　　　%s　%d", emojiLevel, level, emojiSun, assets.Coins))
	ctx.SendChain(message.Text(builder.String()))
}

func printPets(ctx *zero.Ctx) {
	// 取得数据
	assets := assets(sendUser(ctx))
	level := level(assets.Exp)
	pets := pets(sendUser(ctx))
	var builder strings.Builder
	builder.WriteString(emojiLevel + " 　　　　     　　" + emojiSun + "　" + emojiStock + "\n")
	for _, pet := range petList {
		builder.WriteString(fmt.Sprintf("%02d %s %s %d ", pet.Level, getEmojiDogByType(pet.Name), pet.Name, pet.Price))
		if ContainsInt(pets.Pets, pet.Level) {
			builder.WriteString("🈶️")
		} else {
			builder.WriteString("🈚️")
		}
		builder.WriteString("\n")
	}
	builder.WriteString(fmt.Sprintf("\n%s %d   %s %d", emojiLevel, level, emojiSun, assets.Coins))
	ctx.SendChain(message.Text(builder.String()))
}

func printArms(ctx *zero.Ctx) {
	// 取得数据
	assets := assets(sendUser(ctx))
	level := level(assets.Exp)
	arms := arms(sendUser(ctx))
	var builder strings.Builder
	builder.WriteString(emojiLevel + " 　　　　     　　" + emojiSun + "　" + emojiStock + "\n")
	for _, arm := range armList {
		builder.WriteString(fmt.Sprintf("%02d %s %s %d ", arm.Level, getEmojiArmByType(arm.Name), arm.Name, arm.Price))
		if ContainsInt(arms.Arms, arm.Level) {
			builder.WriteString("🈶️")
		} else {
			builder.WriteString("🈚️")
		}
		builder.WriteString("\n")
	}
	builder.WriteString(fmt.Sprintf("\n%s %d   %s %d", emojiLevel, level, emojiSun, assets.Coins))
	ctx.SendChain(message.Text(builder.String()))
}

func getEmojiArmByType(petName string) string {
	if petName == "弩" {
		return emojiGong
	} else if petName == "麻醉针" {
		return emojiMZ
	} else {
		return emojiArm
	}
}

func getEmojiDogByType(petName string) string {
	if petName == "霸王龙" {
		return emojiTyrannosaurusRex
	} else if petName == "奥特曼" {
		return emojiATM
	} else {
		return emojiDog
	}
}

func buy(ctx *zero.Ctx, name string, number int) bool {
	// 加锁
	lock, err := lockUnit(sendUser(ctx))
	if err != nil {
		panic(err)
	}
	defer lock.Unlock()
	//
	for _, crop := range cropList {
		if strings.EqualFold(crop.Name, name) {
			buyCrop(ctx, crop, number)
			return true
		}
	}
	for _, pet := range petList {
		if strings.EqualFold(pet.Name, name) {
			buyPet(ctx, pet)
			return true
		}
	}
	for _, arm := range armList {
		if strings.EqualFold(arm.Name, name) {
			buyArm(ctx, arm)
			return true
		}
	}
	return false
}

func buyCrop(ctx *zero.Ctx, crop Crop, number int) {
	assets := assets(sendUser(ctx))
	level := level(assets.Exp)
	stock := stock(sendUser(ctx))
	if crop.Level > level {
		ctx.SendChain(message.Text(fmt.Sprintf("您不能购买超过您自身等级的作物种子, 购买%s需要%d级, 您当前为%d级. ", crop.Name, crop.Level, level)))
		return
	}
	downCoin := int64(crop.SeedPrice * number)
	if downCoin > assets.Coins {
		ctx.SendChain(message.Text(fmt.Sprintf("您的阳光不足, 购买%d枚%s种子需要%d阳光, 您只有%d阳光. ", number, crop.Name, downCoin, assets.Coins)))
		return
	}
	inStock, _ := stock.CropCount[strconv.Itoa(crop.Level)]
	toInStock := inStock + number
	if toInStock > 99 {
		ctx.SendChain(message.Text("一种种子持有量不能超过99枚"))
		return
	}
	stock.CropCount[strconv.Itoa(crop.Level)] = toInStock
	stockUpdate(stock)
	assetsCoinsInc(assets.GroupCode, assets.Uin, -downCoin)
	ctx.SendChain(message.Text(fmt.Sprintf("购买成功\n\n%s ↑ %d => %d\n%s ↓ %d => %d", crop.FruitEmoji, number, toInStock, emojiSun, downCoin, assets.Coins-downCoin)))
}

func buyPet(ctx *zero.Ctx, pet Pet) {
	assets := assets(sendUser(ctx))
	pets := pets(sendUser(ctx))
	downCoin := int64(pet.Price)
	if downCoin > assets.Coins {
		ctx.SendChain(message.Text(fmt.Sprintf("您的阳光不足, 购买%s需要%d阳光, 您只有%d阳光. ", pet.Name, downCoin, assets.Coins)))
		return
	}
	if ContainsInt(pets.Pets, pet.Level) {
		ctx.SendChain(message.Text("您已经有了该守卫"))
		return
	}
	pets.Pets = append(pets.Pets, pet.Level)
	petsUpdate(pets)
	assetsCoinsInc(assets.GroupCode, assets.Uin, -downCoin)
	ctx.SendChain(message.Text(fmt.Sprintf("购买成功\n\n%s %s\n%s ↓ %d => %d", getEmojiDogByType(pet.Name), pet.Name, emojiSun, downCoin, assets.Coins-downCoin)))
}

func buyArm(ctx *zero.Ctx, arm Arm) {
	assets := assets(sendUser(ctx))
	arms := arms(sendUser(ctx))
	downCoin := int64(arm.Price)
	if downCoin > assets.Coins {
		ctx.SendChain(message.Text(fmt.Sprintf("您的阳光不足, 购买%s需要%d阳光, 您只有%d阳光. ", arm.Name, downCoin, assets.Coins)))
		return
	}
	if ContainsInt(arms.Arms, arm.Level) {
		ctx.SendChain(message.Text("您已经有了该守卫"))
		return
	}
	arms.Arms = append(arms.Arms, arm.Level)
	armsUpdate(arms)
	assetsCoinsInc(assets.GroupCode, assets.Uin, -downCoin)
	ctx.SendChain(message.Text(fmt.Sprintf("购买成功\n\n%s %s\n%s ↓ %d => %d", getEmojiDogByType(arm.Name), arm.Name, emojiSun, downCoin, assets.Coins-downCoin)))
}

func plant(ctx *zero.Ctx, name string) bool {
	// 加锁
	lock, err := lockUnit(sendUser(ctx))
	if err != nil {
		panic(err)
	}
	defer lock.Unlock()
	//
	for _, crop := range cropList {
		if strings.EqualFold(crop.Name, name) {
			plantCrop(ctx, crop)
			return true
		}
	}
	return false
}

func plantCrop(ctx *zero.Ctx, crop Crop) {
	now := now()
	builder := strings.Builder{}
	assets := assets(sendUser(ctx))
	stock := stock(sendUser(ctx))
	land := land(sendUser(ctx))
	expUp := int64(0)
	for i := 0; i < assets.Fields; i++ {
		builder.WriteString(fmt.Sprintf("土地(%d) ", i+1))
		field, _ := land.Fields[strconv.Itoa(i)]
		if field.Level > 0 {
			cropPlanted := cropMap[field.Level]
			_, emoji := cropState(cropPlanted, field.PlantTime, now)
			builder.WriteString(fmt.Sprintf("%s (%s 已存在)", emoji, cropPlanted.Name))
		} else {
			if stock.CropCount[strconv.Itoa(crop.Level)] > 0 {
				stock.CropCount[strconv.Itoa(crop.Level)]--
				land.Fields[strconv.Itoa(i)] = Field{
					Level:     crop.Level,
					PlantTime: now,
					Watered:   map[string]int64{},
					Stealer:   []int64{},
					Alerted:   []int64{},
				}
				expUp += int64(crop.FruitExp)
				builder.WriteString(fmt.Sprintf(" => %s", crop.FruitEmoji))
			} else {
				builder.WriteString(fmt.Sprintf("%s种子不足", crop.Name))
			}
		}
		builder.WriteString("\n")
	}
	if expUp > 0 {
		stockUpdate(stock)
		landUpdate(land)
		assetsExpInc(assets.GroupCode, assets.Uin, expUp)
		builder.WriteString(fmt.Sprintf("\n%s ↑ %d => %d", emojiExp, expUp, assets.Exp+expUp))
	}
	ctx.SendChain(message.Text(builder.String()))
}

func collect(ctx *zero.Ctx) {
	// 加锁
	lock, err := lockUnit(sendUser(ctx))
	if err != nil {
		panic(err)
	}
	defer lock.Unlock()
	//
	now := now()
	builder := strings.Builder{}
	assets := assets(sendUser(ctx))
	land := land(sendUser(ctx))
	//
	expUp := int64(0)
	coinsUp := int64(0)
	var waterSet []int64
	var stealerSet []int64
	matureData := make(map[string]HarvestMessage)
	notMatureData := make(map[string]HarvestMessage)
	noData := make([]int, 0)
	for i := 0; i < assets.Fields; i++ {
		//builder.WriteString(fmt.Sprintf("土地(%d) ", i+1))
		field, _ := land.Fields[strconv.Itoa(i)]
		if field.Level > 0 {
			cropPlanted := cropMap[field.Level]
			state, emoji := cropState(cropPlanted, field.PlantTime, now)
			if state == MATURE {
				var fruitNumber int
				if cropPlanted.FruitsMax > cropPlanted.FruitsMin {
					fruitNumber = (rand.Int() % (1 + cropPlanted.FruitsMax - cropPlanted.FruitsMin)) + cropPlanted.FruitsMin
				} else {
					fruitNumber = cropPlanted.FruitsMax
				}
				fruitNumber += len(field.Watered) * 1
				for _, water := range field.Watered {
					if !ContainsInt64(waterSet, water) && ctx.Event.Sender.ID != water {
						waterSet = append(waterSet, water)
					}
				}
				fruitNumber -= len(field.Stealer) * 1
				//builder.WriteString(fmt.Sprintf("%s (%s %d枚)", emoji, cropPlanted.Name, fruitNumber))
				if len(field.Stealer) > 0 {
					//builder.WriteString(fmt.Sprintf("(被偷%d枚)", len(field.Stealer)*1))
					for _, stealer := range field.Stealer {
						if !ContainsInt64(stealerSet, stealer) {
							stealerSet = append(stealerSet, stealer)
						}
					}
				}
				expUp += int64(fruitNumber * cropPlanted.FruitExp)
				coinsUp += int64(fruitNumber * cropPlanted.FruitPrice)
				if da, ok := matureData[cropPlanted.Name]; ok {
					da.StolenCount += len(field.Stealer) * 1
					da.LandNumber = append(da.LandNumber, i+1)
					da.FruitCount += fruitNumber
					da.Emoji = emoji
					matureData[cropPlanted.Name] = da
				} else {
					matureData[cropPlanted.Name] = HarvestMessage{
						StolenCount: len(field.Stealer) * 1,
						LandNumber:  []int{i + 1},
						FruitCount:  fruitNumber,
						Emoji:       emoji,
					}
				}
				delete(land.Fields, strconv.Itoa(i))
			} else {
				if _, ok := field.Watered[strconv.Itoa(state)]; ok {
					if da, ok := notMatureData[cropPlanted.Name]; ok {
						da.LandNumber = append(da.LandNumber, i+1)
						da.Emoji = emoji
						da.IsWatered = true
						notMatureData[cropPlanted.Name] = da
					} else {
						notMatureData[cropPlanted.Name] = HarvestMessage{
							LandNumber: []int{i + 1},
							Emoji:      emoji,
							IsWatered:  true,
						}
					}
					//builder.WriteString(fmt.Sprintf("%s (%s 未成熟)", emoji+emojiWater, cropPlanted.Name))
				} else {
					//builder.WriteString(fmt.Sprintf("%s (%s 未成熟)", emoji, cropPlanted.Name))
					if da, ok := notMatureData[cropPlanted.Name]; ok {
						da.LandNumber = append(da.LandNumber, i+1)
						da.Emoji = emoji
						da.IsWatered = false
						notMatureData[cropPlanted.Name] = da
					} else {
						notMatureData[cropPlanted.Name] = HarvestMessage{
							LandNumber: []int{i + 1},
							Emoji:      emoji,
							IsWatered:  false,
						}
					}
				}

			}
		} else {
			//builder.WriteString(fmt.Sprintf("未种植"))
			noData = append(noData, i+1)
		}
		//builder.WriteString("\n")
	}
	if len(matureData) > 0 {
		for cropPlantedName, harvestMessage := range matureData {
			builder.WriteString(fmt.Sprintf("%v %s (%s %d枚)(被偷%d枚)\n",
				harvestMessage.LandNumber, harvestMessage.Emoji, cropPlantedName, harvestMessage.FruitCount, harvestMessage.StolenCount))
		}
	}

	if len(notMatureData) > 0 {
		for cropPlantedName, harvestMessage := range notMatureData {
			if harvestMessage.IsWatered {
				builder.WriteString(fmt.Sprintf("%v %s (%s 未成熟)\n", harvestMessage.LandNumber, harvestMessage.Emoji+emojiWater, cropPlantedName))
			} else {
				builder.WriteString(fmt.Sprintf("%v %s (%s 未成熟)\n", harvestMessage.LandNumber, harvestMessage.Emoji, cropPlantedName))
			}
		}
	}

	if len(noData) > 0 {
		builder.WriteString(fmt.Sprintf("%v 未种植\n", noData))
	}

	if expUp > 0 {
		landUpdate(land)
		assetsExpInc(assets.GroupCode, assets.Uin, expUp)
		assetsCoinsInc(assets.GroupCode, assets.Uin, coinsUp)
		if len(waterSet) > 0 {
			builder.WriteString("\n")
			builder.WriteString("帮你浇水的群友 : \n")
			for _, water := range waterSet {
				builder.WriteString("    " + ctx.GetGroupMemberInfo(ctx.Event.GroupID, water, false).Get("nickname").String() + "\n")
			}
		}
		if len(stealerSet) > 0 {
			builder.WriteString("\n")
			builder.WriteString("偷你菜的群友 : \n")
			for _, stealer := range stealerSet {
				builder.WriteString("    " + xiaer.CardNameInGroup(ctx, stealer) + "\n")
			}
		}
		builder.WriteString(fmt.Sprintf("\n%s ↑ %d => %d\n%s ↑ %d => %d", emojiExp, expUp, assets.Exp+expUp, emojiSun, coinsUp, assets.Coins+coinsUp))
	}
	ctx.SendChain(message.Text(builder.String()))
}

type HarvestMessage struct {
	LandNumber  []int  // 土地
	FruitCount  int    // 收获
	StolenCount int    // 被偷
	Emoji       string // 蔬菜表情
	IsWatered   bool
}

func steal(ctx *zero.Ctx) {
	// 加锁
	lock, err := lockUnit(sendUser(ctx))
	if err != nil {
		panic(err)
	}
	defer lock.Unlock()
	//
	//client.MessageFirstAt(groupMessage)
	firstAt := ctx.Event.Sender.ID
	if isAt, getFirstAt := GetFirstAt(ctx); isAt {
		firstAt = getFirstAt
	}

	if firstAt > 0 {
		builder := strings.Builder{}
		uin := ctx.Event.Sender.ID
		targetUin := firstAt
		if uin == targetUin {
			ctx.SendChain(message.Text("你不能偷自己的菜"))
			return
		} else {
			info := ctx.GetGroupMemberInfo(ctx.Event.GroupID, targetUin, true).Get("nickname")
			builder.WriteString("偷偷进入了 " + info.String() + "的农场\n\n")
		}
		now := now()
		targetAssets := assets(ctx.Event.GroupID, targetUin)

		// 判断被偷的人有没有狗
		targetPets := pets(ctx.Event.GroupID, targetUin)
		// 判断当前用户有没有枪
		targetArms := arms(ctx.Event.GroupID, uin)

		logger.Info("查询守卫", targetPets)
		if len(targetPets.Pets) > 0 {
			injuryReduction := 1
			maxLevelDog := -1
			for _, i2 := range targetPets.Pets {
				if maxLevelDog < i2 {
					maxLevelDog = i2
				}
			}

			dog, _ := petMap[maxLevelDog]

			builder.WriteString(fmt.Sprintf("%s%s ", getEmojiDogByType(dog.Name), dog.Name))

			sleepy := int(math.Abs(float64(rand.Int63()-targetUin)))%100 > dog.Probability
			alertPercentage := dog.Probability
			if sleepy {
				builder.WriteString("正在瞌睡 \n")
				alertPercentage /= 2
			}

			if len(targetArms.Arms) > 0 {
				maxLevelArm := -1
				for _, i2 := range targetArms.Arms {
					if maxLevelArm < i2 {
						maxLevelArm = i2
					}
				}

				arm, _ := armMap[maxLevelArm]

				builder.WriteString(fmt.Sprintf("%s%s ", getEmojiArmByType(arm.Name), arm.Name))
				hit := (int(math.Abs(float64(rand.Int63()-targetUin))) % 100) > arm.Probability

				if hit {
					builder.WriteString("命中 减伤50%")
					injuryReduction *= 2
					alertPercentage /= 3
				}

				wakeUp := (int(math.Abs(float64(rand.Int63()-targetUin))) % 100) < arm.Noise

				if wakeUp {
					if alertPercentage > 50 {
						alertPercentage += (100 - alertPercentage) / 2
					} else {
						alertPercentage *= 2
					}

					builder.WriteString("噪音太大 增加吵醒几率\n")
				}
			}

			alert := (int(math.Abs(float64(rand.Int63()-targetUin))) % 100) < alertPercentage
			if alert {
				bittenCoin := int64(math.Floor(float64(-1 * dog.Probability)))
				assetsCoinsInc(ctx.Event.GroupID, ctx.Event.Sender.ID, bittenCoin)
				builder.WriteString("把你咬了 损失 " + emojiSun + strconv.Itoa(dog.Probability/injuryReduction))
				ctx.SendChain(message.Text(builder.String()))
				return
			}
		}

		targetLand := land(ctx.Event.GroupID, targetUin)
		expUp := int64(0)
		coinsUp := int64(0)
		for i := 0; i < targetAssets.Fields; i++ {
			builder.WriteString(fmt.Sprintf("土地(%d) ", i+1))
			field, _ := targetLand.Fields[strconv.Itoa(i)]
			if field.Level > 0 {
				cropPlanted := cropMap[field.Level]
				state, emoji := cropState(cropPlanted, field.PlantTime, now)
				if state != MATURE {
					builder.WriteString(fmt.Sprintf("%s (%s 未成熟)", emoji, cropPlanted.Name))
				} else {
					if ContainsInt64(field.Stealer, ctx.Event.Sender.ID) {
						builder.WriteString(fmt.Sprintf("%s (%s 偷过了)", emoji, cropPlanted.Name))
					} else if len(field.Stealer) >= 2 {
						builder.WriteString(fmt.Sprintf("%s (%s 快被偷光了)", emoji, cropPlanted.Name))
					} else {
						expUp += int64(1 * cropPlanted.FruitExp)
						coinsUp += int64(1 * cropPlanted.FruitPrice)
						field.Stealer = append(field.Stealer, uin)
						targetLand.Fields[strconv.Itoa(i)] = field
						builder.WriteString(fmt.Sprintf("%s (%s %d枚)", emoji, cropPlanted.Name, 1))
					}
				}
			} else {
				builder.WriteString(fmt.Sprintf("未种植"))
			}
			builder.WriteString("\n")
		}
		if expUp > 0 {
			assets := assets(sendUser(ctx))
			landUpdate(targetLand)
			assetsExpInc(ctx.Event.GroupID, uin, expUp)
			assetsCoinsInc(ctx.Event.GroupID, uin, coinsUp)
			builder.WriteString(fmt.Sprintf("\n%s ↑ %d => %d\n%s ↑ %d => %d", emojiExp, expUp, assets.Exp+expUp, emojiSun, coinsUp, assets.Coins+coinsUp))
		}
		ctx.SendChain(message.Text(builder.String()))
	} else {
		printHelpSteal(ctx)
	}
}

func GetFirstAt(ctx *zero.Ctx) (bool, int64) {
	for _, singleMessage := range ctx.Event.Message {
		fmt.Println(singleMessage)
		if singleMessage.Type == "at" {
			fmt.Println(singleMessage.Data["qq"])
			parseInt, err := strconv.ParseInt(singleMessage.Data["qq"], 10, 64)
			if err != nil {
				continue
			}
			return true, parseInt
		}
	}
	return false, 0
}

func water(ctx *zero.Ctx) {
	// 加锁
	lock, err := lockUnit(sendUser(ctx))
	if err != nil {
		panic(err)
	}
	defer lock.Unlock()
	//
	now := now()
	uin := ctx.Event.Sender.ID
	targetUin := uin

	if isAt, getFirstAt := GetFirstAt(ctx); isAt {
		targetUin = getFirstAt
	}

	targetAssets := assets(ctx.Event.GroupID, targetUin)
	targetLand := land(ctx.Event.GroupID, targetUin)
	builder := strings.Builder{}
	if uin != targetUin {
		builder.WriteString(ctx.GetGroupMemberInfo(ctx.Event.GroupID, targetUin, false).Get("nickname").String() + "的农场\n\n")
	} else {
		builder.WriteString("浇水@为群友浇水\n\n")
	}
	expUp := int64(0)
	matureList := make([]int, 0)
	noeNeedList := make([]int, 0)
	successList := make([]int, 0)
	noneList := make([]int, 0)
	for i := 0; i < targetAssets.Fields; i++ {
		//builder.WriteString(fmt.Sprintf("地(%d) ", i+1))
		field, _ := targetLand.Fields[strconv.Itoa(i)]
		if field.Level > 0 {
			cropPlanted := cropMap[field.Level]
			state, _ := cropState(cropPlanted, field.PlantTime, now)
			if state == MATURE {
				matureList = append(matureList, i+1)
				//builder.WriteString(fmt.Sprintf("%s (%s 成熟)", emoji, cropPlanted.Name))
			} else {
				if _, ok := field.Watered[strconv.Itoa(state)]; ok {
					noeNeedList = append(noeNeedList, i+1)
					//builder.WriteString(fmt.Sprintf("%s (%s 无需)", emoji+emojiWater, cropPlanted.Name))
				} else {
					successList = append(successList, i+1)
					targetLand.Fields[strconv.Itoa(i)].Watered[strconv.Itoa(state)] = uin
					expUp += int64(cropPlanted.FruitExp)
					//builder.WriteString(fmt.Sprintf("%s (%s 成功)", emoji+emojiRain, cropPlanted.Name))
				}
			}
		} else {
			noneList = append(noneList, i+1)
		}
	}
	if len(matureList) > 0 {
		builder.WriteString(fmt.Sprintf("%v (成熟)\n", matureList))
	}
	if len(noeNeedList) > 0 {
		builder.WriteString(fmt.Sprintf("%v (%s 无需浇水)\n", noeNeedList, emojiWater))
	}
	if len(successList) > 0 {
		builder.WriteString(fmt.Sprintf("%v (%s 浇水成功)\n", successList, emojiRain))
	}
	if len(noneList) > 0 {
		builder.WriteString(fmt.Sprintf("%v 空地\n", noneList))
	}

	if expUp > 0 {
		assets := assets(sendUser(ctx))
		landUpdate(targetLand)
		assetsExpInc(ctx.Event.GroupID, uin, expUp)
		builder.WriteString(fmt.Sprintf("\n%s ↑ %d => %d", emojiExp, expUp, assets.Exp+expUp))
	}
	ctx.SendChain(message.Text(builder.String()))
}

//func sendUser(groupMessage *message.GroupMessage) (groupCode int64, uin int64) {
//	return ctx.Event.GroupID, ctx.Event.Sender.ID
//}

func sendUser(ctx *zero.Ctx) (groupCode int64, uin int64) {
	return ctx.Event.GroupID, ctx.Event.UserID
	//return ctx.Event.GroupID, ctx.Event.Sender.ID
}

func lockUnit(groupCode int64, uin int64) (*redis.Lock, error) {
	return redis.TryLock(fmt.Sprintf("BOT::GAME::FARM::%v::%v::LOCK", groupCode, uin), time.Second*5, time.Minute)
}
