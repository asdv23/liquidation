import { NestFactory } from '@nestjs/core';
import { AppModule } from './app.module';
import { ConsoleLogger } from '@nestjs/common';

// 自定义 Logger 类
class CustomLogger extends ConsoleLogger {
    log(message: any, ...optionalParams: any[]) {
        super.log(message, ...optionalParams);
    }

    warn(message: any, ...optionalParams: any[]) {
        super.warn(message, ...optionalParams);
    }

    error(message: any, ...optionalParams: any[]) {
        super.error(message, ...optionalParams);
    }
}

async function bootstrap() {
    const app = await NestFactory.create(AppModule, {
        logger: new CustomLogger(),
        bufferLogs: true,
    });

    const port = process.env.PORT || 3000;
    await app.listen(port);

    const logger = new CustomLogger('Bootstrap');
    logger.log(`Application is running on: http://localhost:${port}`);
}

bootstrap(); 