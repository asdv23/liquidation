package aavev3

import (
	"context"
	"fmt"
	"liquidation-bot/internal/models"
	"math/big"
	"strings"
	"sync"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
)

var (
	reserveBorrowedAmountField      string
	reserveCollateralAmountField    string
	reserveIsBorrowingField         string
	reserveIsUsingAsCollateralField string
)

func init() {
	schema, err := schema.Parse(&models.Reserve{}, &sync.Map{}, schema.NamingStrategy{})
	if err != nil {
		panic(fmt.Sprintf("failed to parse Reserve schema: %v", err))
	}
	for _, field := range schema.Fields {
		switch field.Name {
		case "BorrowedAmount":
			reserveBorrowedAmountField = field.DBName
		case "CollateralAmount":
			reserveCollateralAmountField = field.DBName
		case "IsBorrowing":
			reserveIsBorrowingField = field.DBName
		case "IsUsingAsCollateral":
			reserveIsUsingAsCollateralField = field.DBName
		}
	}
}

type DBWrapper struct {
	db *gorm.DB
}

func NewDBWrapper(db *gorm.DB) (*DBWrapper, error) {
	w := &DBWrapper{db: db}
	return w, nil
}

func (w *DBWrapper) GetDB() *gorm.DB {
	return w.db
}

func (w *DBWrapper) GetTokenInfoMap(chainName string) (map[string]*models.Token, error) {
	token := make([]*models.Token, 0)
	if err := w.db.Where(&models.Token{ChainName: chainName}).Find(&token).Error; err != nil {
		return nil, fmt.Errorf("failed to get token infos: %w", err)
	}
	tokenMap := make(map[string]*models.Token, 0)
	for _, t := range token {
		tokenMap[t.Address] = t
	}
	return tokenMap, nil
}

func (w *DBWrapper) GetTokenInfo(chainName string, address string) (*models.Token, error) {
	token := &models.Token{}
	if err := w.db.Where(&models.Token{ChainName: chainName, Address: address}).First(&token).Error; err != nil {
		return nil, fmt.Errorf("failed to get token info: %w", err)
	}
	return token, nil
}

func (w *DBWrapper) UpsertTokenInfo(chainName string, address string, symbol string, decimals, price *big.Int) (*models.Token, error) {
	token := &models.Token{
		ChainName: chainName,
		Address:   address,
		Symbol:    symbol,
		Decimals:  (*models.BigInt)(decimals),
		Price:     (*models.BigInt)(price),
	}
	if err := w.db.Where(&models.Token{ChainName: chainName, Address: address}).
		Assign(token).
		FirstOrCreate(&token).Error; err != nil {
		return nil, fmt.Errorf("failed to add token info: %w", err)
	}

	return token, nil
}

func (w *DBWrapper) UpdateTokenPrice(chainName string, address string, price *models.BigInt) error {
	if err := w.db.Model(&models.Token{}).Where(&models.Token{ChainName: chainName, Address: address}).
		Where("price <> ?", price).
		Update("price", price).Error; err != nil {
		return fmt.Errorf("failed to update token price: %w", err)
	}
	return nil
}

// user -> loan
func (w *DBWrapper) ChainActiveLoans(chainName string) ([]*models.Loan, error) {
	loans := make([]*models.Loan, 0)
	if err := w.db.Where(&models.Loan{ChainName: chainName, IsActive: true}).Find(&loans).Error; err != nil {
		return nil, fmt.Errorf("failed to get active loans: %w", err)
	}
	return loans, nil
}

func (w *DBWrapper) GetActiveLoansByToken(chainName string, tokenAddress string) ([]*models.Loan, error) {
	loans := make([]*models.Loan, 0)

	reserves := make([]*models.Reserve, 0)
	if err := w.db.Model(&models.Reserve{}).Where("chain_name = ? AND reserve = ?", chainName, tokenAddress).
		Where("is_using_as_collateral = ? OR is_borrowing = ?", true, true).
		Find(&reserves).Error; err != nil {
		return nil, fmt.Errorf("failed to get active loans by token: %w", err)
	}

	userMap := make(map[string]struct{}, 0)
	for _, reserve := range reserves {
		userMap[reserve.User] = struct{}{}
	}

	users := make([]string, 0)
	for user := range userMap {
		users = append(users, user)
	}

	if err := w.db.Where(&models.Loan{ChainName: chainName, IsActive: true}).Where("user IN (?)", users).
		Find(&loans).Error; err != nil {
		return nil, fmt.Errorf("failed to get active loans by token: %w", err)
	}
	return loans, nil
}

func (w *DBWrapper) GetLiquidationLoans(ctx context.Context, chainName string) ([]*models.Loan, error) {
	loans := make([]*models.Loan, 0)
	if err := w.db.Where(&models.Loan{ChainName: chainName, IsActive: true}).Where(
		"health_factor < ?", 1,
	).Find(&loans).Error; err != nil {
		return nil, fmt.Errorf("failed to get liquidation loans: %w", err)
	}
	return loans, nil
}

