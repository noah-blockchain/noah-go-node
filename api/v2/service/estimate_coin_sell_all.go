package service

import (
	"context"
	"fmt"
	"github.com/noah-blockchain/noah-go-node/core/code"
	"github.com/noah-blockchain/noah-go-node/core/commissions"
	"github.com/noah-blockchain/noah-go-node/core/transaction"
	"github.com/noah-blockchain/noah-go-node/core/types"
	"github.com/noah-blockchain/noah-go-node/formula"
	pb "github.com/noah-blockchain/node-grpc-gateway/api_pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"math/big"
)

// EstimateCoinSellAll return estimate of sell all coin transaction.
func (s *Service) EstimateCoinSellAll(ctx context.Context, req *pb.EstimateCoinSellAllRequest) (*pb.EstimateCoinSellAllResponse, error) {
	valueToSell, ok := big.NewInt(0).SetString(req.ValueToSell, 10)
	if !ok {
		return nil, status.Error(codes.InvalidArgument, "Value to sell not specified")
	}

	cState, err := s.blockchain.GetStateForHeight(req.Height)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	cState.RLock()
	defer cState.RUnlock()

	var coinToBuy types.CoinID
	if req.GetCoinToBuy() != "" {
		symbol := cState.Coins().GetCoinBySymbol(types.StrToCoinBaseSymbol(req.GetCoinToBuy()), types.GetVersionFromSymbol(req.GetCoinToBuy()))
		if symbol == nil {
			return nil, s.createError(status.New(codes.NotFound, "Coin to buy not exists"), transaction.EncodeError(code.NewCoinNotExists(req.GetCoinToBuy(), "")))
		}
		coinToBuy = symbol.ID()
	} else {
		coinToBuy = types.CoinID(req.GetCoinIdToBuy())
		if !cState.Coins().Exists(coinToBuy) {
			return nil, s.createError(status.New(codes.NotFound, "Coin to buy not exists"), transaction.EncodeError(code.NewCoinNotExists("", coinToBuy.String())))
		}
	}

	var coinToSell types.CoinID
	if req.GetCoinToSell() != "" {
		symbol := cState.Coins().GetCoinBySymbol(types.StrToCoinBaseSymbol(req.GetCoinToSell()), types.GetVersionFromSymbol(req.GetCoinToSell()))
		if symbol == nil {
			return nil, s.createError(status.New(codes.NotFound, "Coin to sell not exists"), transaction.EncodeError(code.NewCoinNotExists(req.GetCoinToSell(), "")))
		}
		coinToSell = symbol.ID()
	} else {
		coinToSell = types.CoinID(req.GetCoinIdToSell())
		if !cState.Coins().Exists(coinToSell) {
			return nil,
				s.createError(status.New(codes.NotFound, "Coin to sell not exists"), transaction.EncodeError(code.NewCoinNotExists("", coinToSell.String())))
		}
	}

	if coinToSell == coinToBuy {
		return nil, s.createError(status.New(codes.InvalidArgument, "\"From\" coin equals to \"to\" coin"),
			transaction.EncodeError(code.NewCrossConvert(coinToSell.String(), cState.Coins().GetCoin(coinToSell).GetFullSymbol(), coinToBuy.String(), cState.Coins().GetCoin(coinToBuy).GetFullSymbol())))
	}

	commissionInBaseCoin := big.NewInt(commissions.ConvertTx)
	if req.GasPrice > 1 {
		commissionInBaseCoin.Mul(commissionInBaseCoin, big.NewInt(int64(req.GasPrice)))
	}
	commissionInBaseCoin = big.NewInt(0).Mul(commissionInBaseCoin, transaction.CommissionMultiplier)

	coinFrom := cState.Coins().GetCoin(coinToSell)
	coinTo := cState.Coins().GetCoin(coinToBuy)

	value := valueToSell

	if !coinToSell.IsBaseCoin() {
		if coinFrom.Reserve().Cmp(commissionInBaseCoin) < 0 {
			return nil, s.createError(
				status.New(codes.InvalidArgument, fmt.Sprintf("Coin reserve balance is not sufficient for transaction. Has: %s, required %s",
					coinFrom.Reserve().String(), commissionInBaseCoin.String())),
				transaction.EncodeError(code.NewCoinReserveNotSufficient(
					coinFrom.GetFullSymbol(),
					coinFrom.ID().String(),
					coinFrom.Reserve().String(),
					commissionInBaseCoin.String(),
				)),
			)
		}

		value = formula.CalculateSaleReturn(coinFrom.Volume(), coinFrom.Reserve(), coinFrom.Crr(), valueToSell)
		if errResp := transaction.CheckReserveUnderflow(coinFrom, value); errResp != nil {
			return nil, s.createError(status.New(codes.FailedPrecondition, errResp.Log), errResp.Info)
		}
	}

	if !coinToBuy.IsBaseCoin() {
		if errResp := transaction.CheckForCoinSupplyOverflow(coinTo, value); errResp != nil {
			return nil, s.createError(status.New(codes.FailedPrecondition, errResp.Log), errResp.Info)
		}
		value.Sub(value, valueToSell)
		if value.Sign() != 1 {
			return nil, status.New(codes.FailedPrecondition, "Not enough coins to pay commission").Err()
		}
		value = formula.CalculatePurchaseReturn(coinTo.Volume(), coinTo.Reserve(), coinTo.Crr(), value)
	}

	return &pb.EstimateCoinSellAllResponse{
		WillGet: value.String(),
	}, nil
}
