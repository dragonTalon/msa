package stock

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	log "github.com/sirupsen/logrus"
	"msa/pkg/logic/tools/safetool"
	"msa/pkg/model"
	mas_utils "msa/pkg/utils"
)

type BoardRankParam struct {
	BoardType string `json:"board_type" jsonschema:"description=board type: 01=Shenwan industry, 02=concept board, 03=region board, default: 01"`
	Order     int    `json:"order" jsonschema:"description=sort order: 0=descending (top gainers), 1=ascending (top losers), default: 0"`
	Count     int    `json:"count" jsonschema:"description=number of boards to return, default: 20"`
}

type BoardRank struct{}

func (br *BoardRank) GetToolInfo() (tool.BaseTool, error) {
	return utils.InferTool(br.GetName(), br.GetDescription(), GetBoardRank)
}

func (br *BoardRank) GetName() string { return "get_board_rank" }

func (br *BoardRank) GetDescription() string {
	return "获取概念/行业/地域板块排行，返回板块名称、涨跌幅、领涨股等 | Get board ranking by concept/industry/region, returns board name, change%, top gainer"
}

func (br *BoardRank) GetToolGroup() model.ToolGroup { return model.StockToolGroup }

func GetBoardRank(ctx context.Context, param *BoardRankParam) (string, error) {
	return safetool.SafeExecute("get_board_rank", fmt.Sprintf("board_type: %s", param.BoardType), func() (string, error) {
		return doGetBoardRank(ctx, param)
	})
}

func doGetBoardRank(ctx context.Context, param *BoardRankParam) (string, error) {
	if param == nil {
		param = &BoardRankParam{}
	}
	if param.BoardType == "" {
		param.BoardType = "01"
	}
	if param.Count <= 0 || param.Count > 100 {
		param.Count = 20
	}

	apiURL := fmt.Sprintf("%s?l=%d&p=1&t=%s/averatio&ordertype=&o=%d",
		model.FinanceBoardRank, param.Count, param.BoardType, param.Order)

	resp, err := mas_utils.GetRestyClient().R().Get(apiURL)
	if err != nil {
		return model.NewErrorResult(fmt.Sprintf("HTTP request failed: %v", err)), nil
	}

	var brResp model.BoardRankResp
	if err := json.Unmarshal(resp.Body(), &brResp); err != nil {
		return model.NewErrorResult(fmt.Sprintf("JSON解析失败: %v", err)), nil
	}
	if brResp.Code != 0 {
		return model.NewErrorResult(fmt.Sprintf("API error: code=%d, msg=%s", brResp.Code, brResp.Msg)), nil
	}

	log.Infof("get_board_rank: type=%s order=%d count=%d", param.BoardType, param.Order, len(brResp.Data))
	return model.NewSuccessResult(brResp.Data, fmt.Sprintf("获取板块排行成功, %d条", len(brResp.Data))), nil
}