func (w *DBWrapper) GetNoLiquidationInfoLoans(ctx context.Context, chainName string) ([]*models.Loan, error) {
	loans := make([]*models.Loan, 0)
	if err := w.db.Where(&models.Loan{ChainName: chainName, IsActive: true}).Where(
		"liquidation_collateral_asset IS NULL OR liquidation_debt_asset IS NULL OR liquidation_liquidation_threshold = 0",
	).Find(&loans).Error; err != nil {
		return nil, fmt.Errorf("failed to get liquidation loans: %w", err)
	}
	return loans, nil
}

func (w *DBWrapper) GetLoan(ctx context.Context, chainName, user string) (*models.Loan, error) {
	loan := &models.Loan{}
	if err := w.db.Where(&models.Loan{ChainName: chainName, User: user}).First(&loan).Error; err != nil {
		return nil, fmt.Errorf("failed to get active loan: %w", err)
	}
	return loan, nil
}

func (w *DBWrapper) GetActiveLoan(ctx context.Context, chainName, user string) (*models.Loan, error) {
	loan := &models.Loan{}
	if err := w.db.Where(&models.Loan{ChainName: chainName, User: user, IsActive: true}).First(&loan).Error; err != nil {
		return nil, fmt.Errorf("failed to get active loan: %w", err)
	}
	return loan, nil
}

func (w *DBWrapper) CreateOrUpdateActiveLoan(chainName string, user string) (*models.Loan, error) {
	loan := models.Loan{
		ChainName: chainName,
		User:      user,
		IsActive:  true,
	}
	if err := w.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "chain_name"}, {Name: "user"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"is_active",
		}),
	}).Create(&loan).Error; err != nil {
		return nil, fmt.Errorf("failed to upsert active loan: %w", err)
	}
	return &loan, nil
}

func (w *DBWrapper) DeactivateActiveLoan(chainName string, user []string) error {
	if err := w.db.Model(&models.Loan{}).Where("chain_name = ? AND user IN (?)", chainName, user).Update("is_active", false).Error; err != nil {
		return fmt.Errorf("failed to deactivate active loan: %w", err)
	}
	return nil
}

func (w *DBWrapper) GetUserReserves(chainName string, user string) ([]*models.Reserve, error) {
	reserves := make([]*models.Reserve, 0)
	if err := w.db.Where(&models.Reserve{ChainName: chainName, User: user}).Find(&reserves).Error; err != nil {
		return nil, fmt.Errorf("failed to get user reserves: %w", err)
	}
	return reserves, nil
}

func (w *DBWrapper) GetUserLoansAndReservesByReserves(chainName string, reserves []string) ([]*models.Loan, []*models.Reserve, error) {
	// Step 1: 找到符合条件的用户 user 列表
	var users []string
	if err := w.db.Model(&models.Reserve{}).
		Select("DISTINCT reserves.user").
		Joins("LEFT JOIN loans ON reserves.chain_name = loans.chain_name AND reserves.user = loans.user").
		Where("loans.is_active = ? AND reserves.chain_name = ? AND reserves.reserve IN (?)", true, chainName, reserves).
		Pluck("reserves.user", &users).Error; err != nil {
		return nil, nil, fmt.Errorf("failed to query users: %w", err)
	}
	if len(users) == 0 {
		return nil, nil, gorm.ErrRecordNotFound
	}

	// Step 2: 查询所有这些用户的激活 loans
	var allLoans []*models.Loan
	if err := w.db.Where("chain_name = ? AND user IN (?) AND is_active = ?", chainName, users, true).
		Find(&allLoans).Error; err != nil {
		return nil, nil, fmt.Errorf("failed to query loans: %w", err)
	}

	// Step 3: 查询所有这些用户的 reserves
	var allReserves []*models.Reserve
	if err := w.db.Where("chain_name = ? AND user IN (?)", chainName, users).
		Find(&allReserves).Error; err != nil {
		return nil, nil, fmt.Errorf("failed to query reserves: %w", err)
	}

	return allLoans, allReserves, nil
}

func (w *DBWrapper) GetLiquidationInfo(chainName string, user string) (*models.LiquidationInfo, error) {
	loan := &models.Loan{}
	if err := w.db.Where(&models.Loan{ChainName: chainName, User: user}).First(&loan).Error; err != nil {
		return nil, fmt.Errorf("failed to get active loan: %w", err)
	}
	return loan.LiquidationInfo, nil
}

func (w *DBWrapper) AddUserReserves(chainName string, user string, reserves []*models.Reserve) error {
	if len(reserves) == 0 {
		return nil
	}

	// 批量 upsert
	if err := w.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "chain_name"}, {Name: "user"}, {Name: "reserve"}},
		DoUpdates: clause.AssignmentColumns([]string{
			reserveBorrowedAmountField,
			reserveCollateralAmountField,
			reserveIsBorrowingField,
			reserveIsUsingAsCollateralField,
		}),
	}).CreateInBatches(reserves, 100).Error; err != nil {
		return fmt.Errorf("failed to batch upsert user reserves: %w", err)
	}

	return nil
}

