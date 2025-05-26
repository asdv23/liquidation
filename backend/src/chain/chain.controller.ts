import { Controller, Get } from '@nestjs/common';
import { ChainService } from './chain.service';

@Controller('chains')
export class ChainController {
    constructor(private readonly chainService: ChainService) { }

    @Get()
    getActiveChains() {
        return this.chainService.getActiveChains();
    }
} 