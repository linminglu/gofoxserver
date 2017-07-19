package pk_base

import (
	"mj/gameServer/common/pk"

	"github.com/lovelly/leaf/util"
)

type BaseLogic struct {
	ConfigIdx int //配置索引
}

func NewBaseLogic(ConfigIdx int) *BaseLogic {
	bl := new(BaseLogic)
	bl.ConfigIdx = ConfigIdx
	return bl
}

func (lg *BaseLogic) GetCfg() *PK_CFG {
	return GetCfg(lg.ConfigIdx)
}

func (lg *BaseLogic) RandCardList(cbCardBuffer, OriDataArray []int) {

	//混乱准备
	cbBufferCount := int(len(cbCardBuffer))
	cbCardDataTemp := make([]int, cbBufferCount)
	util.DeepCopy(&cbCardDataTemp, &OriDataArray)

	//混乱扑克
	var cbRandCount int
	var cbPosition int
	for {
		if cbRandCount >= cbBufferCount {
			break
		}
		cbPosition = int(util.RandInterval(0, int(cbBufferCount-cbRandCount-1)))
		cbCardBuffer[cbRandCount] = cbCardDataTemp[cbPosition]
		cbRandCount++
		cbCardDataTemp[cbPosition] = cbCardDataTemp[cbBufferCount-cbRandCount]
	}

	return
}

//排列扑克
func (lg *BaseLogic) SortCardList(cardData []int, cardCount int) {
	logicValue := make([]int, cardCount)
	for i := 0; i < cardCount; i++ {
		logicValue[i] = lg.GetCardValue(cardData[i])
	}
	sorted := true
	last := cardCount - 1
	for {
		sorted = true
		for i := 0; i < last; i++ {
			if (logicValue[i] < logicValue[i+1]) || (logicValue[i] == logicValue[i+1] && (cardData[i] < cardData[i+1])) {
				tempData := cardData[i]
				cardData[i] = cardData[i+1]
				cardData[i+1] = tempData
				tempData = logicValue[i]
				logicValue[i] = logicValue[i+1]
				logicValue[i+1] = tempData
				sorted = false
			}
		}
		last--
		if sorted == true {
			break
		}
	}
}

//获取数值
func (lg *BaseLogic) GetCardValue(CardData int) int {
	return CardData & LOGIC_MASK_VALUE
}

//获取花色
func (lg *BaseLogic) GetCardColor(CardData int) int {
	return CardData & LOGIC_MASK_COLOR
}

func (lg *BaseLogic) CompareCard(firstCardData []int, lastCardData []int) bool {
	return false
}
func (lg *BaseLogic) GetCardType(cardData []int) int {
	return 0
}

func (lg *BaseLogic) SetParamToLogic(args interface{}) {

}

func (lg *BaseLogic) RemoveCardList(cbRemoveCard []int, cbCardData []int) ([]int, bool) {
	return nil, false
}

func (lg *BaseLogic) GetType(bCardData []int, bCardCount int) *pk.TagAnalyseType {
	return nil
}

func (lg *BaseLogic) CompareSSSCard(bInFirstList []int, bInNextList []int, bFirstCount int, bNextCount int, bComPerWithOther bool) bool {
	return false
}

func (lg *BaseLogic) GetSSSCardType(cardData []int, bCardCount int, btSpecialCard []int) int {
	return 0
}

func (lg *BaseLogic) GetCardTimes(cardType int) int {
	return 1
}

func (lg *BaseLogic) CompareCardWithParam(firstCardData []int, lastCardData []int, args []interface{}) bool {
	return false
}