// BatchUpdateLoanLiquidationInfos 高性能批量更新贷款清算信息
func (w *DBWrapper) BatchUpdateLoanLiquidationInfos(chainName string, liquidationInfos []*UpdateLiquidationInfo) error {
	if len(liquidationInfos) == 0 {
		return nil
	}

	// 构建批量更新 SQL
	var cases []string
	var args []interface{}

	// 构建每个用户的 CASE WHEN 语句
	for _, info := range liquidationInfos {
		cases = append(cases, "WHEN user = ? THEN ?")
		args = append(args, info.User, info.HealthFactor)
	}
	healthFactorCase := strings.Join(cases, " ")

	// 重置并构建 total_collateral_base 的 CASE WHEN
	cases = cases[:0]
	for _, info := range liquidationInfos {
		cases = append(cases, "WHEN user = ? THEN ?")
		args = append(args, info.User, info.LiquidationInfo.TotalCollateralBase.String())
	}
	totalCollateralBaseCase := strings.Join(cases, " ")

	// 重置并构建 total_debt_base 的 CASE WHEN
	cases = cases[:0]
	for _, info := range liquidationInfos {
		cases = append(cases, "WHEN user = ? THEN ?")
		args = append(args, info.User, info.LiquidationInfo.TotalDebtBase.String())
	}
	totalDebtBaseCase := strings.Join(cases, " ")

	// 重置并构建 liquidation_threshold 的 CASE WHEN
	cases = cases[:0]
	for _, info := range liquidationInfos {
		cases = append(cases, "WHEN user = ? THEN ?")
		args = append(args, info.User, info.LiquidationInfo.LiquidationThreshold.String())
	}
	liquidationThresholdCase := strings.Join(cases, " ")

	// 重置并构建 collateral_asset 的 CASE WHEN
	cases = cases[:0]
	for _, info := range liquidationInfos {
		cases = append(cases, "WHEN user = ? THEN ?")
		args = append(args, info.User, info.LiquidationInfo.CollateralAsset)
	}
	collateralAssetCase := strings.Join(cases, " ")

	// 重置并构建 collateral_amount 的 CASE WHEN
	cases = cases[:0]
	for _, info := range liquidationInfos {
		cases = append(cases, "WHEN user = ? THEN ?")
		args = append(args, info.User, info.LiquidationInfo.CollateralAmount.String())
	}
	collateralAmountCase := strings.Join(cases, " ")

	// 重置并构建 collateral_amount_base 的 CASE WHEN
	cases = cases[:0]
	for _, info := range liquidationInfos {
		cases = append(cases, "WHEN user = ? THEN ?")
		args = append(args, info.User, info.LiquidationInfo.CollateralAmountBase.String())
	}
	collateralAmountBaseCase := strings.Join(cases, " ")

	// 重置并构建 debt_asset 的 CASE WHEN
	cases = cases[:0]
	for _, info := range liquidationInfos {
		cases = append(cases, "WHEN user = ? THEN ?")
		args = append(args, info.User, info.LiquidationInfo.DebtAsset)
	}
	debtAssetCase := strings.Join(cases, " ")

	// 重置并构建 debt_amount 的 CASE WHEN
	cases = cases[:0]
	for _, info := range liquidationInfos {
		cases = append(cases, "WHEN user = ? THEN ?")
		args = append(args, info.User, info.LiquidationInfo.DebtAmount.String())
	}
	debtAmountCase := strings.Join(cases, " ")

	// 重置并构建 debt_amount_base 的 CASE WHEN
	cases = cases[:0]
	for _, info := range liquidationInfos {
		cases = append(cases, "WHEN user = ? THEN ?")
		args = append(args, info.User, info.LiquidationInfo.DebtAmountBase.String())
	}
	debtAmountBaseCase := strings.Join(cases, " ")

	// 构建完整的更新 SQL
	sql := fmt.Sprintf(`
		UPDATE loans 
		SET health_factor = CASE %s END,
			liquidation_total_collateral_base = CASE %s END,
			liquidation_total_debt_base = CASE %s END,
			liquidation_liquidation_threshold = CASE %s END,
			liquidation_collateral_asset = CASE %s END,
			liquidation_collateral_amount = CASE %s END,
			liquidation_collateral_amount_base = CASE %s END,
			liquidation_debt_asset = CASE %s END,
			liquidation_debt_amount = CASE %s END,
			liquidation_debt_amount_base = CASE %s END,
			updated_at = ?
		WHERE chain_name = ? AND user IN (?);`,
		healthFactorCase,
		totalCollateralBaseCase,
		totalDebtBaseCase,
		liquidationThresholdCase,
		collateralAssetCase,
		collateralAmountCase,
		collateralAmountBaseCase,
		debtAssetCase,
		debtAmountCase,
		debtAmountBaseCase,
	)

	// 添加更新时间和查询条件参数
	args = append(args, time.Now(), chainName)

	// 构建用户列表
	users := make([]string, len(liquidationInfos))
	for i, info := range liquidationInfos {
		users[i] = info.User
	}
	args = append(args, users)

	// 执行批量更新
	if err := w.db.Exec(sql, args...).Error; err != nil {
		return fmt.Errorf("failed to batch update loan liquidation infos: %w", err)
	}

	return nil
}
