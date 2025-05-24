import { OnModuleInit } from '@nestjs/common';
import { ChainService } from '../chain/chain.service';
export declare class BorrowDiscoveryService implements OnModuleInit {
    private readonly chainService;
    private readonly logger;
    private readonly unsafeLoans;
    private readonly activeLoans;
    private readonly liquidationTimes;
    private readonly aaveV3PoolABI;
    private readonly pollingInterval;
    constructor(chainService: ChainService);
    onModuleInit(): Promise<void>;
    private startListening;
    private checkHealthFactor;
    private startHealthFactorPolling;
    private recordLiquidationTime;
    getUnsafeLoans(chain: string): any[];
    getLiquidationTimes(chain: string): any[];
}
