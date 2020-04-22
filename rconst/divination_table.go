package rconst

// Divination 吐槽结构体
type Divination struct {
	PlayerID     string `json:"playerid"`     // 玩家id
	DivinationID int64  `json:"divinationid"` // 吐槽id
	Time         int64  `json:"time"`         // 时间
	Content      string `json:"content"`      // 内容
	Name         string `json:"name"`         // 昵称
	Portrait     string `json:"portrait"`     // 头像
}

const (
	// StringDivinationID 吐槽ID
	StringDivinationID = "wecalendar:id"
	// HashDivinationPrefix 吐槽表+日期
	HashDivinationPrefix = "wecalendar:divination:"
	// ZSetDivinationRecordPrefix 吐槽记录表+日期
	ZSetDivinationRecordPrefix = "wecalendar:divinationrecord:"
	// ZSetDivinationRank 吐槽排行榜
	ZSetDivinationRank = "wecalendar:rank"
	// StringDivinationBestPrefix 最佳吐槽+日期
	StringDivinationBestPrefix = "wecalendar:divinationbest:"

	// HashDivinationConfig 吐槽配置
	HashDivinationConfig = "wecalendar:divination:config"
	// FieldDivinationFirst 虚假第一名开关
	FieldDivinationFirst = "first"
)
