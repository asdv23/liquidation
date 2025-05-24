import { Module } from '@nestjs/common';
import { BorrowDiscoveryService } from './borrow-discovery.service';
import { ChainModule } from '../chain/chain.module';

@Module({
    imports: [ChainModule],
    providers: [BorrowDiscoveryService],
    exports: [BorrowDiscoveryService],
})
export class BorrowDiscoveryModule { } 