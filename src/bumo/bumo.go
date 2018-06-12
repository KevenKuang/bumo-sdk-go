// bumo
package bumo

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/bumoproject/bumo-sdk-go/src/3rd/proto"
	"github.com/bumoproject/bumo-sdk-go/src/keypair"
	"github.com/bumoproject/bumo-sdk-go/src/protocol"
	"github.com/bumoproject/bumo-sdk-go/src/signature"
)

const Conversion float64 = 100000000

type deal struct {
	Items []Items `json:"items"`
}
type Items struct {
	Transaction_blob string       `json:"transaction_blob"`
	Signatures       []Signatures `json:"signatures"`
}
type Signatures struct {
	Sign_data  string `json:"sign_data"`
	Public_key string `json:"public_key"`
}

type BumoSdk struct {
	Account  AccountOperation
	Contract ContractOperation
}

//新建链接
func (bumosdk *BumoSdk) Newbumo(ip string) Error {
	if ip == "" {
		return sdkErr(INVALID_PARAMETER)
	}
	bumosdk.Account.url = ip
	bumosdk.Contract.url = ip
	Err.Code = SUCCESS
	Err.Err = nil
	return Err
}

//获取区块高度
func (bumosdk *BumoSdk) GetBlockNumber() (int64, Error) {
	var buf bytes.Buffer
	get := "/getLedger"
	buf.WriteString(bumosdk.Account.url)
	buf.WriteString(get)
	url := buf.String()
	client := &http.Client{}
	reqest, err := http.NewRequest("GET", url, nil)
	if err != nil {
		Err.Code = HTTP_NEWREQUEST_ERROR
		Err.Err = err
		return 0, Err
	}
	response, err := client.Do(reqest)
	if err != nil {
		Err.Code = CLIENT_DO_ERROR
		Err.Err = err
		return 0, Err
	}
	if response.StatusCode == 200 {
		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			Err.Code = IOUTIL_READALL_ERROR
			Err.Err = err
			return 0, Err
		}
		var data map[string]interface{}
		err = json.Unmarshal(body, &data)
		if err != nil {
			Err.Code = JSON_UNMARSHAL_ERROR
			Err.Err = err
			return 0, Err
		}
		if data["error_code"].(float64) == 0 {
			result := data["result"].(map[string]interface{})
			header := result["header"].(map[string]interface{})
			seq := header["seq"].(float64)
			Err.Code = SUCCESS
			Err.Err = nil
			return int64(seq), Err
		} else {
			if data["error_code"].(float64) == 4 {
				return 0, sdkErr(BLOCK_NOT_EXIST)
			}
			return 0, getErr(data["error_code"].(float64))
		}
	} else {
		Err.Code = response.StatusCode
		Err.Err = errors.New(response.Status)
		return 0, Err
	}
}

//检查区块同步
func (bumosdk *BumoSdk) CheckBlockStatus() (bool, Error) {
	var buf bytes.Buffer
	var ret bool
	get := "getModulesStatus"
	buf.WriteString(bumosdk.Account.url)
	buf.WriteString("/")
	buf.WriteString(get)
	url := buf.String()
	client := &http.Client{}
	reqest, err := http.NewRequest("GET", url, nil)
	if err != nil {
		Err.Code = HTTP_NEWREQUEST_ERROR
		Err.Err = err
		return false, Err
	}
	response, err := client.Do(reqest)
	if err != nil {
		Err.Code = CLIENT_DO_ERROR
		Err.Err = err
		return false, Err
	}
	if response.StatusCode == 200 {
		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			Err.Code = IOUTIL_READALL_ERROR
			Err.Err = err
			return false, Err
		}
		var data map[string]interface{}
		err = json.Unmarshal(body, &data)
		if err != nil {
			Err.Code = JSON_UNMARSHAL_ERROR
			Err.Err = err
			return false, Err
		}
		ledger_manager := data["ledger_manager"].(map[string]interface{})
		if ledger_manager["chain_max_ledger_seq"] == ledger_manager["ledger_sequence"] {
			ret = true
		}
	}
	Err.Code = SUCCESS
	Err.Err = nil
	return ret, Err
}

