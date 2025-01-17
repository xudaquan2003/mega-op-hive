package test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"runtime"
	"strings"

	api "github.com/ethereum/go-ethereum/beacon/engine"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/ethereum/hive/simulators/ethereum/engine/client"
	"github.com/ethereum/hive/simulators/ethereum/engine/config/cancun"
	"github.com/ethereum/hive/simulators/ethereum/engine/globals"
	typ "github.com/ethereum/hive/simulators/ethereum/engine/types"
)

// Print the caller to this file
func PrintExpectStack(t *Env) {
	_, currentFile, _, _ := runtime.Caller(0)
	for i := 1; ; i++ {
		_, file, line, ok := runtime.Caller(i)
		if !ok {
			return
		}
		if file == currentFile {
			continue
		}
		fmt.Printf("DEBUG (%s): Failed `Expect*` routine called from: file=%s, line=%d\n", t.TestName, file, line)
		return
	}
}

type PayloadStatus string

const (
	Unknown          PayloadStatus = ""
	Valid                          = "VALID"
	Invalid                        = "INVALID"
	Syncing                        = "SYNCING"
	Accepted                       = "ACCEPTED"
	InvalidBlockHash               = "INVALID_BLOCK_HASH"
)

func StatusesToString(statuses []PayloadStatus) []string {
	result := make([]string, len(statuses))
	for i, s := range statuses {
		result[i] = string(s)
	}
	return result
}

// Test Engine API Helper Structs
type ExpectEnv struct {
	*Env
	ExpectationDescription string
}

func (e ExpectEnv) Fatalf(format string, values ...interface{}) {
	PrintExpectStack(e.Env)
	if e.ExpectationDescription != "" {
		e.Env.Logf("DEBUG (%s): %s", e.TestName, e.ExpectationDescription)
	}
	e.Env.Fatalf(format, values...)
}

type TestEngineClient struct {
	*ExpectEnv
	EngineAPIVersionResolver client.EngineAPIVersionResolver
	Engine                   client.EngineClient
}

func NewTestEngineClient(t *Env, ec client.EngineClient) *TestEngineClient {
	return &TestEngineClient{
		ExpectEnv:                &ExpectEnv{Env: t},
		EngineAPIVersionResolver: t.ForkConfig,
		Engine:                   ec,
	}
}

func (tec *TestEngineClient) WithEngineAPIVersionResolver(res client.EngineAPIVersionResolver) *TestEngineClient {
	newTestEngineClient := *tec
	newTestEngineClient.EngineAPIVersionResolver = res
	return &newTestEngineClient
}

// ForkchoiceUpdated

type ForkchoiceResponseExpectObject struct {
	*ExpectEnv
	Response  api.ForkChoiceResponse
	Version   int
	Error     error
	ErrorCode int
}

func (tec *TestEngineClient) TestEngineForkchoiceUpdatedV1(fcState *api.ForkchoiceStateV1, pAttributes *typ.PayloadAttributes) *ForkchoiceResponseExpectObject {
	ctx, cancel := context.WithTimeout(tec.TestContext, globals.RPCTimeout)
	defer cancel()
	resp, err := tec.Engine.ForkchoiceUpdatedV1(ctx, fcState, pAttributes)
	ret := &ForkchoiceResponseExpectObject{
		ExpectEnv: &ExpectEnv{Env: tec.Env},
		Response:  resp,
		Version:   1,
		Error:     err,
	}
	if err, ok := err.(rpc.Error); ok {
		ret.ErrorCode = err.ErrorCode()
	}
	return ret
}

func (tec *TestEngineClient) TestEngineForkchoiceUpdatedV2(fcState *api.ForkchoiceStateV1, pAttributes *typ.PayloadAttributes) *ForkchoiceResponseExpectObject {
	ctx, cancel := context.WithTimeout(tec.TestContext, globals.RPCTimeout)
	defer cancel()
	resp, err := tec.Engine.ForkchoiceUpdatedV2(ctx, fcState, pAttributes)
	ret := &ForkchoiceResponseExpectObject{
		ExpectEnv: &ExpectEnv{Env: tec.Env},
		Response:  resp,
		Version:   2,
		Error:     err,
	}
	if err, ok := err.(rpc.Error); ok {
		ret.ErrorCode = err.ErrorCode()
	}
	return ret
}

func (tec *TestEngineClient) TestEngineForkchoiceUpdated(fcState *api.ForkchoiceStateV1, pAttributes *typ.PayloadAttributes, headTimestamp uint64) *ForkchoiceResponseExpectObject {
	var pAttrTimestamp *uint64
	if pAttributes != nil {
		pAttrTimestamp = &pAttributes.Timestamp
	}
	version := tec.EngineAPIVersionResolver.ForkchoiceUpdatedVersion(headTimestamp, pAttrTimestamp)
	ctx, cancel := context.WithTimeout(tec.TestContext, globals.RPCTimeout)
	defer cancel()
	resp, err := tec.Engine.ForkchoiceUpdated(ctx, version, fcState, pAttributes)
	ret := &ForkchoiceResponseExpectObject{
		ExpectEnv: &ExpectEnv{Env: tec.Env},
		Response:  resp,
		Version:   version,
		Error:     err,
	}
	if err, ok := err.(rpc.Error); ok {
		ret.ErrorCode = err.ErrorCode()
	}
	return ret
}

func (exp *ForkchoiceResponseExpectObject) ExpectNoError() {
	if exp.Error != nil {
		exp.Fatalf("FAIL (%s): Unexpected error on EngineForkchoiceUpdatedV%d: %v, expected=<None>", exp.TestName, exp.Version, exp.Error)
	}
}

func (exp *ForkchoiceResponseExpectObject) ExpectNoValidationError() {
	if exp.Response.PayloadStatus.ValidationError != nil {
		exp.Fatalf("FAIL (%s): Unexpected validation error on EngineForkchoiceUpdatedV%d: %v, expected=<None>", exp.TestName, exp.Version, exp.Response.PayloadStatus.ValidationError)
	}
}

func (exp *ForkchoiceResponseExpectObject) ExpectError() {
	if exp.Error == nil {
		exp.Fatalf("FAIL (%s): Expected error on EngineForkchoiceUpdatedV%d: response=%v", exp.TestName, exp.Version, exp.Response)
	}
}

