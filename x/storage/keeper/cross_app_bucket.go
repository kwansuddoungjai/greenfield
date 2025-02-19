package keeper

import (
	"encoding/hex"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bnb-chain/greenfield/types/common"
	"github.com/bnb-chain/greenfield/x/storage/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

var _ sdk.CrossChainApplication = &BucketApp{}

type BucketApp struct {
	storageKeeper types.StorageKeeper
}

func NewBucketApp(keeper types.StorageKeeper) *BucketApp {
	return &BucketApp{
		storageKeeper: keeper,
	}
}

func (app *BucketApp) ExecuteAckPackage(ctx sdk.Context, appCtx *sdk.CrossChainAppContext, payload []byte) sdk.ExecuteResult {
	pack, err := types.DeserializeCrossChainPackage(payload, types.BucketChannelId, sdk.AckCrossChainPackageType)
	if err != nil {
		app.storageKeeper.Logger(ctx).Error("deserialize bucket cross chain package error", "payload", hex.EncodeToString(payload), "error", err.Error())
		panic("deserialize bucket cross chain package error")
	}

	var operationType uint8
	var result sdk.ExecuteResult
	switch p := pack.(type) {
	case *types.MirrorBucketAckPackage:
		operationType = types.OperationMirrorBucket
		result = app.handleMirrorBucketAckPackage(ctx, appCtx, p)
	case *types.CreateBucketAckPackage:
		operationType = types.OperationCreateBucket
		result = app.handleCreateBucketAckPackage(ctx, appCtx, p)
	case *types.DeleteBucketAckPackage:
		operationType = types.OperationDeleteBucket
		result = app.handleDeleteBucketAckPackage(ctx, appCtx, p)
	default:
		panic("unknown cross chain ack package type")
	}

	if len(result.Payload) != 0 {
		wrapPayload := types.CrossChainPackage{
			OperationType: operationType,
			Package:       result.Payload,
		}
		result.Payload = wrapPayload.MustSerialize()
	}

	return result
}

func (app *BucketApp) ExecuteFailAckPackage(ctx sdk.Context, appCtx *sdk.CrossChainAppContext, payload []byte) sdk.ExecuteResult {
	var pack interface{}
	var err error
	if ctx.IsUpgraded(upgradetypes.Pampas) {
		pack, err = types.DeserializeCrossChainPackageV2(payload, types.BucketChannelId, sdk.FailAckCrossChainPackageType)
	} else {
		pack, err = types.DeserializeCrossChainPackage(payload, types.BucketChannelId, sdk.FailAckCrossChainPackageType)
	}
	if err != nil {
		app.storageKeeper.Logger(ctx).Error("deserialize bucket cross chain package error", "payload", hex.EncodeToString(payload), "error", err.Error())
		panic("deserialize bucket cross chain package error")
	}

	var operationType uint8
	var result sdk.ExecuteResult
	switch p := pack.(type) {
	case *types.MirrorBucketSynPackage:
		operationType = types.OperationMirrorBucket
		result = app.handleMirrorBucketFailAckPackage(ctx, appCtx, p)
	case *types.CreateBucketSynPackage:
		operationType = types.OperationCreateBucket
		result = app.handleCreateBucketFailAckPackage(ctx, appCtx, p)
	case *types.CreateBucketSynPackageV2:
		operationType = types.OperationCreateBucket
		result = app.handleCreateBucketFailAckPackageV2(ctx, appCtx, p)
	case *types.DeleteBucketSynPackage:
		operationType = types.OperationDeleteBucket
		result = app.handleDeleteBucketFailAckPackage(ctx, appCtx, p)
	default:
		panic("unknown cross chain ack package type")
	}

	if len(result.Payload) != 0 {
		wrapPayload := types.CrossChainPackage{
			OperationType: operationType,
			Package:       result.Payload,
		}
		result.Payload = wrapPayload.MustSerialize()
	}

	return result
}

func (app *BucketApp) ExecuteSynPackage(ctx sdk.Context, appCtx *sdk.CrossChainAppContext, payload []byte) sdk.ExecuteResult {
	var pack interface{}
	var err error
	if ctx.IsUpgraded(upgradetypes.Pampas) {
		pack, err = types.DeserializeCrossChainPackageV2(payload, types.BucketChannelId, sdk.FailAckCrossChainPackageType)
	} else {
		pack, err = types.DeserializeCrossChainPackage(payload, types.BucketChannelId, sdk.FailAckCrossChainPackageType)
	}
	if err != nil {
		app.storageKeeper.Logger(ctx).Error("deserialize bucket cross chain package error", "payload", hex.EncodeToString(payload), "error", err.Error())
		panic("deserialize bucket cross chain package error")
	}

	var operationType uint8
	var result sdk.ExecuteResult
	switch p := pack.(type) {
	case *types.MirrorBucketSynPackage:
		operationType = types.OperationMirrorBucket
		result = app.handleMirrorBucketSynPackage(ctx, appCtx, p)
	case *types.CreateBucketSynPackage:
		operationType = types.OperationCreateBucket
		result = app.handleCreateBucketSynPackage(ctx, appCtx, p)
	case *types.CreateBucketSynPackageV2:
		operationType = types.OperationCreateBucket
		result = app.handleCreateBucketSynPackageV2(ctx, appCtx, p)
	case *types.DeleteBucketSynPackage:
		operationType = types.OperationDeleteBucket
		result = app.handleDeleteBucketSynPackage(ctx, appCtx, p)
	default:
		panic("unknown cross chain ack package type")
	}

	if len(result.Payload) != 0 {
		wrapPayload := types.CrossChainPackage{
			OperationType: operationType,
			Package:       result.Payload,
		}
		result.Payload = wrapPayload.MustSerialize()
	}

	return result
}

func (app *BucketApp) handleMirrorBucketAckPackage(ctx sdk.Context, appCtx *sdk.CrossChainAppContext, ackPackage *types.MirrorBucketAckPackage) sdk.ExecuteResult {
	bucketInfo, found := app.storageKeeper.GetBucketInfoById(ctx, math.NewUintFromBigInt(ackPackage.Id))
	if !found {
		app.storageKeeper.Logger(ctx).Error("bucket does not exist", "bucket id", ackPackage.Id.String())
		return sdk.ExecuteResult{
			Err: types.ErrNoSuchBucket,
		}
	}

	// update bucket
	if ackPackage.Status == types.StatusSuccess {
		sourceType, err := app.storageKeeper.GetSourceTypeByChainId(ctx, appCtx.SrcChainId)
		if err != nil {
			return sdk.ExecuteResult{
				Err: err,
			}
		}

		bucketInfo.SourceType = sourceType
		app.storageKeeper.SetBucketInfo(ctx, bucketInfo)
	}

	if err := ctx.EventManager().EmitTypedEvents(&types.EventMirrorBucketResult{
		Status:      uint32(ackPackage.Status),
		BucketName:  bucketInfo.BucketName,
		BucketId:    bucketInfo.Id,
		DestChainId: uint32(appCtx.SrcChainId),
	}); err != nil {
		return sdk.ExecuteResult{
			Err: err,
		}
	}

	return sdk.ExecuteResult{}
}

func (app *BucketApp) handleMirrorBucketFailAckPackage(ctx sdk.Context, appCtx *sdk.CrossChainAppContext, mirrorBucketPackage *types.MirrorBucketSynPackage) sdk.ExecuteResult {
	bucketInfo, found := app.storageKeeper.GetBucketInfoById(ctx, math.NewUintFromBigInt(mirrorBucketPackage.Id))
	if !found {
		app.storageKeeper.Logger(ctx).Error("bucket does not exist", "bucket id", mirrorBucketPackage.Id.String())
		return sdk.ExecuteResult{
			Err: types.ErrNoSuchBucket,
		}
	}

	bucketInfo.SourceType = types.SOURCE_TYPE_ORIGIN
	app.storageKeeper.SetBucketInfo(ctx, bucketInfo)

	if err := ctx.EventManager().EmitTypedEvents(&types.EventMirrorBucketResult{
		Status:      uint32(types.StatusFail),
		BucketName:  bucketInfo.BucketName,
		BucketId:    bucketInfo.Id,
		DestChainId: uint32(appCtx.SrcChainId),
	}); err != nil {
		return sdk.ExecuteResult{
			Err: err,
		}
	}

	return sdk.ExecuteResult{}
}

func (app *BucketApp) handleMirrorBucketSynPackage(ctx sdk.Context, header *sdk.CrossChainAppContext, synPackage *types.MirrorBucketSynPackage) sdk.ExecuteResult {
	app.storageKeeper.Logger(ctx).Error("received mirror bucket syn package ")

	return sdk.ExecuteResult{}
}

func (app *BucketApp) handleCreateBucketAckPackage(ctx sdk.Context, appCtx *sdk.CrossChainAppContext, ackPackage *types.CreateBucketAckPackage) sdk.ExecuteResult {
	app.storageKeeper.Logger(ctx).Error("received create bucket ack package ")

	return sdk.ExecuteResult{}
}

func (app *BucketApp) handleCreateBucketFailAckPackage(ctx sdk.Context, appCtx *sdk.CrossChainAppContext, synPackage *types.CreateBucketSynPackage) sdk.ExecuteResult {
	app.storageKeeper.Logger(ctx).Error("received create bucket fail ack package ")

	return sdk.ExecuteResult{}
}

func (app *BucketApp) handleCreateBucketFailAckPackageV2(ctx sdk.Context, appCtx *sdk.CrossChainAppContext, synPackage *types.CreateBucketSynPackageV2) sdk.ExecuteResult {
	app.storageKeeper.Logger(ctx).Error("received create bucket fail ack package ")

	return sdk.ExecuteResult{}
}

func (app *BucketApp) handleCreateBucketSynPackage(ctx sdk.Context, appCtx *sdk.CrossChainAppContext, createBucketPackage *types.CreateBucketSynPackage) sdk.ExecuteResult {
	err := createBucketPackage.ValidateBasic()
	if err != nil {
		return sdk.ExecuteResult{
			Payload: types.CreateBucketAckPackage{
				Status:    types.StatusFail,
				Creator:   createBucketPackage.Creator,
				ExtraData: createBucketPackage.ExtraData,
			}.MustSerialize(),
			Err: err,
		}
	}
	app.storageKeeper.Logger(ctx).Info("process create bucket syn package", "bucket name", createBucketPackage.BucketName)

	sourceType, err := app.storageKeeper.GetSourceTypeByChainId(ctx, appCtx.SrcChainId)
	if err != nil {
		return sdk.ExecuteResult{
			Err: err,
		}
	}

	bucketId, err := app.storageKeeper.CreateBucket(ctx,
		createBucketPackage.Creator,
		createBucketPackage.BucketName,
		createBucketPackage.PrimarySpAddress,
		&types.CreateBucketOptions{
			Visibility:       types.VisibilityType(createBucketPackage.Visibility),
			SourceType:       sourceType,
			ChargedReadQuota: createBucketPackage.ChargedReadQuota,
			PaymentAddress:   createBucketPackage.PaymentAddress.String(),
			PrimarySpApproval: &common.Approval{
				ExpiredHeight: createBucketPackage.PrimarySpApprovalExpiredHeight,
				Sig:           createBucketPackage.PrimarySpApprovalSignature,
			},
			ApprovalMsgBytes: createBucketPackage.GetApprovalBytes(),
		},
	)
	if err != nil {
		return sdk.ExecuteResult{
			Payload: types.CreateBucketAckPackage{
				Status:    types.StatusFail,
				Creator:   createBucketPackage.Creator,
				ExtraData: createBucketPackage.ExtraData,
			}.MustSerialize(),
			Err: err,
		}
	}

	return sdk.ExecuteResult{
		Payload: types.CreateBucketAckPackage{
			Status:    types.StatusSuccess,
			Id:        bucketId.BigInt(),
			Creator:   createBucketPackage.Creator,
			ExtraData: createBucketPackage.ExtraData,
		}.MustSerialize(),
	}
}

func (app *BucketApp) handleCreateBucketSynPackageV2(ctx sdk.Context, appCtx *sdk.CrossChainAppContext, createBucketPackageV2 *types.CreateBucketSynPackageV2) sdk.ExecuteResult {
	err := createBucketPackageV2.ValidateBasic()
	if err != nil {
		return sdk.ExecuteResult{
			Payload: types.CreateBucketAckPackage{
				Status:    types.StatusFail,
				Creator:   createBucketPackageV2.Creator,
				ExtraData: createBucketPackageV2.ExtraData,
			}.MustSerialize(),
			Err: err,
		}
	}
	app.storageKeeper.Logger(ctx).Info("process create bucket syn package v2", "bucket name", createBucketPackageV2.BucketName)

	sourceType, err := app.storageKeeper.GetSourceTypeByChainId(ctx, appCtx.SrcChainId)
	if err != nil {
		return sdk.ExecuteResult{
			Err: err,
		}
	}

	bucketId, err := app.storageKeeper.CreateBucket(ctx,
		createBucketPackageV2.Creator,
		createBucketPackageV2.BucketName,
		createBucketPackageV2.PrimarySpAddress,
		&types.CreateBucketOptions{
			Visibility:       types.VisibilityType(createBucketPackageV2.Visibility),
			SourceType:       sourceType,
			ChargedReadQuota: createBucketPackageV2.ChargedReadQuota,
			PaymentAddress:   createBucketPackageV2.PaymentAddress.String(),
			PrimarySpApproval: &common.Approval{
				ExpiredHeight:              createBucketPackageV2.PrimarySpApprovalExpiredHeight,
				GlobalVirtualGroupFamilyId: createBucketPackageV2.GlobalVirtualGroupFamilyId,
				Sig:                        createBucketPackageV2.PrimarySpApprovalSignature,
			},
			ApprovalMsgBytes: createBucketPackageV2.GetApprovalBytes(),
		},
	)
	if err != nil {
		return sdk.ExecuteResult{
			Payload: types.CreateBucketAckPackage{
				Status:    types.StatusFail,
				Creator:   createBucketPackageV2.Creator,
				ExtraData: createBucketPackageV2.ExtraData,
			}.MustSerialize(),
			Err: err,
		}
	}

	return sdk.ExecuteResult{
		Payload: types.CreateBucketAckPackage{
			Status:    types.StatusSuccess,
			Id:        bucketId.BigInt(),
			Creator:   createBucketPackageV2.Creator,
			ExtraData: createBucketPackageV2.ExtraData,
		}.MustSerialize(),
	}
}

func (app *BucketApp) handleDeleteBucketAckPackage(ctx sdk.Context, appCtx *sdk.CrossChainAppContext, ackPackage *types.DeleteBucketAckPackage) sdk.ExecuteResult {
	app.storageKeeper.Logger(ctx).Error("received delete bucket ack package ")

	return sdk.ExecuteResult{}
}

func (app *BucketApp) handleDeleteBucketFailAckPackage(ctx sdk.Context, appCtx *sdk.CrossChainAppContext, synPackage *types.DeleteBucketSynPackage) sdk.ExecuteResult {
	app.storageKeeper.Logger(ctx).Error("received delete bucket fail ack package ")

	return sdk.ExecuteResult{}
}

func (app *BucketApp) handleDeleteBucketSynPackage(ctx sdk.Context, appCtx *sdk.CrossChainAppContext, deleteBucketPackage *types.DeleteBucketSynPackage) sdk.ExecuteResult {
	err := deleteBucketPackage.ValidateBasic()
	if err != nil {
		return sdk.ExecuteResult{
			Payload: types.DeleteBucketAckPackage{
				Status:    types.StatusFail,
				Id:        deleteBucketPackage.Id,
				ExtraData: deleteBucketPackage.ExtraData,
			}.MustSerialize(),
			Err: err,
		}
	}

	app.storageKeeper.Logger(ctx).Info("process delete group syn package", "bucket id", deleteBucketPackage.Id.String())

	bucketInfo, found := app.storageKeeper.GetBucketInfoById(ctx, math.NewUintFromBigInt(deleteBucketPackage.Id))
	if !found {
		app.storageKeeper.Logger(ctx).Error("bucket does not exist", "bucket id", deleteBucketPackage.Id.String())
		return sdk.ExecuteResult{
			Payload: types.DeleteBucketAckPackage{
				Status:    types.StatusFail,
				Id:        deleteBucketPackage.Id,
				ExtraData: deleteBucketPackage.ExtraData,
			}.MustSerialize(),
			Err: types.ErrNoSuchBucket,
		}
	}

	sourceType, err := app.storageKeeper.GetSourceTypeByChainId(ctx, appCtx.SrcChainId)
	if err != nil {
		return sdk.ExecuteResult{
			Err: err,
		}
	}

	err = app.storageKeeper.DeleteBucket(ctx,
		deleteBucketPackage.Operator,
		bucketInfo.BucketName,
		types.DeleteBucketOptions{
			SourceType: sourceType,
		},
	)
	if err != nil {
		return sdk.ExecuteResult{
			Payload: types.DeleteBucketAckPackage{
				Status:    types.StatusFail,
				Id:        deleteBucketPackage.Id,
				ExtraData: deleteBucketPackage.ExtraData,
			}.MustSerialize(),
			Err: err,
		}
	}
	return sdk.ExecuteResult{
		Payload: types.DeleteBucketAckPackage{
			Status:    types.StatusSuccess,
			Id:        bucketInfo.Id.BigInt(),
			ExtraData: deleteBucketPackage.ExtraData,
		}.MustSerialize(),
	}
}
