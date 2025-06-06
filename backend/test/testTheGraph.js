const { request } = require('graphql-request');
const { PrismaClient } = require('@prisma/client');

// node test/testTheGraph.js
const prisma = new PrismaClient();

// eth
// const chainName = 'eth';
// const endpoint = 'https://gateway.thegraph.com/api/subgraphs/id/JCNWRypm7FYwV8fx5HhzZPSFaMxgkPuw4TnR3Gpi81zk';
// base
// const chainName = 'base';
// const endpoint = 'https://gateway.thegraph.com/api/subgraphs/id/D7mapexM5ZsQckLJai2FawTKXJ7CqYGKM8PErnS3cJi9';
// op
const chainName = 'op';
const endpoint = 'https://gateway.thegraph.com/api/deployments/id/QmRMNoAvjrr4DXT4tBJafCAPr2TQuRztMScyu51kKt542j';
// arb
// const chainName = 'arb';
// const endpoint = 'https://gateway.thegraph.com/api/subgraphs/id/4xyasjQeREe7PxnF6wVdobZvCw5mhoHZq3T7guRpuNPf';
// avalanche
// const chainName = 'avax';
// const endpoint = 'https://gateway.thegraph.com/api/subgraphs/id/72Cez54APnySAn6h8MswzYkwaL9KjvuuKnKArnPJ8yxb';

const PAGE_SIZE = 1000; // 每页查询数量

function formatTimestamp(timestamp) {
    if (!timestamp) return 'N/A';
    const date = new Date(Number(timestamp) * 1000);
    return date.toLocaleString('zh-CN', {
        year: 'numeric',
        month: '2-digit',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit',
        hour12: false
    });
}
// {
//     accounts(where: {id: "0x6216525b4107bca9e83fe302605ff2d81fa36992"}) {
//       borrows(orderBy: timestamp, orderDirection: desc) {
//         amountUSD
//         timestamp
//       }
//     }
//   }

// 1727712000: Tue Oct 01 2024 00:00:00 GMT+0800 (中国标准时间)
// 1738223465: Thu Jan 30 2025 15:51:05 GMT+0800 (中国标准时间)
// 1745555511: Fri Apr 25 2025 12:31:51 GMT+0800 (中国标准时间)
// 1745979694: Wed Apr 30 2025 10:21:34 GMT+0800 (中国标准时间)
async function fetchBorrowsPage(lastTimestamp = null) {
    const query = `{
        borrows(
            first: ${PAGE_SIZE}
            orderBy: timestamp
            orderDirection: desc
            where: {
                timestamp_gte: "1727712000"
                timestamp_lte: "1738223465"
                amountUSD_gte: "5"
                ${lastTimestamp ? `timestamp_lt: "${lastTimestamp}"` : ''}
            }
        ) {
            account {
                id
            }
            timestamp
        }
    }`;

    const headers = {
        Authorization: 'Bearer ceb62ad7a9ad0cc24afbfa7c916cea3a',
    };

    try {
        const data = await request(endpoint, query, {}, headers);
        return data.borrows;
    } catch (error) {
        console.error(`Error fetching page with timestamp ${lastTimestamp}:`, error);
        return [];
    }
}

async function fetchAndStoreBorrows() {
    let lastTimestamp = null;
    let hasMore = true;
    let totalProcessed = 0;
    let pageCount = 0;

    try {
        while (hasMore) {
            pageCount++;
            console.log(`Fetching page ${pageCount}...`);
            const borrows = await fetchBorrowsPage(lastTimestamp);
            
            if (borrows.length === 0) {
                hasMore = false;
                break;
            }

            console.log(`Processing ${borrows.length} borrows...`);
            
            for (const borrow of borrows) {
                try {
                    await prisma.loan.upsert({
                        where: {
                            chainName_user: {
                                chainName: chainName,
                                user: borrow.account.id.toLowerCase(),
                            },
                        },
                        update: {
                            isActive: true,
                            updatedAt: new Date(),
                        },
                        create: {
                            chainName: chainName,
                            user: borrow.account.id.toLowerCase(),
                            isActive: true,
                        },
                    });
                } catch (error) {
                    console.error(`Error storing borrow for user ${borrow.account.id}:`, error);
                }
            }

            totalProcessed += borrows.length;
            lastTimestamp = borrows[borrows.length - 1].timestamp;
            console.log(`Processed ${totalProcessed} borrows so far... Current timestamp: ${formatTimestamp(lastTimestamp)}`);

            if (borrows.length < PAGE_SIZE) {
                hasMore = false;
            }

            // 添加一个小延迟以避免请求过于频繁
            await new Promise(resolve => setTimeout(resolve, 1000));
        }

        console.log(`Successfully ${chainName} completed! Total processed: ${totalProcessed}`);
    } catch (error) {
        console.error('Error in main process:', error);
    } finally {
        await prisma.$disconnect();
    }
}

fetchAndStoreBorrows();