func (exp *ForkchoiceResponseExpectObject) ExpectErrorCode(code int) {
	exp.ExpectError()
	if exp.ErrorCode != code {
		exp.Fatalf("FAIL (%s): Expected error code on EngineForkchoiceUpdatedV%d: want=%d, got=%d", exp.TestName, exp.Version, code, exp.ErrorCode)
	}
}

func (exp *ForkchoiceResponseExpectObject) ExpectPayloadStatus(ps PayloadStatus) {
	exp.ExpectNoError()
	if PayloadStatus(exp.Response.PayloadStatus.Status) != ps {
		exp.Fatalf("FAIL (%s): Unexpected status response on EngineForkchoiceUpdatedV%d: %v, expected=%v", exp.TestName, exp.Version, exp.Response.PayloadStatus.Status, ps)
	}
}

func (exp *ForkchoiceResponseExpectObject) ExpectAnyPayloadStatus(statuses ...PayloadStatus) {
	exp.ExpectNoError()
	for _, status := range statuses {
		if PayloadStatus(exp.Response.PayloadStatus.Status) == status {
			return
		}
	}
	exp.Fatalf("FAIL (%s): Unexpected status response on EngineForkchoiceUpdatedV%d: %v, expected=%v", exp.TestName, exp.Version, exp.Response.PayloadStatus.Status, StatusesToString(statuses))
}

func (exp *ForkchoiceResponseExpectObject) ExpectLatestValidHash(lvh *common.Hash) {
	exp.ExpectNoError()
	if ((lvh == nil || exp.Response.PayloadStatus.LatestValidHash == nil) && exp.Response.PayloadStatus.LatestValidHash != lvh) ||
		(lvh != nil && exp.Response.PayloadStatus.LatestValidHash != nil && *exp.Response.PayloadStatus.LatestValidHash != *lvh) {
		exp.Fatalf("FAIL (%v): Unexpected LatestValidHash on EngineForkchoiceUpdatedV%d: %v, expected=%v", exp.TestName, exp.Version, exp.Response.PayloadStatus.LatestValidHash, lvh)
	}
}

func (exp *ForkchoiceResponseExpectObject) ExpectPayloadID(pid *api.PayloadID) {
	exp.ExpectNoError()
	if ((exp.Response.PayloadID == nil || pid == nil) && exp.Response.PayloadID != pid) ||
		(exp.Response.PayloadID != nil && pid != nil && *exp.Response.PayloadID != *pid) {
		exp.Fatalf("FAIL (%v): Unexpected PayloadID on EngineForkchoiceUpdatedV%d: %v, expected=%v", exp.TestName, exp.Version, exp.Response.PayloadID, pid)
	}
}

func (exp *ForkchoiceResponseExpectObject) ExpectUpdatedPayloadID(previousID *api.PayloadID) {
	exp.ExpectNoError()
	if exp.Response.PayloadID == nil || previousID == nil {
		if exp.Response.PayloadID == previousID {
			// Both are null
			exp.Fatalf("FAIL (%v): Unexpected PayloadID on EngineForkchoiceUpdatedV%d: Expected change from %v", exp.TestName, exp.Version, previousID)
		}
	} else {
		// Both are different from null
		if *exp.Response.PayloadID == *previousID {
			exp.Fatalf("FAIL (%v): Unexpected PayloadID on EngineForkchoiceUpdatedV%d: Expected change from %s", exp.TestName, exp.Version, previousID.String())
		}
	}
}

// NewPayload

type NewPayloadResponseExpectObject struct {
	*ExpectEnv
	Payload   *typ.ExecutableData
	Status    api.PayloadStatusV1
	Version   int
	Error     error
	ErrorCode int
}

func (tec *TestEngineClient) TestEngineNewPayloadV1(payload *typ.ExecutableData) *NewPayloadResponseExpectObject {
	ctx, cancel := context.WithTimeout(tec.TestContext, globals.RPCTimeout)
	defer cancel()
	status, err := tec.Engine.NewPayloadV1(ctx, payload)
	ret := &NewPayloadResponseExpectObject{
		ExpectEnv: &ExpectEnv{Env: tec.Env},
		Payload:   payload,
		Status:    status,
		Version:   1,
		Error:     err,
	}
	if err, ok := err.(rpc.Error); ok {
		ret.ErrorCode = err.ErrorCode()
	}
	return ret
}

func (tec *TestEngineClient) TestEngineNewPayloadV2(payload *typ.ExecutableData) *NewPayloadResponseExpectObject {
	ctx, cancel := context.WithTimeout(tec.TestContext, globals.RPCTimeout)
	defer cancel()
	status, err := tec.Engine.NewPayloadV2(ctx, payload)
	ret := &NewPayloadResponseExpectObject{
		ExpectEnv: &ExpectEnv{Env: tec.Env},
		Payload:   payload,
		Status:    status,
		Version:   2,
		Error:     err,
	}
	if err, ok := err.(rpc.Error); ok {
		ret.ErrorCode = err.ErrorCode()
	}
	return ret
}

func (tec *TestEngineClient) TestEngineNewPayloadV3(payload *typ.ExecutableData) *NewPayloadResponseExpectObject {
	ctx, cancel := context.WithTimeout(tec.TestContext, globals.RPCTimeout)
	defer cancel()
	status, err := tec.Engine.NewPayloadV3(ctx, payload)
	ret := &NewPayloadResponseExpectObject{
		ExpectEnv: &ExpectEnv{Env: tec.Env},
		Payload:   payload,
		Status:    status,
		Version:   3,
		Error:     err,
	}
	if err, ok := err.(rpc.Error); ok {
		ret.ErrorCode = err.ErrorCode()
	}
	return ret
}

func (tec *TestEngineClient) TestEngineNewPayload(payload *typ.ExecutableData) *NewPayloadResponseExpectObject {
	if payload == nil {
		panic("Payload is nil")
	}
	version := tec.EngineAPIVersionResolver.NewPayloadVersion(payload.Timestamp)
	if version == 3 {
		return tec.TestEngineNewPayloadV3(payload)
	} else if version == 2 {
		return tec.TestEngineNewPayloadV2(payload)
	} else if version == 1 {
		return tec.TestEngineNewPayloadV1(payload)
	}
	panic(fmt.Sprintf("Unsupported version: %d", version))
}