//根据hash查询交易
func (bumosdk *BumoSdk) GetTransaction(transactionHash string) (string, Error) {
	if len(transactionHash) != 64 {
		return "", sdkErr(INVALID_PARAMETER)
	}
	str := "/getTransactionHistory?hash="
	var buf bytes.Buffer
	buf.WriteString(bumosdk.Account.url)
	buf.WriteString(str)
	buf.WriteString(transactionHash)
	url := buf.String()
	client := &http.Client{}
	reqest, err := http.NewRequest("GET", url, nil)
	if err != nil {
		Err.Code = HTTP_NEWREQUEST_ERROR
		Err.Err = err
		return "", Err
	}
	response, err := client.Do(reqest)
	if err != nil {
		Err.Code = CLIENT_DO_ERROR
		Err.Err = err
		return "", Err
	}
	if response.StatusCode == 200 {
		body, _ := ioutil.ReadAll(response.Body)
		var data map[string]interface{}
		err := json.Unmarshal(body, &data)
		if err != nil {
			Err.Code = JSON_UNMARSHAL_ERROR
			Err.Err = err
			return "", Err
		}
		if data["error_code"].(float64) == 0 {
			result := data["result"]
			Mdata, err := json.Marshal(&result)
			if err != nil {
				Err.Code = JSON_MARSHAL_ERROR
				Err.Err = err
				return "", Err
			}
			Err.Code = SUCCESS
			Err.Err = nil
			return string(Mdata), Err
		} else {
			if data["error_code"].(float64) == 4 {
				return "", sdkErr(TRANSACTION_NOT_EXIST)
			}
			return "", getErr(data["error_code"].(float64))
		}
	} else {
		Err.Code = response.StatusCode
		Err.Err = errors.New(response.Status)
		return "", Err
	}
}

//根据高度查询交易
func (bumosdk *BumoSdk) GetBlock(blockNumber int64) (string, Error) {
	if blockNumber <= 0 {
		return "", sdkErr(INVALID_PARAMETER)
	}
	bnstr := strconv.FormatInt(blockNumber, 10)
	str := "/getTransactionHistory?ledger_seq="
	var buf bytes.Buffer
	buf.WriteString(bumosdk.Account.url)
	buf.WriteString(str)
	buf.WriteString(bnstr)
	url := buf.String()
	client := &http.Client{}
	reqest, err := http.NewRequest("GET", url, nil)
	if err != nil {
		Err.Code = HTTP_NEWREQUEST_ERROR
		Err.Err = err
		return "", Err
	}
	response, err := client.Do(reqest)
	if err != nil {
		Err.Code = CLIENT_DO_ERROR
		Err.Err = err
		return "", Err
	}
	if response.StatusCode == 200 {
		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			Err.Code = IOUTIL_READALL_ERROR
			Err.Err = err
			return "", Err
		}
		var data map[string]interface{}
		err = json.Unmarshal(body, &data)
		if err != nil {
			Err.Code = JSON_UNMARSHAL_ERROR
			Err.Err = err
			return "", Err
		}
		if data["error_code"].(float64) == 0 {
			result := data["result"]
			Mdata, err := json.Marshal(&result)
			if err != nil {
				Err.Code = JSON_MARSHAL_ERROR
				Err.Err = err
				return "", Err
			}
			Err.Code = SUCCESS
			Err.Err = nil
			return string(Mdata), Err
		} else {
			if data["error_code"].(float64) == 4 {
				return "", sdkErr(TRANSACTION_NOT_EXIST)
			}
			return "", getErr(data["error_code"].(float64))
		}
	} else {
		Err.Code = response.StatusCode
		Err.Err = errors.New(response.Status)
		return "", Err
	}
}

//查询区块头
func (bumosdk *BumoSdk) GetLedger(blockNumber int64) (string, Error) {
	if blockNumber <= 0 {
		return "", sdkErr(SUCCESS)
	}
	bnstr := strconv.FormatInt(blockNumber, 10)
	str := "/getLedger?seq="
	var buf bytes.Buffer
	buf.WriteString(bumosdk.Account.url)
	buf.WriteString(str)
	buf.WriteString(bnstr)
	url := buf.String()
	client := &http.Client{}
	reqest, err := http.NewRequest("GET", url, nil)
	if err != nil {
		Err.Code = HTTP_NEWREQUEST_ERROR
		Err.Err = err
		return "", Err
	}
	response, err := client.Do(reqest)
	if err != nil {
		Err.Code = CLIENT_DO_ERROR
		Err.Err = err
		return "", Err
	}
	if response.StatusCode == 200 {
		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			Err.Code = IOUTIL_READALL_ERROR
			Err.Err = err
			return "", Err
		}
		var data map[string]interface{}
		err = json.Unmarshal(body, &data)
		if err != nil {
			Err.Code = JSON_UNMARSHAL_ERROR
			Err.Err = err
			return "", Err
		}
		if data["error_code"].(float64) == 0 {
			result := data["result"]
			Mdata, err := json.Marshal(&result)
			if err != nil {
				Err.Code = JSON_MARSHAL_ERROR
				Err.Err = err
				return "", Err
			}
			Err.Code = SUCCESS
			Err.Err = nil
			return string(Mdata), Err
		} else {
			if data["error_code"].(float64) == 4 {
				return "", sdkErr(BLOCK_NOT_EXIST)
			}
			return "", getErr(data["error_code"].(float64))
		}
	} else {
		Err.Code = response.StatusCode
		Err.Err = errors.New(response.Status)
		return "", Err
	}
}

