import { Module } from '@nestjs/common';
import { ConfigModule } from '@nestjs/config';
import { ScheduleModule } from '@nestjs/schedule';
import { ChainModule } from './chain/chain.module';
import { BorrowDiscoveryModule } from './borrow-discovery/borrow-discovery.module';
import { DatabaseService } from './database/database.service';

@Module({
    imports: [
        ConfigModule.forRoot({
            isGlobal: true,
        }),
        ScheduleModule.forRoot(),
        ChainModule,
        BorrowDiscoveryModule,
    ],
    providers: [DatabaseService],
})
export class AppModule { } 