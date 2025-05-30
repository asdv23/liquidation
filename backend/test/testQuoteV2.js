import { ethers } from "ethers";

const QUOTE_V2_ADDRESS = "0x3d4e44Eb1374240CE5F1B871ab261CD16335B76a";

const params = {
  tokenIn: "0x4200000000000000000000000000000000000006",
  tokenOut: "0x833589fcd6edb6e08f4c7c32d4f71b54bda02913",
  amount: 120917531,
  fee: 500,
  sqrtPriceLimitX96: 0,
};

const quoteExactInputSingle = async () => {
  const provider = new ethers.JsonRpcProvider("http://localhost:8546");
  const quoteV2Abi = [
    {
      "inputs": [{
        "components": [{
          "name": "tokenIn",
          "type": "address"
        }, {
          "name": "tokenOut",
          "type": "address"
        }, {
          "name": "amount",
          "type": "uint256"
        }, {
          "name": "fee",
          "type": "uint24"
        }, {
          "name": "sqrtPriceLimitX96",
          "type": "uint160"
        }],
        "name": "params",
        "type": "tuple"
      }],
      "name": "quoteExactOutputSingle",
      "outputs": [{
        "name": "amountIn",
        "type": "uint256"
      }, {
        "name": "sqrtPriceX96After",
        "type": "uint160"
      }, {
        "name": "initializedTicksCrossed",
        "type": "uint32"
      }, {
        "name": "gasEstimate",
        "type": "uint256"
      }],
      "stateMutability": "nonpayable",
      "type": "function"
    }
  ];

  const quoteV2Contract = new ethers.Contract(QUOTE_V2_ADDRESS, quoteV2Abi, provider);
  try {
    const result = await quoteV2Contract.quoteExactOutputSingle.staticCall(params);
    console.log("Quote result:", result);
  } catch (error) {
    console.error("Error calling quoteExactOutputSingle:", error);
  }
};

quoteExactInputSingle();
