"use strict";
var __decorate = (this && this.__decorate) || function (decorators, target, key, desc) {
    var c = arguments.length, r = c < 3 ? target : desc === null ? desc = Object.getOwnPropertyDescriptor(target, key) : desc, d;
    if (typeof Reflect === "object" && typeof Reflect.decorate === "function") r = Reflect.decorate(decorators, target, key, desc);
    else for (var i = decorators.length - 1; i >= 0; i--) if (d = decorators[i]) r = (c < 3 ? d(r) : c > 3 ? d(target, key, r) : d(target, key)) || r;
    return c > 3 && r && Object.defineProperty(target, key, r), r;
};
Object.defineProperty(exports, "__esModule", { value: true });
exports.BorrowDiscoveryModule = void 0;
const common_1 = require("@nestjs/common");
const borrow_discovery_service_1 = require("./borrow-discovery.service");
const chain_module_1 = require("../chain/chain.module");
const database_module_1 = require("../database/database.module");
let BorrowDiscoveryModule = class BorrowDiscoveryModule {
};
exports.BorrowDiscoveryModule = BorrowDiscoveryModule;
exports.BorrowDiscoveryModule = BorrowDiscoveryModule = __decorate([
    (0, common_1.Module)({
        imports: [chain_module_1.ChainModule, database_module_1.DatabaseModule],
        providers: [borrow_discovery_service_1.BorrowDiscoveryService],
        exports: [borrow_discovery_service_1.BorrowDiscoveryService],
    })
], BorrowDiscoveryModule);
//# sourceMappingURL=borrow-discovery.module.js.map