//生成交易(默认费用)
func (bumosdk *BumoSdk) CreateTransactionWithDefaultFee(sourceAddress string, nonce int64, operation []byte) (string, Error) {
	if !keypair.CheckAddress(sourceAddress) {
		return "", sdkErr(INVALID_PARAMETER)
	}
	if nonce < 0 {
		return "", sdkErr(INVALID_NONCE)
	}
	if operation == nil {
		return "", sdkErr(INVALID_OPERATION)
	}
	var feeLimit int64
	gasPrice := bumosdk.getGasPrice()
	operations := &protocol.Operation{}
	err := proto.Unmarshal(operation, operations)
	if err != nil {
		Err.Code = PROTO_UNMARSHAL_ERROR
		Err.Err = err
		return "", Err
	}
	if operations.Type == protocol.Operation_ISSUE_ASSET {
		feeLimit = (5000000 + 1000) * gasPrice
	} else if operations.Type == protocol.Operation_CREATE_ACCOUNT {
		feeLimit = (1000000 + 1000) * gasPrice
	} else {
		feeLimit = 1000 * gasPrice
	}
	Operations := []*protocol.Operation{
		{},
	}
	err = proto.Unmarshal(operation, Operations[0])
	if err != nil {
		Err.Code = PROTO_UNMARSHAL_ERROR
		Err.Err = err
		return "", Err
	}
	Transaction := &protocol.Transaction{
		SourceAddress: sourceAddress,
		Nonce:         nonce,
		FeeLimit:      feeLimit,
		GasPrice:      gasPrice,
		Operations:    Operations,
	}
	data, err := proto.Marshal(Transaction)
	if err != nil {
		Err.Code = PROTO_MARSHAL_ERROR
		Err.Err = err
		return "", Err
	}
	dataHash := hex.EncodeToString(data)
	Err.Code = SUCCESS
	Err.Err = nil
	return dataHash, Err
}

//生成交易(传入费用)
func (bumosdk *BumoSdk) CreateTransactionWithFee(sourceAddress string, nonce int64, gasPrice int64, feeLimit int64, operation []byte) (string, Error) {
	if !keypair.CheckAddress(sourceAddress) {
		return "", sdkErr(INVALID_SOURCEADDRESS)
	}
	newgasPrice := bumosdk.getGasPrice()
	if nonce < 0 {
		return "", sdkErr(INVALID_NONCE)
	}
	if gasPrice < newgasPrice {
		return "", sdkErr(INVALID_GASPRICE)
	}
	if feeLimit < newgasPrice*1000 {
		return "", sdkErr(INVALID_FEELIMIT)
	}
	if operation == nil {
		return "", sdkErr(INVALID_OPERATION)
	}
	operations := &protocol.Operation{}
	err := proto.Unmarshal(operation, operations)
	if err != nil {
		Err.Code = PROTO_UNMARSHAL_ERROR
		Err.Err = err
		return "", Err
	}
	Operations := []*protocol.Operation{
		{},
	}
	err = proto.Unmarshal(operation, Operations[0])
	Transaction := &protocol.Transaction{
		SourceAddress: sourceAddress,
		Nonce:         nonce,
		FeeLimit:      feeLimit,
		GasPrice:      gasPrice,
		Operations:    Operations,
	}
	data, err := proto.Marshal(Transaction)
	if err != nil {
		Err.Code = PROTO_MARSHAL_ERROR
		Err.Err = err
		return "", Err
	}
	dataHash := hex.EncodeToString(data)
	Err.Code = SUCCESS
	Err.Err = nil
	return dataHash, Err
}

