# 多链监听配置

RPC_URLS = {
    "op": {"url":"https://opt-mainnet.g.alchemy.com/v2/821_LFssCCQnEG3mHnP7tSrc87IQKsUp","tick": "2s"},
    "op_sepolia": "https://opt-sepolia.g.alchemy.com/v2/821_LFssCCQnEG3mHnP7tSrc87IQKsUp",
    "base": "https://base-mainnet.g.alchemy.com/v2/LgcBLZ4hzWKCEjxKHFHKtYqC4fYCu_59",
    "mantle": "https://mantle-mainnet.g.alchemy.com/v2/821_LFssCCQnEG3mHnP7tSrc87IQKsUp"
}

config={
    "chain_rpc": "base",
    "protocol": "AAVE_V3",
    "contract": "0xA238Dd80C259a72e81d7e4664a9801593F98d1c5"
    "abi": "abis/AAVE_V3_POOL.json"
}

config={
    "chain_rpc": "op",
    "protocol": "AAVE_V3",
    "contract": "0x794a61358D6845594F94dc1DB02A252b5b4814aD"
    "abi": "abis/AAVE_V3_POOL_OP.json"
}

