package phone

import (
	"encoding/json"
	"weagent/gconst"
	"weagent/pb"
	"weagent/rconst"
	"weagent/server"

	"github.com/golang/protobuf/proto"
	"github.com/gomodule/redigo/redis"
)

type getcodeReq struct {
	Phone string `json:"phone"`
}

type getcodeRsp struct {
	Remain int32 `json:"remain"`
}

func getcodeHandle(c *server.StupidContext) {
	log := c.Log.WithField("func", "phone.getcodeHandle")

	httpRsp := pb.HTTPResponse{
		Result: proto.Int32(int32(gconst.UnknownError)),
	}
	defer c.WriteJSONRsp(&httpRsp)

	// req
	req := &getcodeReq{}
	if err := json.Unmarshal(c.Body, req); err != nil {
		httpRsp.Result = proto.Int32(int32(gconst.ErrParse))
		httpRsp.Msg = proto.String("请求信息解析失败")
		log.Errorf("code:%d msg:%s json Unmarshal err:%s", httpRsp.GetResult(), httpRsp.GetMsg(), err.Error())
		return
	}

	log.Info("getoutRecordHandle enter, req:", string(c.Body))

	conn := c.RedisConn
	playerid := c.UserID

	phoneno := req.Phone

	// 先验证手机号格式
	phonevalid := verifyPhonFormat(req.Phone)
	if !phonevalid {
		httpRsp.Result = proto.Int32(int32(gconst.ErrPhoneFormat))
		httpRsp.Msg = proto.String("手机号格式错误")
		log.Errorf("code:%d msg:%s, phone no is invalid, phone:%s", httpRsp.GetResult(), httpRsp.GetMsg(), phoneno)
		return
	}

	conn.Send("MULTI")
	conn.Send("EXISTS", rconst.StringPhoneGetCodeTagPrefix+playerid)
	redisMDArray, err := redis.Values(conn.Do("EXEC"))
	if err != nil {
		httpRsp.Result = proto.Int32(int32(gconst.ErrRedis))
		httpRsp.Msg = proto.String("统一获取缓存操作失败")
		log.Errorf("code:%d msg:%s, redisMDArray Values err, err:%s", httpRsp.GetResult(), httpRsp.GetMsg(), err.Error())
		return
	}

	// 间隔标识
	codetag, _ := redis.Bool(redisMDArray[0], err)
	if codetag {
		httpRsp.Result = proto.Int32(int32(gconst.ErrPhoneGetCodeFast))
		httpRsp.Msg = proto.String("获取验证码速度过快，请稍后再试")
		log.Errorf("code:%d msg:%s, get phone code fast", httpRsp.GetResult(), httpRsp.GetMsg())
		return
	}

	// 生成验证码
	phonecode := getValidCode()

	// 手机信息
	msg := &rconst.PhoneMsg{
		Phone: phoneno,
		Code:  phonecode,
	}

	conn.Send("MULTI")
	conn.Send("SETEX", rconst.StringPhoneGetCodeTagPrefix+playerid, getCodeInterval, "0")
	msgbyte, err := json.Marshal(msg)
	if err != nil {
		httpRsp.Result = proto.Int32(int32(gconst.ErrParse))
		httpRsp.Msg = proto.String("手机信息marshal解析失败")
		log.Errorf("code:%d msg:%s, phone msg marshal err, err:%s", httpRsp.GetResult(), httpRsp.GetMsg(), err.Error())
		return
	}
	conn.Send("SETEX", rconst.StringPhoneCodePrefix+playerid, codeLiveTime, msgbyte)
	_, err = conn.Do("EXEC")
	if err != nil {
		httpRsp.Result = proto.Int32(int32(gconst.ErrRedis))
		httpRsp.Msg = proto.String("统一存储缓存操作失败")
		log.Errorf("code:%d msg:%s, redis exec err, err:%s", httpRsp.GetResult(), httpRsp.GetMsg(), err.Error())
		return
	}
	log.Info("## phone msg:", string(msgbyte))

	// 发短信
	err = sendPhoneMsg(phoneno, phonecode)
	if err != nil {
		httpRsp.Result = proto.Int32(int32(gconst.ErrPhoneSendMsg))
		httpRsp.Msg = proto.String("发送手机短信失败")
		log.Errorf("code:%d msg:%s, send phone msg err, err:%s", httpRsp.GetResult(), httpRsp.GetMsg(), err.Error())
		return
	}

	rsp := &getcodeRsp{
		Remain: int32(getCodeInterval),
	}
	data, err := json.Marshal(rsp)
	if err != nil {
		httpRsp.Result = proto.Int32(int32(gconst.ErrParse))
		httpRsp.Msg = proto.String("返回信息marshal解析失败")
		log.Errorf("code:%d msg:%s json marshal err, err:%s", httpRsp.GetResult(), httpRsp.GetMsg(), err.Error())
		return
	}
	httpRsp.Result = proto.Int32(int32(gconst.Success))
	httpRsp.Data = data

	log.Info("getcodeHandle rsp, rsp:", string(data))

	return
}