//评估费用
func (bumosdk *BumoSdk) EvaluationFee(sourceAddress string, nonce int64, operation []byte, signatureNumber int64) (actualFee int64, gasPrice int64, Err Error) {
	if !keypair.CheckAddress(sourceAddress) {
		return 0, 0, sdkErr(INVALID_SOURCEADDRESS)
	}
	if nonce <= 0 {
		return 0, 0, sdkErr(INVALID_NONCE)
	}
	if operation == nil {
		return 0, 0, sdkErr(INVALID_OPERATION)
	}
	if signatureNumber < 0 {
		return 0, 0, sdkErr(INVALID_SIGNATURENUMBER)
	}

	operations := &protocol.Operation{}
	err := proto.Unmarshal(operation, operations)
	if err != nil {
		Err.Code = PROTO_UNMARSHAL_ERROR
		Err.Err = err
		return 0, 0, Err
	}
	Operations := []*protocol.Operation{
		{},
	}
	err = proto.Unmarshal(operation, Operations[0])
	request := make(map[string]interface{})
	transactionJson := make(map[string]interface{})
	transactionJson["source_address"] = sourceAddress
	transactionJson["nonce"] = nonce
	transactionJson["operations"] = Operations
	transactionJson["signature_number"] = signatureNumber
	items := make([]map[string]interface{}, 1)
	items[0] = make(map[string]interface{})
	items[0]["transaction_json"] = transactionJson
	request["items"] = items
	deal_js, err := json.Marshal(request)
	if err != nil {
		Err.Code = JSON_MARSHAL_ERROR
		Err.Err = err
		return 0, 0, Err
	}
	str := "/testTransaction"
	var buf bytes.Buffer
	buf.WriteString(bumosdk.Account.url)
	buf.WriteString(str)
	url := buf.String()
	client := &http.Client{}
	reqest, err := http.NewRequest("POST", url, bytes.NewReader(deal_js))
	if err != nil {
		Err.Code = HTTP_NEWREQUEST_ERROR
		Err.Err = err
		return 0, 0, Err
	}
	response, err := client.Do(reqest)
	if err != nil {
		Err.Code = CLIENT_DO_ERROR
		Err.Err = err
		return 0, 0, Err
	}

	if response.StatusCode == 200 {
		data := make(map[string]interface{})
		decoder := json.NewDecoder(response.Body)
		decoder.UseNumber()
		err = decoder.Decode(&data)
		if err != nil {
			Err.Code = DECODER_DECODE_ERROR
			Err.Err = err
			return 0, 0, Err
		}
		if data["error_code"].(json.Number) == "0" {
			result := data["result"].(map[string]interface{})
			txs, ok := result["txs"].([]interface{})
			if !ok {
				return 0, 0, sdkErr(TRANSACTION_INVALID)
			}
			tx, ok := txs[0].(map[string]interface{})
			if !ok {
				return 0, 0, sdkErr(TRANSACTION_INVALID)
			}
			if tx["actual_fee"] == nil {
				return 0, 0, sdkErr(TRANSACTION_INVALID)
			}
			actualFee := tx["actual_fee"].(float64)
			transactionEnv := tx["transaction_env"].(map[string]interface{})
			transaction := transactionEnv["transaction"].(map[string]interface{})
			if transaction["gas_price"] == nil {
				return 0, 0, sdkErr(TRANSACTION_INVALID)
			}
			gasPrice := transaction["gas_price"].(float64)
			Err.Code = SUCCESS
			Err.Err = nil
			return int64(actualFee), int64(gasPrice), Err
		} else {
			Err.Code = int(data["error_code"].(float64) + 10000)
			Err.Err = errors.New(data["error_desc"].(string))
			return 0, 0, Err
		}
	} else {
		Err.Code = response.StatusCode
		Err.Err = errors.New(response.Status)
		return 0, 0, Err
	}
}

//签名
func (bumosdk *BumoSdk) SignTransaction(transactionBlob string, privateKey string) (string, string, Error) {
	if transactionBlob == "" {
		return "", "", sdkErr(INVALID_TRANSACTIONBLOB)
	}
	if !keypair.CheckPrivateKey(privateKey) {
		return "", "", sdkErr(INVALID_PRIVATEKEY)
	}
	publicKey, err := keypair.GetEncPublicKey(privateKey)
	if err != nil {
		Err.Code = KEYPAIR_GETENCPUBLICKEY_ERROR
		Err.Err = err
		return "", "", Err
	}
	TransactionBlob, err := hex.DecodeString(transactionBlob)
	if err != nil {
		Err.Code = HEX_DECODESTRING_ERROR
		Err.Err = err
		return "", "", Err
	}
	sign_data, err := signature.Sign(privateKey, TransactionBlob)
	if err != nil {
		Err.Code = SIGNATURE_SIGN_ERROR
		Err.Err = err
		return "", "", Err
	}
	return sign_data, publicKey, Err
}

