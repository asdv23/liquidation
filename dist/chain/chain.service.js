"use strict";
var __decorate = (this && this.__decorate) || function (decorators, target, key, desc) {
    var c = arguments.length, r = c < 3 ? target : desc === null ? desc = Object.getOwnPropertyDescriptor(target, key) : desc, d;
    if (typeof Reflect === "object" && typeof Reflect.decorate === "function") r = Reflect.decorate(decorators, target, key, desc);
    else for (var i = decorators.length - 1; i >= 0; i--) if (d = decorators[i]) r = (c < 3 ? d(r) : c > 3 ? d(target, key, r) : d(target, key)) || r;
    return c > 3 && r && Object.defineProperty(target, key, r), r;
};
var ChainService_1;
Object.defineProperty(exports, "__esModule", { value: true });
exports.ChainService = void 0;
const common_1 = require("@nestjs/common");
const ethers_1 = require("ethers");
const chains_config_1 = require("../config/chains.config");
let ChainService = ChainService_1 = class ChainService {
    constructor() {
        this.logger = new common_1.Logger(ChainService_1.name);
        this.providers = new Map();
    }
    onModuleInit() {
        this.initializeProviders();
    }
    initializeProviders() {
        Object.entries(chains_config_1.chainsConfig).forEach(([chainName, config]) => {
            try {
                const provider = new ethers_1.ethers.JsonRpcProvider(config.rpcUrl);
                this.providers.set(chainName, provider);
                this.logger.log(`Initialized provider for ${chainName}`);
            }
            catch (error) {
                this.logger.error(`Failed to initialize provider for ${chainName}: ${error.message}`);
            }
        });
    }
    getProvider(chainName) {
        const provider = this.providers.get(chainName);
        if (!provider) {
            throw new Error(`Provider for chain ${chainName} not found`);
        }
        return provider;
    }
    getChainConfig(chainName) {
        const config = chains_config_1.chainsConfig[chainName];
        if (!config) {
            throw new Error(`Configuration for chain ${chainName} not found`);
        }
        return config;
    }
    getActiveChains() {
        return Object.keys(chains_config_1.chainsConfig);
    }
};
exports.ChainService = ChainService;
exports.ChainService = ChainService = ChainService_1 = __decorate([
    (0, common_1.Injectable)()
], ChainService);
//# sourceMappingURL=chain.service.js.map