func (exp *NewPayloadResponseExpectObject) PayloadJson() string {
	jsonPayload, _ := json.MarshalIndent(exp.Payload, "", " ")
	return string(jsonPayload)
}

func (exp *NewPayloadResponseExpectObject) ExpectNoError() {
	if exp.Error != nil {
		exp.Fatalf("FAIL (%s): Expected no error on EngineNewPayloadV%d: error=%v, payload=%s", exp.TestName, exp.Version, exp.Error, exp.PayloadJson())
	}
}

func (exp *NewPayloadResponseExpectObject) ExpectError() {
	if exp.Error == nil {
		exp.Fatalf("FAIL (%s): Expected error on EngineNewPayloadV%d: status=%v, payload=%s", exp.TestName, exp.Version, exp.Status, exp.PayloadJson())
	}
}

func (exp *NewPayloadResponseExpectObject) ExpectErrorCode(code int) {
	exp.ExpectError()
	if exp.ErrorCode != code {
		exp.Fatalf("FAIL (%s): Expected error code on EngineNewPayloadV%d: want=%d, got=%d", exp.TestName, exp.Version, code, exp.ErrorCode)
	}
}

func (exp *NewPayloadResponseExpectObject) ExpectNoValidationError() {
	if exp.Status.ValidationError != nil {
		exp.Fatalf("FAIL (%s): Unexpected validation error on EngineNewPayloadV%d: %v, expected=<None>", exp.TestName, exp.Version, exp.Status.ValidationError)
	}
}

func (exp *NewPayloadResponseExpectObject) ExpectStatus(ps PayloadStatus) {
	exp.ExpectNoError()
	if PayloadStatus(exp.Status.Status) != ps {
		exp.Fatalf("FAIL (%s): Unexpected status response on EngineNewPayloadV%d: %v, expected=%v, payload=%s", exp.TestName, exp.Version, exp.Status.Status, ps, exp.PayloadJson())
	}
}

func (exp *NewPayloadResponseExpectObject) ExpectStatusEither(statuses ...PayloadStatus) {
	exp.ExpectNoError()
	for _, status := range statuses {
		if PayloadStatus(exp.Status.Status) == status {
			return
		}
	}

	exp.Fatalf("FAIL (%s): Unexpected status response on EngineNewPayloadV%d: %v, expected=%v, payload=%s", exp.TestName, exp.Version, exp.Status.Status, strings.Join(StatusesToString(statuses), ","), exp.PayloadJson())
}

func (exp *NewPayloadResponseExpectObject) ExpectLatestValidHash(lvh *common.Hash) {
	exp.ExpectNoError()
	if ((lvh == nil || exp.Status.LatestValidHash == nil) && exp.Status.LatestValidHash != lvh) ||
		(lvh != nil && exp.Status.LatestValidHash != nil && *exp.Status.LatestValidHash != *lvh) {
		exp.Fatalf("FAIL (%v): Unexpected LatestValidHash on EngineNewPayloadV%d: %v, expected=%v", exp.TestName, exp.Version, exp.Status.LatestValidHash, lvh)
	}
}

// GetPayload
type GetPayloadResponseExpectObject struct {
	*ExpectEnv
	Payload               typ.ExecutableData
	BlockValue            *big.Int
	BlobsBundle           *typ.BlobsBundle
	ShouldOverrideBuilder *bool
	Version               int
	Error                 error
	ErrorCode             int
}

func (tec *TestEngineClient) TestEngineGetPayloadV1(payloadID *api.PayloadID) *GetPayloadResponseExpectObject {
	ctx, cancel := context.WithTimeout(tec.TestContext, globals.RPCTimeout)
	defer cancel()
	payload, err := tec.Engine.GetPayloadV1(ctx, payloadID)
	ret := &GetPayloadResponseExpectObject{
		ExpectEnv:  &ExpectEnv{Env: tec.Env},
		Payload:    payload,
		Version:    1,
		BlockValue: nil,
		Error:      err,
	}
	if err, ok := err.(rpc.Error); ok {
		ret.ErrorCode = err.ErrorCode()
	}
	return ret
}

func (tec *TestEngineClient) TestEngineGetPayloadV2(payloadID *api.PayloadID) *GetPayloadResponseExpectObject {
	ctx, cancel := context.WithTimeout(tec.TestContext, globals.RPCTimeout)
	defer cancel()
	payload, blockValue, err := tec.Engine.GetPayloadV2(ctx, payloadID)
	ret := &GetPayloadResponseExpectObject{
		ExpectEnv:  &ExpectEnv{Env: tec.Env},
		Payload:    payload,
		Version:    2,
		BlockValue: blockValue,
		Error:      err,
	}
	if err, ok := err.(rpc.Error); ok {
		ret.ErrorCode = err.ErrorCode()
	}
	return ret
}

func (tec *TestEngineClient) TestEngineGetPayloadV3(payloadID *api.PayloadID) *GetPayloadResponseExpectObject {
	ctx, cancel := context.WithTimeout(tec.TestContext, globals.RPCTimeout)
	defer cancel()
	payload, blockValue, blobsBundle, shouldOverride, err := tec.Engine.GetPayloadV3(ctx, payloadID)
	ret := &GetPayloadResponseExpectObject{
		ExpectEnv:             &ExpectEnv{Env: tec.Env},
		Payload:               payload,
		Version:               3,
		BlockValue:            blockValue,
		BlobsBundle:           blobsBundle,
		ShouldOverrideBuilder: shouldOverride,
		Error:                 err,
	}
	if err, ok := err.(rpc.Error); ok {
		ret.ErrorCode = err.ErrorCode()
	}
	return ret
}

