import { Test, TestingModule } from '@nestjs/testing';
import { BorrowDiscoveryService } from './borrow-discovery.service';
import { ChainService } from '../chain/chain.service';
import { ConfigService } from '@nestjs/config';
import { DatabaseService } from '../database/database.service';
import { ChainConfig } from '../interfaces/chain-config.interface';

describe('BorrowDiscoveryService', () => {
    let service: BorrowDiscoveryService;
    let chainService: ChainService;

    const mockChainConfig: ChainConfig = {
        chainId: 1,
        name: 'ethereum',
        rpcUrl: 'https://eth-mainnet.alchemyapi.io/v2/your-api-key',
        contracts: {
            aavev3Pool: '0x...',
            flashLoanLiquidation: '0x...',
            usdc: '0x...',
        },
        blockTime: 400,
        minWaitTime: 200,
        nativePrice: 2000,
        minDebtUSD: 100,
    };

    beforeEach(async () => {
        const module: TestingModule = await Test.createTestingModule({
            providers: [
                BorrowDiscoveryService,
                {
                    provide: ChainService,
                    useValue: {
                        getChainConfig: jest.fn().mockImplementation((chainName) => mockChainConfig),
                    },
                },
                {
                    provide: ConfigService,
                    useValue: {
                        get: jest.fn().mockImplementation((key, defaultValue) => {
                            if (key === 'MAX_CHECK_INTERVAL') {
                                return 4 * 60 * 60 * 1000; // 4 hours in milliseconds
                            }
                            return defaultValue;
                        }),
                    },
                },
                {
                    provide: DatabaseService,
                    useValue: {},
                },
            ],
        }).compile();

        service = module.get<BorrowDiscoveryService>(BorrowDiscoveryService);
        chainService = module.get<ChainService>(ChainService);
    });

    describe('calculateWaitTime', () => {
        it('应该返回最小等待时间当健康因子低于或等于清算阈值时', () => {
            const healthFactors = [1.0004, 1.0005];
            healthFactors.forEach(hf => {
                const result = service['calculateWaitTime']('ethereum', hf);
                expect(result).toBe(200);
            });
        });

        it('应该返回最小等待时间当健康因子接近清算阈值时', () => {
            const result = service['calculateWaitTime']('ethereum', 1.0006);
            expect(result).toBe(400);
        });

        it('1.001', () => {
            const result = service['calculateWaitTime']('ethereum', 1.001);
            expect(result).toBe(400); // 4 hours
        });

        it('1.01', () => {
            const result = service['calculateWaitTime']('ethereum', 1.01);
            expect(result).toBe(130126); // 4 hours
        });

        it('1.05', () => {
            const result = service['calculateWaitTime']('ethereum', 1.05);
            expect(result).toBe(706686); // 4 hours
        });

        it('1.1', () => {
            const result = service['calculateWaitTime']('ethereum', 1.1);
            expect(result).toBe(1427387); // 4 hours
        });

        it('1.3', () => {
            const result = service['calculateWaitTime']('ethereum', 1.3);
            expect(result).toBe(4310190); // 4 hours
        });

        it('1.5', () => {
            const result = service['calculateWaitTime']('ethereum', 1.5);
            expect(result).toBe(7192992); // 4 hours
        });

        it('1.8', () => {
            const result = service['calculateWaitTime']('ethereum', 1.8);
            expect(result).toBe(11517197); // 4 hours
        });



        it('应该返回最大等待时间当健康因子等于健康阈值时', () => {
            const result = service['calculateWaitTime']('ethereum', 2.0);
            expect(result).toBe(4 * 60 * 60 * 1000);
        });

        it('应该返回最大等待时间当健康因子高于健康阈值时', () => {
            const result = service['calculateWaitTime']('ethereum', 2.1);
            expect(result).toBe(4 * 60 * 60 * 1000); // 4 hours
        });
    });
}); 