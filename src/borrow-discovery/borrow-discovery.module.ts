import { Module } from '@nestjs/common';
import { BorrowDiscoveryService } from './borrow-discovery.service';
import { ChainModule } from '../chain/chain.module';
import { DatabaseModule } from '../database/database.module';

@Module({
    imports: [ChainModule, DatabaseModule],
    providers: [BorrowDiscoveryService],
    exports: [BorrowDiscoveryService],
})
export class BorrowDiscoveryModule { } 