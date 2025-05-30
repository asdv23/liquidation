import { Module } from '@nestjs/common';
import { ConfigModule } from '@nestjs/config';
import { ScheduleModule } from '@nestjs/schedule';
import { ChainModule } from './chain/chain.module';
import { BorrowDiscoveryModule } from './borrow-discovery/borrow-discovery.module';
import { DatabaseModule } from './database/database.module';

@Module({
    imports: [
        ConfigModule.forRoot({
            isGlobal: true,
        }),
        ScheduleModule.forRoot(),
        ChainModule,
        BorrowDiscoveryModule,
        DatabaseModule,
    ],
})
export class AppModule { } 