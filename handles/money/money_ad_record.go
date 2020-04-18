package money

import (
	"encoding/json"
	"weagent/gconst"
	"weagent/pb"
	"weagent/server"
	"weagent/tables"

	"github.com/golang/protobuf/proto"
)

type adRecordReq struct {
	Start int32 `json:"start"`
	End   int32 `json:"end"`
}

type adRecordItem struct {
	Earning    float32 `json:"earning"`    // 收益
	Money      float32 `json:"money"`      // 当前余额
	CreateTime string  `json:"createtime"` // 创建时间
}

type adRecordRsp struct {
	AdRecords []*adRecordItem `json:"adrecords"`
}

func adRecordHandle(c *server.StupidContext) {
	log := c.Log.WithField("func", "money.adRecordHandle")

	httpRsp := pb.HTTPResponse{
		Result: proto.Int32(int32(gconst.UnknownError)),
	}
	defer c.WriteJSONRsp(&httpRsp)

	// req
	req := &adRecordReq{}
	if err := json.Unmarshal(c.Body, req); err != nil {
		httpRsp.Result = proto.Int32(int32(gconst.ErrParse))
		httpRsp.Msg = proto.String("请求信息解析失败")
		log.Errorf("code:%d msg:%s json Unmarshal err:%s", httpRsp.GetResult(), httpRsp.GetMsg(), err.Error())
		return
	}

	log.Info("adRecordHandle enter, req:", string(c.Body))

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

	adrecords := []*tables.Adrecord{}
	err := db.Where("id = ?", playerid).Limit(end, start).Find(&adrecords)
	if err != nil {
		httpRsp.Result = proto.Int32(int32(gconst.ErrDB))
		httpRsp.Msg = proto.String("查询广告收益记录失败")
		log.Errorf("code:%d msg:%s get adrecords err, err:%s", httpRsp.GetResult(), httpRsp.GetMsg(), err.Error())
		return
	}

	// rsp
	rsp := &adRecordRsp{
		AdRecords: []*adRecordItem{},
	}
	for _, v := range adrecords {
		tmp := &adRecordItem{
			Money:      float32(v.AdMoney) / float32(100),
			CreateTime: v.CreateTime.Format("2006-01-02"),
			Earning:    float32(v.Earnings) / float32(100),
		}

		rsp.AdRecords = append(rsp.AdRecords, tmp)
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

	log.Info("adRecordHandle rsp, rsp:", string(data))

	return
}
