package loader

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/unisat-wallet/libbrc20-indexer/model"
	"github.com/unisat-wallet/libbrc20-indexer/utils"
)

func LoadBRC20InputData(fname string) ([]*model.InscriptionBRC20Data, error) {
	file, err := os.Open(fname)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var brc20Datas []*model.InscriptionBRC20Data
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Split(line, " ")

		if len(fields) != 10 {
			return nil, fmt.Errorf("invalid data format")
		}

		var data model.InscriptionBRC20Data
		data.IsTransfer, err = strconv.ParseBool(fields[0])
		if err != nil {
			return nil, err
		}

		txid, err := hex.DecodeString(fields[1])
		if err != nil {
			return nil, err
		}
		data.TxId = string(txid)

		txIdx, err := strconv.ParseUint(fields[2], 10, 32)
		if err != nil {
			return nil, err
		}
		data.TxIdx = uint32(txIdx)

		satoshi, err := strconv.ParseUint(fields[3], 10, 64)
		if err != nil {
			return nil, err
		}
		data.Satoshi = uint64(satoshi)

		pkScript, err := hex.DecodeString(fields[4])
		if err != nil {
			return nil, err
		}
		data.PkScript = string(pkScript)

		inscriptionNumber, err := strconv.ParseInt(fields[5], 10, 64)
		if err != nil {
			return nil, err
		}
		data.InscriptionNumber = int64(inscriptionNumber)

		data.ContentBody, err = hex.DecodeString(fields[6])
		if err != nil {
			return nil, err
		}

		createIdxKey, err := hex.DecodeString(fields[7])
		if err != nil {
			return nil, err
		}

		data.CreateIdxKey = string(createIdxKey)

		height, err := strconv.ParseUint(fields[8], 10, 32)
		if err != nil {
			return nil, err
		}
		data.Height = uint32(height)

		blockTime, err := strconv.ParseUint(fields[9], 10, 32)
		if err != nil {
			return nil, err
		}
		data.BlockTime = uint32(blockTime)

		brc20Datas = append(brc20Datas, &data)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return brc20Datas, nil
}

func DumpTickerInfoMap(fname string,
	inscriptionsTickerInfoMap map[string]*model.BRC20TokenInfo,
	userTokensBalanceData map[string]map[string]*model.BRC20TokenBalance,
	tokenUsersBalanceData map[string]map[string]*model.BRC20TokenBalance,
	inscriptionsValidTransferDataMap map[string]model.InscriptionBRC20InfoResp,
) {

	file, err := os.OpenFile(fname, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0777)
	if err != nil {
		log.Fatalf("open block index file failed, %s", err)
		return
	}
	defer file.Close()

	var allTickers []string
	for ticker := range inscriptionsTickerInfoMap {
		allTickers = append(allTickers, ticker)
	}
	sort.SliceStable(allTickers, func(i, j int) bool {
		return allTickers[i] < allTickers[j]
	})

	for _, ticker := range allTickers {
		info := inscriptionsTickerInfoMap[ticker]
		nValid := 0
		for _, h := range info.History {
			if h.Valid {
				nValid++
			}
		}

		fmt.Fprintf(file, "%s history: %d, valid: %d, minted: %s, holders: %d\n",
			info.Ticker,
			len(info.History),
			nValid,
			info.Deploy.TotalMinted.String(),
			len(tokenUsersBalanceData[ticker]),
		)

		var allHolders []string
		for holder := range tokenUsersBalanceData[ticker] {
			allHolders = append(allHolders, holder)
		}
		sort.SliceStable(allHolders, func(i, j int) bool {
			return allHolders[i] < allHolders[j]
		})

		// holders
		for _, holder := range allHolders {
			balanceData := tokenUsersBalanceData[ticker][holder]

			address, err := utils.GetAddressFromScript([]byte(balanceData.PkScript), &chaincfg.MainNetParams)
			if err != nil {
				address = hex.EncodeToString([]byte(balanceData.PkScript))
			}
			fmt.Fprintf(file, "%s %s history: %d, transfer: %d, balance: %s, tokens: %d\n",
				info.Ticker,
				address,
				len(balanceData.History),
				len(balanceData.ValidTransferMap),
				balanceData.OverallBalance.String(),
				len(userTokensBalanceData[string(balanceData.PkScript)]),
			)
		}
	}
}
