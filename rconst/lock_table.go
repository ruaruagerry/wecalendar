package rconst

const (
	/* ---------- setup ---------- */

	// StringLockRealModifyHandlePrefix 实名认证锁
	StringLockRealModifyHandlePrefix = "weagent:lock:realmodifyhandle:"

	/* ---------- money ---------- */

	// StringLockMoneyAdSeePrefix 广告上报锁
	StringLockMoneyAdSeePrefix = "weagent:lock:moneyadsee:"
	// StringLockMoneyAdClickPrefix 广告点击锁
	StringLockMoneyAdClickPrefix = "weagent:lock:moneyadclick:"
	// StringLockMoneyGetoutApplyPrefix 申请提现锁
	StringLockMoneyGetoutApplyPrefix = "weagent:lock:moneygetoutapply:"

	/* ---------- game ---------- */

	// StringLockGameRebirthGetPrefix 获取重生次数锁
	StringLockGameRebirthGetPrefix = "weagent:lock:rebirthget:"
	// StringLockGameRebirthUsePrefix 消耗重生次数锁
	StringLockGameRebirthUsePrefix = "weagent:lock:rebirthuse:"
	// StringLockGameScoreUpdatePrefix 更新玩家分数锁
	StringLockGameScoreUpdatePrefix = "weagent:lock:scoreupdate:"
)