func (tec *TestEngineClient) TestEngineGetPayload(payloadID *api.PayloadID, payloadAttributes *typ.PayloadAttributes) *GetPayloadResponseExpectObject {
	version := tec.EngineAPIVersionResolver.GetPayloadVersion(payloadAttributes.Timestamp)
	ctx, cancel := context.WithTimeout(tec.TestContext, globals.RPCTimeout)
	defer cancel()
	payload, blockValue, blobBundle, shouldOverride, err := tec.Engine.GetPayload(ctx, version, payloadID)
	if blobBundle != nil {
		payload.VersionedHashes, err = blobBundle.VersionedHashes(cancun.BLOB_COMMITMENT_VERSION_KZG)
		if err != nil {
			panic(err)
		}
	}
	payload.ParentBeaconBlockRoot = payloadAttributes.BeaconRoot
	// TODO: Where do we get the beacon root from ???
	ret := &GetPayloadResponseExpectObject{
		ExpectEnv:             &ExpectEnv{Env: tec.Env},
		Payload:               payload,
		Version:               version,
		BlockValue:            blockValue,
		BlobsBundle:           blobBundle,
		ShouldOverrideBuilder: shouldOverride,
		Error:                 err,
	}
	if err, ok := err.(rpc.Error); ok {
		ret.ErrorCode = err.ErrorCode()
	}

	return ret
}

func (exp *GetPayloadResponseExpectObject) ExpectNoError() {
	if exp.Error != nil {
		exp.Fatalf("FAIL (%s): Expected no error on EngineGetPayloadV%d: error=%v", exp.TestName, exp.Version, exp.Error)
	}
}

func (exp *GetPayloadResponseExpectObject) ExpectError() {
	if exp.Error == nil {
		exp.Fatalf("FAIL (%s): Expected error on EngineGetPayloadV%d: payload=%v", exp.TestName, exp.Version, exp.Payload)
	}
}

func (exp *GetPayloadResponseExpectObject) ExpectErrorCode(code int) {
	exp.ExpectError()
	if exp.ErrorCode != code {
		exp.Fatalf("FAIL (%s): Expected error code on EngineGetPayloadV%d: want=%d, got=%d", exp.TestName, exp.Version, code, exp.ErrorCode)
	}
}

func ComparePayloads(want *typ.ExecutableData, got *typ.ExecutableData) error {
	if want == nil || got == nil {
		if want == nil && got == nil {
			return nil
		}
		return fmt.Errorf("want: %v, got: %v", want, got)
	}

	if !bytes.Equal(want.ParentHash[:], got.ParentHash[:]) {
		return fmt.Errorf("unexpected ParentHash: want=%v, got=%v", want.ParentHash, got.ParentHash)
	}

	if !bytes.Equal(want.FeeRecipient[:], got.FeeRecipient[:]) {
		return fmt.Errorf("unexpected FeeRecipient: want=%v, got=%v", want.FeeRecipient, got.FeeRecipient)
	}

	if !bytes.Equal(want.StateRoot[:], got.StateRoot[:]) {
		return fmt.Errorf("unexpected StateRoot: want=%v, got=%v", want.StateRoot, got.StateRoot)
	}

	if !bytes.Equal(want.ReceiptsRoot[:], got.ReceiptsRoot[:]) {
		return fmt.Errorf("unexpected ReceiptsRoot: want=%v, got=%v", want.ReceiptsRoot, got.ReceiptsRoot)
	}

	if !bytes.Equal(want.Random[:], got.Random[:]) {
		return fmt.Errorf("unexpected Random: want=%v, got=%v", want.Random, got.Random)
	}

	if !bytes.Equal(want.BlockHash[:], got.BlockHash[:]) {
		return fmt.Errorf("unexpected BlockHash: want=%v, got=%v", want.BlockHash, got.BlockHash)
	}

	if !bytes.Equal(want.LogsBloom, got.LogsBloom) {
		return fmt.Errorf("unexpected LogsBloom: want=%v, got=%v", want.LogsBloom, got.LogsBloom)
	}

	if !bytes.Equal(want.ExtraData, got.ExtraData) {
		return fmt.Errorf("unexpected ExtraData: want=%v, got=%v", want.ExtraData, got.ExtraData)
	}

	if want.Number != got.Number {
		return fmt.Errorf("unexpected Number: want=%d, got=%d", want.Number, got.Number)
	}

	if want.GasLimit != got.GasLimit {
		return fmt.Errorf("unexpected GasLimit: want=%d, got=%d", want.GasLimit, got.GasLimit)
	}

	if want.GasUsed != got.GasUsed {
		return fmt.Errorf("unexpected GasUsed: want=%d, got=%d", want.GasUsed, got.GasUsed)
	}

	if want.Timestamp != got.Timestamp {
		return fmt.Errorf("unexpected Timestamp: want=%d, got=%d", want.Timestamp, got.Timestamp)
	}

	if want.BaseFeePerGas.Cmp(got.BaseFeePerGas) != 0 {
		return fmt.Errorf("unexpected BaseFeePerGas: want=%d, got=%d", want.BaseFeePerGas, got.BaseFeePerGas)
	}

	if err := CompareTransactions(want.Transactions, got.Transactions); err != nil {
		return err
	}

	if err := CompareWithdrawals(want.Withdrawals, got.Withdrawals); err != nil {
		return err
	}

	return nil
}

func (exp *GetPayloadResponseExpectObject) ExpectPayload(expectedPayload *typ.ExecutableData) {
	exp.ExpectNoError()
	if err := ComparePayloads(expectedPayload, &exp.Payload); err != nil {
		exp.Fatalf("FAIL (%s): Unexpected payload returned on EngineGetPayloadV%d: %v", exp.TestName, exp.Version, err)
	}
}

func (exp *GetPayloadResponseExpectObject) ExpectPayloadParentHash(expectedParentHash common.Hash) {
	exp.ExpectNoError()
	if exp.Payload.ParentHash != expectedParentHash {
		exp.Fatalf("FAIL (%s): Unexpected parent hash for payload on EngineGetPayloadV%d: %v, expected=%v", exp.TestName, exp.Version, exp.Payload.ParentHash, expectedParentHash)
	}
}

func (exp *GetPayloadResponseExpectObject) ExpectPayloadBlockNumber(expectedBlockNumber uint64) {
	exp.ExpectNoError()
	if exp.Payload.Number != expectedBlockNumber {
		exp.Fatalf("FAIL (%s): Unexpected block number for payload on EngineGetPayloadV%d: %v, expected=%v", exp.TestName, exp.Version, exp.Payload.Number, expectedBlockNumber)
	}
}

func (exp *GetPayloadResponseExpectObject) ExpectPrevRandao(expectedPrevRandao common.Hash) {
	exp.ExpectNoError()
	if exp.Payload.Random != expectedPrevRandao {
		exp.Fatalf("FAIL (%s): Unexpected prevRandao for payload on EngineGetPayloadV%d: %v, expected=%v", exp.TestName, exp.Version, exp.Payload.Random, expectedPrevRandao)
	}
}

