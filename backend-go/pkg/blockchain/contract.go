package blockchain

import (
	"fmt"

	aavev3 "liquidation-bot/bindings/aavev3"
	bindings "liquidation-bot/bindings/common"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"golang.org/x/sync/errgroup"
)

// ContractType 合约类型
type ContractType string

const (
	ContractTypeMulticall3           ContractType = "multicall3"
	ContractTypeFlashLoanLiquidation ContractType = "flashloanliquidation"
	ContractTypeAaveV3Pool           ContractType = "aavev3pool"
	ContractTypeAddressesProvider    ContractType = "addressesprovider"
	ContractTypeDataProvider         ContractType = "dataprovider"
	ContractTypePriceOracle          ContractType = "priceoracle"
	ContractTypeUSDC                 ContractType = "usdc"
)

type Contracts struct {
	Multicall3           *bindings.Multicall3
	FlashLoanLiquidation *bindings.FlashLoanLiquidation
	AaveV3Pool           *aavev3.Pool
	AddressesProvider    *aavev3.PoolAddressesProvider
	DataProvider         *aavev3.AaveProtocolDataProvider
	PriceOracle          *aavev3.AaveOracle
	USDC                 *bindings.ERC20
	Addresses            map[ContractType]common.Address
}

func NewContracts(backend bind.ContractBackend, contractsMap map[string]string) (*Contracts, error) {
	contracts := Contracts{
		Addresses: make(map[ContractType]common.Address),
	}
	var eg errgroup.Group
	for key, address := range contractsMap {
		contractType := ContractType(key)
		address := common.HexToAddress(address)
		contracts.Addresses[contractType] = address
		eg.Go(func() (err error) {
			switch contractType {
			case ContractTypeMulticall3:
				contracts.Multicall3, err = bindings.NewMulticall3(address, backend)

			case ContractTypeFlashLoanLiquidation:
				contracts.FlashLoanLiquidation, err = bindings.NewFlashLoanLiquidation(address, backend)

			case ContractTypeAaveV3Pool:
				// pool contract
				contracts.AaveV3Pool, err = aavev3.NewPool(address, backend)
				if err != nil {
					return fmt.Errorf("failed to create aaveV3Pool: %w", err)
				}
				// addressesProvider contract
				addressProvider, err := contracts.AaveV3Pool.ADDRESSESPROVIDER(nil)
				if err != nil {
					return fmt.Errorf("failed to get addressesProvider: %w", err)
				}
				contracts.Addresses[ContractTypeAddressesProvider] = addressProvider
				contracts.AddressesProvider, err = aavev3.NewPoolAddressesProvider(addressProvider, backend)
				if err != nil {
					return fmt.Errorf("failed to create addressesProvider: %w", err)
				}
				// dataProvider contract
				dataProvider, err := contracts.AddressesProvider.GetPoolDataProvider(nil)
				if err != nil {
					return fmt.Errorf("failed to get dataProvider: %w", err)
				}
				contracts.Addresses[ContractTypeDataProvider] = dataProvider
				contracts.DataProvider, err = aavev3.NewAaveProtocolDataProvider(dataProvider, backend)
				if err != nil {
					return fmt.Errorf("failed to create dataProvider: %w", err)
				}
				// priceOracle contract
				priceOracle, err := contracts.AddressesProvider.GetPriceOracle(nil)
				if err != nil {
					return fmt.Errorf("failed to get priceOracle: %w", err)
				}
				contracts.Addresses[ContractTypePriceOracle] = priceOracle
				contracts.PriceOracle, err = aavev3.NewAaveOracle(priceOracle, backend)
				if err != nil {
					return fmt.Errorf("failed to create priceOracle: %w", err)
				}
			case ContractTypeUSDC:
				contracts.USDC, err = bindings.NewERC20(address, backend)
			default:
				return fmt.Errorf("unknown contract type: %s", key)
			}
			return err
		})
	}
	if err := eg.Wait(); err != nil {
		return nil, err
	}
	return &contracts, nil
}
