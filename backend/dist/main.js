"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
const core_1 = require("@nestjs/core");
const app_module_1 = require("./app.module");
const common_1 = require("@nestjs/common");
class CustomLogger extends common_1.ConsoleLogger {
    log(message, ...optionalParams) {
        super.log(message, ...optionalParams);
    }
    warn(message, ...optionalParams) {
        super.warn(message, ...optionalParams);
    }
    error(message, ...optionalParams) {
        super.error(message, ...optionalParams);
    }
}
async function bootstrap() {
    const app = await core_1.NestFactory.create(app_module_1.AppModule, {
        logger: new CustomLogger(),
        bufferLogs: true,
    });
    const port = process.env.PORT || 3000;
    await app.listen(port);
    const logger = new CustomLogger('Bootstrap');
    logger.log(`Application is running on: http://localhost:${port}`);
}
bootstrap();
//# sourceMappingURL=main.js.map