func (exp *GetPayloadResponseExpectObject) ExpectTimestamp(expectedTimestamp uint64) {
	exp.ExpectNoError()
	if exp.Payload.Timestamp != expectedTimestamp {
		exp.Fatalf("FAIL (%s): Unexpected timestamp for payload on EngineGetPayloadV%d: %v, expected=%v", exp.TestName, exp.Version, exp.Payload.Timestamp, expectedTimestamp)
	}
}

// GetPayloadBodies
type GetPayloadBodiesResponseExpectObject struct {
	*ExpectEnv
	PayloadBodies []*typ.ExecutionPayloadBodyV1
	BlockValue    *big.Int
	Version       int
	Error         error
	ErrorCode     int
}

func (tec *TestEngineClient) TestEngineGetPayloadBodiesByRangeV1(start uint64, count uint64) *GetPayloadBodiesResponseExpectObject {
	// Get Payload Bodies can take a long time to respond
	ctx, cancel := context.WithTimeout(tec.TestContext, globals.RPCTimeout*6)
	defer cancel()
	payloadBodies, err := tec.Engine.GetPayloadBodiesByRangeV1(ctx, start, count)
	ret := &GetPayloadBodiesResponseExpectObject{
		ExpectEnv:     &ExpectEnv{Env: tec.Env},
		PayloadBodies: payloadBodies,
		Version:       1,
		BlockValue:    nil,
		Error:         err,
	}
	if err, ok := err.(rpc.Error); ok {
		ret.ErrorCode = err.ErrorCode()
	}
	return ret
}

func (tec *TestEngineClient) TestEngineGetPayloadBodiesByHashV1(hashes []common.Hash) *GetPayloadBodiesResponseExpectObject {
	ctx, cancel := context.WithTimeout(tec.TestContext, globals.RPCTimeout*6)
	defer cancel()
	payloadBodies, err := tec.Engine.GetPayloadBodiesByHashV1(ctx, hashes)
	ret := &GetPayloadBodiesResponseExpectObject{
		ExpectEnv:     &ExpectEnv{Env: tec.Env},
		PayloadBodies: payloadBodies,
		Version:       1,
		BlockValue:    nil,
		Error:         err,
	}
	if err, ok := err.(rpc.Error); ok {
		ret.ErrorCode = err.ErrorCode()
	}
	return ret
}

func (exp *GetPayloadBodiesResponseExpectObject) ExpectNoError() {
	if exp.Error != nil {
		exp.Fatalf("FAIL (%s): Expected no error on EngineGetPayloadBodiesV%d: error=%v", exp.TestName, exp.Version, exp.Error)
	}
}

func (exp *GetPayloadBodiesResponseExpectObject) ExpectError() {
	if exp.Error == nil {
		exp.Fatalf("FAIL (%s): Expected error on EngineGetPayloadBodiesV%d: payload=%v", exp.TestName, exp.Version, exp.PayloadBodies)
	}
}

func (exp *GetPayloadBodiesResponseExpectObject) ExpectErrorCode(code int) {
	exp.ExpectError()
	if exp.ErrorCode != code {
		exp.Fatalf("FAIL (%s): Expected error code on EngineGetPayloadBodiesV%d: want=%d, got=%d", exp.TestName, exp.Version, code, exp.ErrorCode)
	}
}

func (exp *GetPayloadBodiesResponseExpectObject) ExpectPayloadBodiesCount(count uint64) {
	exp.ExpectNoError()
	if exp.PayloadBodies == nil {
		exp.Fatalf("FAIL (%s): Expected payload bodies list count on EngineGetPayloadBodiesV%d: want=%d, got=nil", exp.TestName, exp.Version, count)
	}
	if uint64(len(exp.PayloadBodies)) != count {
		exp.Fatalf("FAIL (%s): Expected payload bodies list count on EngineGetPayloadBodiesV%d: want=%d, got=%d", exp.TestName, exp.Version, count, len(exp.PayloadBodies))
	}
}

func CompareTransactions(want [][]byte, got [][]byte) error {
	if len(want) != len(got) {
		return fmt.Errorf("incorrect tx length: want=%d, got=%d", len(want), len(got))
	}

	for i, a_tx := range want {
		b_tx := got[i]
		if !bytes.Equal(a_tx, b_tx) {
			return fmt.Errorf("tx %d not equal: want=%x, got=%x", i, a_tx, b_tx)
		}
	}

	return nil
}

func CompareWithdrawal(want *types.Withdrawal, got *types.Withdrawal) error {
	if want == nil || got == nil {
		if want == nil && got != nil {
			got, _ := json.MarshalIndent(got, "", " ")
			return fmt.Errorf("want=null, got=%s", got)
		} else if want != nil && got == nil {
			want, _ := json.MarshalIndent(want, "", " ")
			return fmt.Errorf("want=%s, got=null", want)
		}
		return nil
	}
	if want.Amount != got.Amount ||
		!bytes.Equal(want.Address[:], got.Address[:]) ||
		want.Index != got.Index ||
		want.Validator != got.Validator {
		want, _ := json.MarshalIndent(want, "", " ")
		got, _ := json.MarshalIndent(got, "", " ")
		return fmt.Errorf("want=%s, got=%s", want, got)
	}
	return nil
}

func CompareWithdrawals(want []*types.Withdrawal, got []*types.Withdrawal) error {
	if want == nil || got == nil {
		if want == nil && got == nil {
			return nil
		}
		if want == nil && got != nil {
			got, _ := json.MarshalIndent(got, "", " ")
			return fmt.Errorf("incorrect withdrawals: want: null, got: %s", got)
		} else {
			want, _ := json.MarshalIndent(want, "", " ")
			return fmt.Errorf("incorrect withdrawals: want: %s, got: null", want)
		}

	}

	if len(want) != len(got) {
		return fmt.Errorf("incorrect withdrawals length: want=%d, got=%d", len(want), len(got))
	}

	for i, a_w := range want {
		b_w := got[i]
		if err := CompareWithdrawal(a_w, b_w); err != nil {
			return fmt.Errorf("withdrawal %d not equal: %v", i, err)
		}
	}

	return nil
}

