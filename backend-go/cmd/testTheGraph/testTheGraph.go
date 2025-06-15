package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"liquidation-bot/internal/models"
	"net/http"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const (
	authToken = "ceb62ad7a9ad0cc24afbfa7c916cea3a"
	dbPath    = "/Users/sh001/Documents/codes/ivan/liquidation/backend-go/tmp.db"
	pageSize  = 1000 // 每页查询数量

	// eth
	// const chainName = 'eth';
	// const endpoint = 'https://gateway.thegraph.com/api/subgraphs/id/JCNWRypm7FYwV8fx5HhzZPSFaMxgkPuw4TnR3Gpi81zk';
	// base
	chainName = "base"
	endpoint  = "https://gateway.thegraph.com/api/subgraphs/id/GQFbb95cE6d8mV989mL5figjaGaKCQB3xqYrr1bRyXqF"
	// op
	// const chainName = 'op';
	// const endpoint = 'https://gateway.thegraph.com/api/deployments/id/QmRMNoAvjrr4DXT4tBJafCAPr2TQuRztMScyu51kKt542j';
	// arb
	// const chainName = 'arb';
	// const endpoint = 'https://gateway.thegraph.com/api/subgraphs/id/4xyasjQeREe7PxnF6wVdobZvCw5mhoHZq3T7guRpuNPf';
	// avalanche
	// const chainName = 'avax';
	// const endpoint = 'https://gateway.thegraph.com/api/subgraphs/id/72Cez54APnySAn6h8MswzYkwaL9KjvuuKnKArnPJ8yxb';
)

type Response struct {
	Message string `json:"message"`
}

type GraphqlReq struct {
	Query         string                  `json:"query"`
	Variables     *map[string]interface{} `json:"variables,omitempty"`
	OperationName *string                 `json:"operationName,omitempty"`
}

type GraphqlResponse struct {
	Data struct {
		Borrows []Borrow `json:"borrows"`
	} `json:"data"`
}

type Borrow struct {
	User struct {
		ID string `json:"id"`
	} `json:"user"`
	Timestamp int64 `json:"timestamp"`
}

// Forward the received GraphqlReq containing the query, variables and operationName to the Subgraph.
func performSubgraphQuery(req *GraphqlReq, apiKey string) (*[]byte, error) {
	// convert GraphqlReq object to JSON
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	// Create an HTTP client that will perform the request
	client := &http.Client{}
	// Forward the request to the Subgraph
	subgraphRequest, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewBuffer(reqBytes))
	if err != nil {
		return nil, err
	}
	// Add the apiKey as a Authorization Bearer token on the request
	subgraphRequest.Header.Set("Authorization", "Bearer "+apiKey)
	subgraphRequest.Header.Set("Accept", "application/json")
	subgraphRequest.Header.Set("Content-Type", "application/json")
	// Perform the GraphQL Request on the Subgraph
	subgraphResponse, err := client.Do(subgraphRequest)
	if err != nil {
		return nil, err
	}
	defer subgraphResponse.Body.Close()

	// Read & return the response body.
	subgraphResponseBody, err := io.ReadAll(subgraphResponse.Body)
	if err != nil {
		return nil, err
	}

	return &subgraphResponseBody, nil
}

func formatTimestamp(timestamp int64) string {
	t := time.Unix(timestamp, 0)
	return t.Format("2006-01-02 15:04:05")
}

// 1727712000: Tue Oct 01 2024 00:00:00 GMT+0800 (中国标准时间)
// 1738223465: Thu Jan 30 2025 15:51:05 GMT+0800 (中国标准时间)
// 1745555511: Fri Apr 25 2025 12:31:51 GMT+0800 (中国标准时间)
// 1745979694: Wed Apr 30 2025 10:21:34 GMT+0800 (中国标准时间)
func fetchBorrowsPage(lastTimestamp int64) ([]Borrow, error) {
	whereClause := `timestamp_gte: 1727712000, amount_gte: 5`
	if lastTimestamp != 0 {
		whereClause += fmt.Sprintf(`, timestamp_lt: %d`, lastTimestamp)
	}

	query := fmt.Sprintf(`{
			borrows(
				first: %d
				orderBy: timestamp
				orderDirection: desc
				where: { %s }
			) {
				user {
					id
				}
				timestamp
			}
		}`, pageSize, whereClause)
	subgraphqlResponse, err := performSubgraphQuery(&GraphqlReq{Query: query}, authToken)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %v", err)
	}
	if subgraphqlResponse == nil {
		return nil, fmt.Errorf("没有响应")
	}

	var result GraphqlResponse
	if err := json.Unmarshal(*subgraphqlResponse, &result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	return result.Data.Borrows, nil
}

func upsertLoan(db *gorm.DB, user string) error {
	result := db.Where(models.Loan{ChainName: chainName, User: user}).
		Assign(models.Loan{IsActive: true}).
		FirstOrCreate(&models.Loan{
			ChainName: chainName,
			User:      user,
			IsActive:  true,
		})
	return result.Error
}

func main() {
	// 连接数据库
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		fmt.Printf("无法连接到数据库: %v\n", err)
		return
	}

	// 自动迁移数据库结构
	err = db.AutoMigrate(&models.Loan{})
	if err != nil {
		fmt.Printf("数据库迁移失败: %v\n", err)
		return
	}

	var lastTimestamp int64
	hasMore := true
	totalProcessed := 0
	pageCount := 0

	for hasMore {
		pageCount++
		fmt.Printf("正在获取第 %d 页...\n", pageCount)

		borrows, err := fetchBorrowsPage(lastTimestamp)
		if err != nil {
			fmt.Printf("获取数据失败: %v\n", err)
			break
		}

		if len(borrows) == 0 {
			hasMore = false
			break
		}

		fmt.Printf("正在处理 %d 条借款记录...\n", len(borrows))

		for _, borrow := range borrows {
			if err := upsertLoan(db, borrow.User.ID); err != nil {
				fmt.Printf("存储用户 %s 的借款记录失败: %v\n", borrow.User.ID, err)
				continue
			}
		}

		totalProcessed += len(borrows)
		lastTimestamp = borrows[len(borrows)-1].Timestamp
		fmt.Printf("已处理 %d 条借款记录... 当前时间戳: %s\n", totalProcessed, formatTimestamp(lastTimestamp))

		if len(borrows) < pageSize {
			hasMore = false
		}

		// 添加延迟以避免请求过于频繁
		time.Sleep(time.Second)
	}

	fmt.Printf("%s 链处理完成！总共处理: %d\n", chainName, totalProcessed)
}
