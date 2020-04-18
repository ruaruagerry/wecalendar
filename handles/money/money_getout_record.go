package money

import (
	"encoding/json"
	"weagent/gconst"
	"weagent/pb"
	"weagent/server"
	"weagent/tables"

	"github.com/golang/protobuf/proto"
)

type getoutRecordReq struct {
	Start int32 `json:"start"`
	End   int32 `json:"end"`
}

type getoutRecordItem struct {
	GetoutMoney float32 `json:"getoutmoney"`
	CreateTime  string  `json:"createtime"`
	Status      string  `json:"status"`
}

type getoutRecordRsp struct {
	GetoutRecords []*getoutRecordItem `json:"getoutrecords"`
}

func getoutRecordHandle(c *server.StupidContext) {
	log := c.Log.WithField("func", "money.getoutRecordHandle")

	httpRsp := pb.HTTPResponse{
		Result: proto.Int32(int32(gconst.UnknownError)),
	}
	defer c.WriteJSONRsp(&httpRsp)

	// req
	req := &getoutRecordReq{}
	if err := json.Unmarshal(c.Body, req); err != nil {
		httpRsp.Result = proto.Int32(int32(gconst.ErrParse))
		httpRsp.Msg = proto.String("请求信息解析失败")
		log.Errorf("code:%d msg:%s json Unmarshal err:%s", httpRsp.GetResult(), httpRsp.GetMsg(), err.Error())
		return
	}

	log.Info("getoutRecordHandle enter, req:", string(c.Body))

	start := int(req.Start)
	end := int(req.End)
	if start >= end {
		httpRsp.Result = proto.Int32(int32(gconst.ErrParam))
		httpRsp.Msg = proto.String("请求参数错误")
		log.Errorf("code:%d msg:%s req param err, start:%d end:%d", httpRsp.GetResult(), httpRsp.GetMsg(), start, end)
		return
	}

	db := c.DbConn
	playerid := c.UserID

	getoutrecords := []*tables.Getoutrecord{}
	err := db.Where("id = ?", playerid).Limit(end, start).Find(&getoutrecords)
	if err != nil {
		httpRsp.Result = proto.Int32(int32(gconst.ErrDB))
		httpRsp.Msg = proto.String("查询提现记录失败")
		log.Errorf("code:%d msg:%s get getoutrecords err, err:%s", httpRsp.GetResult(), httpRsp.GetMsg(), err.Error())
		return
	}

	// rsp
	rsp := &getoutRecordRsp{
		GetoutRecords: []*getoutRecordItem{},
	}
	for _, v := range getoutrecords {
		tmp := &getoutRecordItem{
			GetoutMoney: float32(v.GetoutMoney) / float32(100),
			CreateTime:  v.CreateTime.Format("2006-01-02"),
			Status:      recordStatusForString(v.Status),
		}

		rsp.GetoutRecords = append(rsp.GetoutRecords, tmp)
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

	log.Info("getoutRecordHandle rsp, rsp:", string(data))

	return
}

func recordStatusForString(status int32) string {
	statusstr := ""

	switch status {
	case tables.GetoutStatusReview:
		statusstr = "审核中"
	case tables.GetoutStatusRefused:
		statusstr = "审核拒绝"
	case tables.GetoutStatusSuccess:
		statusstr = "提现成功"
	case tables.GetoutStatusFailed:
		statusstr = "提现失败"
	}

	return statusstr
}
