import { Module } from '@nestjs/common';
import { ChainService } from './chain.service';
import { ChainController } from './chain.controller';

@Module({
    providers: [ChainService],
    controllers: [ChainController],
    exports: [ChainService],
})
export class ChainModule { } 