import { OnModuleInit } from '@nestjs/common';
import { ChainService } from '../chain/chain.service';
export declare class BorrowDiscoveryService implements OnModuleInit {
    private readonly chainService;
    private readonly logger;
    private activeLoans;
    private liquidationTimes;
    constructor(chainService: ChainService);
    onModuleInit(): Promise<void>;
    private startListening;
    private checkHealthFactor;
    private recordLiquidationTime;
}
