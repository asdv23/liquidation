"use strict";
var __decorate = (this && this.__decorate) || function (decorators, target, key, desc) {
    var c = arguments.length, r = c < 3 ? target : desc === null ? desc = Object.getOwnPropertyDescriptor(target, key) : desc, d;
    if (typeof Reflect === "object" && typeof Reflect.decorate === "function") r = Reflect.decorate(decorators, target, key, desc);
    else for (var i = decorators.length - 1; i >= 0; i--) if (d = decorators[i]) r = (c < 3 ? d(r) : c > 3 ? d(target, key, r) : d(target, key)) || r;
    return c > 3 && r && Object.defineProperty(target, key, r), r;
};
var __metadata = (this && this.__metadata) || function (k, v) {
    if (typeof Reflect === "object" && typeof Reflect.metadata === "function") return Reflect.metadata(k, v);
};
Object.defineProperty(exports, "__esModule", { value: true });
exports.ChainController = void 0;
const common_1 = require("@nestjs/common");
const chain_service_1 = require("./chain.service");
let ChainController = class ChainController {
    constructor(chainService) {
        this.chainService = chainService;
    }
    getActiveChains() {
        return this.chainService.getActiveChains();
    }
};
exports.ChainController = ChainController;
__decorate([
    (0, common_1.Get)(),
    __metadata("design:type", Function),
    __metadata("design:paramtypes", []),
    __metadata("design:returntype", void 0)
], ChainController.prototype, "getActiveChains", null);
exports.ChainController = ChainController = __decorate([
    (0, common_1.Controller)('chains'),
    __metadata("design:paramtypes", [chain_service_1.ChainService])
], ChainController);
//# sourceMappingURL=chain.controller.js.map