func ComparePayloadBodies(want *typ.ExecutionPayloadBodyV1, got *typ.ExecutionPayloadBodyV1) error {
	if want == nil || got == nil {
		if want == nil && got == nil {
			return nil
		}
		return fmt.Errorf("want: %v, got: %v", want, got)
	}

	if err := CompareTransactions(want.Transactions, got.Transactions); err != nil {
		return err
	}

	if err := CompareWithdrawals(want.Withdrawals, got.Withdrawals); err != nil {
		return err
	}

	return nil
}

func (exp *GetPayloadBodiesResponseExpectObject) ExpectPayloadBody(index uint64, payloadBody *typ.ExecutionPayloadBodyV1) {
	exp.ExpectNoError()
	if exp.PayloadBodies == nil {
		exp.Fatalf("FAIL (%s): Expected payload body in list on EngineGetPayloadBodiesV%d, but list is nil", exp.TestName, exp.Version)
	}
	if index >= uint64(len(exp.PayloadBodies)) {
		exp.Fatalf("FAIL (%s): Expected payload body in list on EngineGetPayloadBodiesV%d, list is smaller than index: want=%s", exp.TestName, exp.Version, index)
	}

	checkItem := exp.PayloadBodies[index]

	if err := ComparePayloadBodies(payloadBody, checkItem); err != nil {
		exp.Fatalf("FAIL (%s): Unexpected payload body on EngineGetPayloadBodiesV%d at index %d: %v", exp.TestName, exp.Version, index, err)

	}
}

// BlockNumber
type BlockNumberResponseExpectObject struct {
	*ExpectEnv
	Call      string
	Number    uint64
	Error     error
	ErrorCode int
}

func (tec *TestEngineClient) TestBlockNumber() *BlockNumberResponseExpectObject {
	ctx, cancel := context.WithTimeout(tec.TestContext, globals.RPCTimeout)
	defer cancel()
	number, err := tec.Engine.BlockNumber(ctx)
	ret := &BlockNumberResponseExpectObject{
		ExpectEnv: &ExpectEnv{Env: tec.Env},
		Call:      "BlockNumber",
		Number:    number,
		Error:     err,
	}
	if err, ok := err.(rpc.Error); ok {
		ret.ErrorCode = err.ErrorCode()
	}
	return ret
}

func (exp *BlockNumberResponseExpectObject) ExpectNoError() {
	if exp.Error != nil {
		exp.Fatalf("FAIL (%s): Unexpected error on %s: %v, expected=<None>", exp.TestName, exp.Call, exp.Error)
	}
}

func (exp *BlockNumberResponseExpectObject) ExpectNumber(number uint64) {
	exp.ExpectNoError()
	if exp.Number != number {
		exp.Fatalf("FAIL (%s): Unexpected block number on %s: %d, expected=%d", exp.TestName, exp.Call, exp.Number, number)
	}
}

// Header

type HeaderResponseExpectObject struct {
	*ExpectEnv
	Call      string
	Header    *types.Header
	Error     error
	ErrorCode int
}

func (tec *TestEngineClient) TestHeaderByNumber(number *big.Int) *HeaderResponseExpectObject {
	ctx, cancel := context.WithTimeout(tec.TestContext, globals.RPCTimeout)
	defer cancel()
	header, err := tec.Engine.HeaderByNumber(ctx, number)
	ret := &HeaderResponseExpectObject{
		ExpectEnv: &ExpectEnv{Env: tec.Env},
		Call:      "HeaderByNumber",
		Header:    header,
		Error:     err,
	}
	if err, ok := err.(rpc.Error); ok {
		ret.ErrorCode = err.ErrorCode()
	}
	return ret
}

func (tec *TestEngineClient) TestHeaderByHash(hash common.Hash) *HeaderResponseExpectObject {
	ctx, cancel := context.WithTimeout(tec.TestContext, globals.RPCTimeout)
	defer cancel()
	header, err := tec.Engine.HeaderByHash(ctx, hash)
	ret := &HeaderResponseExpectObject{
		ExpectEnv: &ExpectEnv{Env: tec.Env},
		Call:      "HeaderByNumber",
		Header:    header,
		Error:     err,
	}
	if err, ok := err.(rpc.Error); ok {
		ret.ErrorCode = err.ErrorCode()
	}
	return ret
}

func (exp *HeaderResponseExpectObject) ExpectNoError() {
	if exp.Error != nil {
		exp.Fatalf("FAIL (%s): Unexpected error on %s: %v, expected=<None>", exp.TestName, exp.Call, exp.Error)
	}
}

func (exp *HeaderResponseExpectObject) ExpectError() {
	if exp.Error == nil {
		exp.Fatalf("FAIL (%s): Expected error on %s: header=%v", exp.TestName, exp.Call, exp.Header)
	}
}

func (exp *HeaderResponseExpectObject) ExpectErrorCode(code int) {
	exp.ExpectError()
	if exp.ErrorCode != code {
		exp.Fatalf("FAIL (%s): Expected error code on %s: want=%d, got=%d", exp.TestName, exp.Call, code, exp.ErrorCode)
	}
}

func (exp *HeaderResponseExpectObject) ExpectHash(expHash common.Hash) {
	exp.ExpectNoError()
	if exp.Header.Hash() != expHash {
		exp.Fatalf("FAIL (%s): Unexpected hash on %s: %v, expected=%v", exp.TestName, exp.Call, exp.Header.Hash(), expHash)
	}
}

func (exp *HeaderResponseExpectObject) ExpectWithdrawalsRoot(expectedRoot *common.Hash) {
	exp.ExpectNoError()
	actualWithdrawalsRoot := exp.Header.WithdrawalsHash
	if ((expectedRoot == nil || actualWithdrawalsRoot == nil) && actualWithdrawalsRoot != expectedRoot) ||
		(expectedRoot != nil && actualWithdrawalsRoot != nil && *actualWithdrawalsRoot != *expectedRoot) {
		exp.Fatalf("FAIL (%s): Unexpected WithdrawalsRoot on %s: %v, expected=%v", exp.TestName, exp.Call, actualWithdrawalsRoot, expectedRoot)
	}
}

// Block

type BlockResponseExpectObject struct {
	*ExpectEnv
	Call      string
	Block     *types.Block
	Error     error
	ErrorCode int
}

