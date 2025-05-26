import { ChainService } from './chain.service';
export declare class ChainController {
    private readonly chainService;
    constructor(chainService: ChainService);
    getActiveChains(): string[];
}
