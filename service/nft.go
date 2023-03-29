package service

import (
	"encoding/json"
	"fmt"
	"ggslayer/utils"
)

//FindBalanceNftByPage 查询地址下面所有nft
func (s *BalanceService) FindBalanceNftByPage(network string, page int) (nftData *NftData, err error) {
	nurl := fmt.Sprintf("https://%s.tokenview.io/api/tokens/address/nft/all/%s/%s/%d/50", network, network, s.Address, page)
	nftRes, err := utils.NetLibGetV3(nurl, nil)
	if err != nil {
		return
	}
	nftData = &NftData{}
	rs := &Result{Data: nftData}
	err = json.Unmarshal(nftRes, rs)
	return
}