func (tec *TestEngineClient) TestBlockByNumber(number *big.Int) *BlockResponseExpectObject {
	ctx, cancel := context.WithTimeout(tec.TestContext, globals.RPCTimeout)
	defer cancel()
	block, err := tec.Engine.BlockByNumber(ctx, number)
	ret := &BlockResponseExpectObject{
		ExpectEnv: &ExpectEnv{Env: tec.Env},
		Call:      "BlockByNumber",
		Block:     block,
		Error:     err,
	}
	if err, ok := err.(rpc.Error); ok {
		ret.ErrorCode = err.ErrorCode()
	}
	return ret
}

func (tec *TestEngineClient) TestBlockByHash(hash common.Hash) *BlockResponseExpectObject {
	ctx, cancel := context.WithTimeout(tec.TestContext, globals.RPCTimeout)
	defer cancel()
	block, err := tec.Engine.BlockByHash(ctx, hash)
	ret := &BlockResponseExpectObject{
		ExpectEnv: &ExpectEnv{Env: tec.Env},
		Call:      "BlockByHash",
		Block:     block,
		Error:     err,
	}
	if err, ok := err.(rpc.Error); ok {
		ret.ErrorCode = err.ErrorCode()
	}
	return ret
}

func (exp *BlockResponseExpectObject) ExpectError() {
	if exp.Error == nil {
		exp.Fatalf("FAIL (%s): Expected error on %s: block=%v", exp.TestName, exp.Call, exp.Block)
	}
}

func (exp *BlockResponseExpectObject) ExpectErrorCode(code int) {
	exp.ExpectError()
	if exp.ErrorCode != code {
		exp.Fatalf("FAIL (%s): Expected error code on %s: want=%d, got=%d", exp.TestName, exp.Call, code, exp.ErrorCode)
	}
}

func (exp *BlockResponseExpectObject) ExpectNoError() {
	if exp.Error != nil {
		exp.Fatalf("FAIL (%s): Unexpected error on %s: %v, expected=<None>", exp.TestName, exp.Call, exp.Error)
	}
}

func (exp *BlockResponseExpectObject) ExpectHash(expHash common.Hash) {
	exp.ExpectNoError()
	if exp.Block.Hash() != expHash {
		exp.Fatalf("FAIL (%s): Unexpected hash on %s: %v, expected=%v", exp.TestName, exp.Call, exp.Block.Hash(), expHash)
	}
}

func (exp *BlockResponseExpectObject) ExpectWithdrawalsRoot(expectedRoot *common.Hash) {
	exp.ExpectNoError()
	actualWithdrawalsRoot := exp.Block.Header().WithdrawalsHash
	if ((expectedRoot == nil || actualWithdrawalsRoot == nil) && actualWithdrawalsRoot != expectedRoot) ||
		(expectedRoot != nil && actualWithdrawalsRoot != nil && *actualWithdrawalsRoot != *expectedRoot) {
		exp.Fatalf("FAIL (%s): Unexpected WithdrawalsRoot on %s: %v, expected=%v", exp.TestName, exp.Call, actualWithdrawalsRoot, expectedRoot)
	}
}

func (exp *BlockResponseExpectObject) ExpectTransactionCountEqual(expTxCount int) {
	exp.ExpectNoError()
	if len(exp.Block.Transactions()) != expTxCount {
		exp.Fatalf("FAIL (%s): Unexpected transaction count on %s: %v, expected=%v", exp.TestName, exp.Call, len(exp.Block.Transactions()), expTxCount)
	}
}

func (exp *BlockResponseExpectObject) ExpectTransactionCountGreaterThan(expTxCount int) {
	exp.ExpectNoError()
	if len(exp.Block.Transactions()) <= expTxCount {
		exp.Fatalf("FAIL (%s): Unexpected transaction count on %s: %v, expected > %v", exp.TestName, exp.Call, len(exp.Block.Transactions()), expTxCount)
	}
}

func (exp *BlockResponseExpectObject) ExpectCoinbase(expCoinbase common.Address) {
	exp.ExpectNoError()
	if exp.Block.Coinbase() != expCoinbase {
		exp.Fatalf("FAIL (%s): Unexpected coinbase on %s: %v, expected=%v", exp.TestName, exp.Call, exp.Block.Coinbase(), expCoinbase)
	}
}

// Balance

type BalanceResponseExpectObject struct {
	*ExpectEnv
	Call      string
	Account   common.Address
	Block     *big.Int
	Balance   *big.Int
	Error     error
	ErrorCode int
}

func (tec *TestEngineClient) TestBalanceAt(account common.Address, number *big.Int) *BalanceResponseExpectObject {
	ctx, cancel := context.WithTimeout(tec.TestContext, globals.RPCTimeout)
	defer cancel()
	balance, err := tec.Engine.BalanceAt(ctx, account, number)
	ret := &BalanceResponseExpectObject{
		ExpectEnv: &ExpectEnv{Env: tec.Env},
		Call:      "BalanceAt",
		Account:   account,
		Balance:   balance,
		Block:     number,
		Error:     err,
	}
	if err, ok := err.(rpc.Error); ok {
		ret.ErrorCode = err.ErrorCode()
	}
	return ret
}

func (exp *BalanceResponseExpectObject) ExpectNoError() {
	if exp.Error != nil {
		exp.Fatalf("FAIL (%s): Unexpected error on %s: %v, expected=<None>", exp.TestName, exp.Call, exp.Error)
	}
}

func (exp *BalanceResponseExpectObject) ExpectBalanceEqual(expBalance *big.Int) {
	exp.Logf("INFO (%s): Testing balance for account %s on block %d: actual=%d, expected=%d",
		exp.TestName,
		exp.Account,
		exp.Block,
		exp.Balance,
		expBalance,
	)
	exp.ExpectNoError()
	if ((expBalance == nil || exp.Balance == nil) && expBalance != exp.Balance) ||
		(expBalance != nil && exp.Balance != nil && expBalance.Cmp(exp.Balance) != 0) {
		exp.Fatalf("FAIL (%s): Unexpected balance on %s, for account %s at block %v: %v, expected=%v", exp.TestName, exp.Call, exp.Account, exp.Block, exp.Balance, expBalance)
	}
}

// Storage

type StorageResponseExpectObject struct {
	*ExpectEnv
	Call      string
	Account   common.Address
	Key       common.Hash
	Number    *big.Int
	Storage   []byte
	Error     error
	ErrorCode int
}