//提交交易
func (bumosdk *BumoSdk) SubmitTransaction(transactionBlob string, signData string, publicKey string) (string, Error) {
	if signData == "" {
		return "", sdkErr(INVALID_SIGNDATA)
	}
	if transactionBlob == "" {
		return "", sdkErr(INVALID_TRANSACTIONBLOB)
	}
	if publicKey == "" {
		return "", sdkErr(INVALID_PUBLICKEY)
	}
	request := make(map[string]interface{})
	items := make([]map[string]interface{}, 1)
	items[0] = make(map[string]interface{})
	signatures := make([]map[string]string, 1)
	signatures[0] = make(map[string]string)
	items[0]["transaction_blob"] = transactionBlob
	items[0]["signatures"] = signatures
	signatures[0]["sign_data"] = signData
	signatures[0]["public_key"] = publicKey
	request["items"] = items
	deal_js, err := json.Marshal(request)
	if err != nil {
		Err.Code = JSON_MARSHAL_ERROR
		Err.Err = err
		return "", Err
	}
	str := "/submitTransaction"
	var buf bytes.Buffer
	buf.WriteString(bumosdk.Account.url)
	buf.WriteString(str)
	url := buf.String()
	client := &http.Client{}
	reqest, err := http.NewRequest("POST", url, bytes.NewReader(deal_js))
	if err != nil {
		Err.Code = HTTP_NEWREQUEST_ERROR
		Err.Err = err
		return "", Err
	}
	response, err := client.Do(reqest)
	if err != nil {
		Err.Code = CLIENT_DO_ERROR
		Err.Err = err
		return "", Err
	}
	if response.StatusCode == 200 {
		data := make(map[string]interface{})
		decoder := json.NewDecoder(response.Body)
		decoder.UseNumber()
		err = decoder.Decode(&data)
		if err != nil {
			Err.Code = DECODER_DECODE_ERROR
			Err.Err = err
			return "", Err
		}
		results := data["results"].([]interface{})
		result := results[0].(map[string]interface{})
		if result["error_code"].(json.Number) == "0" {
			hash := make(map[string]interface{})
			hash["hash"] = result["hash"]
			Mdata, err := json.Marshal(&hash)
			if err != nil {
				Err.Code = JSON_MARSHAL_ERROR
				Err.Err = err
				return "", Err
			}
			Err.Code = SUCCESS
			Err.Err = nil
			return string(Mdata), Err
		} else {
			errorCodejs := result["error_code"].(json.Number)
			errorCode, err := strconv.ParseInt(string(errorCodejs), 10, 64)
			if err != nil {
				Err.Code = STRCONV_PARSEINT_ERROR
				Err.Err = err
				return "", Err
			}
			Err.Code = int(float64(errorCode) + 10000)
			Err.Err = errors.New(result["error_desc"].(string))
			return "", Err
		}
	} else {
		Err.Code = response.StatusCode
		Err.Err = errors.New(response.Status)
		return "", Err
	}
}

//获取最新gasPrice
func (bumosdk *BumoSdk) getGasPrice() int64 {
	var buf bytes.Buffer
	get := "/getLedger?with_fee=true"
	buf.WriteString(bumosdk.Account.url)
	buf.WriteString(get)
	url := buf.String()
	client := &http.Client{}
	reqest, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0
	}
	response, err := client.Do(reqest)
	if err != nil {
		return 0
	}
	if response.StatusCode == 200 {
		data := make(map[string]interface{})
		decoder := json.NewDecoder(response.Body)
		decoder.UseNumber()
		err = decoder.Decode(&data)
		if err != nil {
			return 0
		}
		if data["error_code"].(json.Number) == "0" {
			result := data["result"].(map[string]interface{})
			fees := result["fees"].(map[string]interface{})
			gasPricejs := fees["gas_price"].(json.Number)
			gasPrice, err := strconv.ParseInt(string(gasPricejs), 10, 64)
			if err != nil {
				return 0
			}
			return gasPrice
		} else {
			return 0
		}
	} else {
		return 0
	}
}