func (tec *TestEngineClient) TestStorageAt(account common.Address, key common.Hash, number *big.Int) *StorageResponseExpectObject {
	ctx, cancel := context.WithTimeout(tec.TestContext, globals.RPCTimeout)
	defer cancel()
	storage, err := tec.Engine.StorageAt(ctx, account, key, number)
	ret := &StorageResponseExpectObject{
		ExpectEnv: &ExpectEnv{Env: tec.Env},
		Call:      "StorageAt",
		Account:   account,
		Key:       key,
		Number:    number,
		Storage:   storage,
		Error:     err,
	}
	if err, ok := err.(rpc.Error); ok {
		ret.ErrorCode = err.ErrorCode()
	}
	return ret
}

func (exp *StorageResponseExpectObject) ExpectNoError() {
	if exp.Error != nil {
		exp.Fatalf("FAIL (%s): Unexpected error on %s: %v, expected=<None>", exp.TestName, exp.Call, exp.Error)
	}
}

func (exp *StorageResponseExpectObject) ExpectBigIntStorageEqual(expBigInt *big.Int) {
	exp.ExpectNoError()
	bigInt := big.NewInt(0)
	bigInt.SetBytes(exp.Storage)
	if ((bigInt == nil || expBigInt == nil) && bigInt != expBigInt) ||
		(bigInt != nil && expBigInt != nil && bigInt.Cmp(expBigInt) != 0) {
		exp.Fatalf("FAIL (%s): Unexpected storage on %s (addr=%s, key=%s, block=%d): got=%d, expected=%d", exp.TestName, exp.Call, exp.Account, exp.Key, exp.Number, bigInt, expBigInt)
	}
}

func (exp *StorageResponseExpectObject) ExpectStorageEqual(expStorage common.Hash) {
	exp.ExpectNoError()
	if expStorage != common.BytesToHash(exp.Storage) {
		exp.Fatalf("FAIL (%s): Unexpected storage on %s: %v, expected=%v", exp.TestName, exp.Call, exp.Storage, expStorage)
	}
}

// Transaction Receipt
type TransactionReceiptExpectObject struct {
	*ExpectEnv
	Call      string
	Receipt   *types.Receipt
	Error     error
	ErrorCode int
}

func (tec *TestEngineClient) TestTransactionReceipt(txHash common.Hash) *TransactionReceiptExpectObject {
	ctx, cancel := context.WithTimeout(tec.TestContext, globals.RPCTimeout)
	defer cancel()
	receipt, err := tec.Engine.TransactionReceipt(ctx, txHash)
	ret := &TransactionReceiptExpectObject{
		ExpectEnv: &ExpectEnv{Env: tec.Env},
		Call:      "TransactionReceipt",
		Receipt:   receipt,
		Error:     err,
	}
	if err, ok := err.(rpc.Error); ok {
		ret.ErrorCode = err.ErrorCode()
	}
	return ret
}

func (tr *TransactionReceiptExpectObject) ExpectError() {
	if tr.Error == nil {
		tr.Fatalf("FAIL (%s): Expected error on %s: block=%v", tr.TestName, tr.Call, tr.Receipt)
	}
}

func (tr *TransactionReceiptExpectObject) ExpectErrorCode(code int) {
	tr.ExpectError()
	if tr.ErrorCode != code {
		tr.Fatalf("FAIL (%s): Expected error code on %s: want=%d, got=%d", tr.TestName, tr.Call, code, tr.ErrorCode)
	}
}

func (tr *TransactionReceiptExpectObject) ExpectNoError() {
	if tr.Error != nil {
		tr.Fatalf("FAIL (%s): Unexpected error on %s: %v, expected=<None>", tr.TestName, tr.Call, tr.Error)
	}
}

func (tr *TransactionReceiptExpectObject) ExpectTransactionHash(expectedHash common.Hash) {
	tr.ExpectNoError()
	if tr.Receipt.TxHash != expectedHash {
		tr.Fatalf("FAIL (%s): Unexpected transaction hash on %s: %v, expected=%v", tr.TestName, tr.Call, tr.Receipt.TxHash, expectedHash)
	}
}

func (tr *TransactionReceiptExpectObject) ExpectBlockHash(expectedHash common.Hash) {
	tr.ExpectNoError()
	if tr.Receipt.BlockHash != expectedHash {
		tr.Fatalf("FAIL (%s): Unexpected transaction block hash on %s: %v, blockhash=%v, expected=%v", tr.TestName, tr.Call, tr.Receipt.TxHash, tr.Receipt.BlockHash, expectedHash)
	}
}

func (tr *TransactionReceiptExpectObject) ExpectBlobGasUsed(expectedBlobGasUsed uint64) {
	tr.ExpectNoError()

	if expectedBlobGasUsed != tr.Receipt.BlobGasUsed {
		tr.Fatalf("FAIL (%s): Unexpected transaction blob gas used on %s: %v, blobGasUsed=%d, expected=%d", tr.TestName, tr.Call, tr.Receipt.TxHash, tr.Receipt.BlobGasUsed, expectedBlobGasUsed)
	}
}

func (tr *TransactionReceiptExpectObject) ExpectBlobGasPrice(expectedBlobGasPrice *big.Int) {
	tr.ExpectNoError()
	if (expectedBlobGasPrice == nil || tr.Receipt.BlobGasPrice == nil) && expectedBlobGasPrice != tr.Receipt.BlobGasPrice {
		tr.Fatalf("FAIL (%s): Unexpected transaction blob gas price on %s: %v, BlobGasPrice=%v, expected=%v", tr.TestName, tr.Call, tr.Receipt.TxHash, tr.Receipt.BlobGasPrice, expectedBlobGasPrice)
	} else if expectedBlobGasPrice != nil && tr.Receipt.BlobGasPrice != nil && expectedBlobGasPrice.Cmp(tr.Receipt.BlobGasPrice) != 0 {
		tr.Fatalf("FAIL (%s): Unexpected transaction blob gas price on %s: %v, BlobGasPrice=0x%x, expected=0x%x", tr.TestName, tr.Call, tr.Receipt.TxHash, *tr.Receipt.BlobGasPrice, *expectedBlobGasPrice)
	}
}
