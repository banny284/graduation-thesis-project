import { Contract, Signer, Provider, ethers } from 'ethers';
import dotenv from 'dotenv';
dotenv.config();
import { abi } from '../abis/PriceProvider.json';

import OracleController from '../controllers/oracle';
import { Ticker } from '../domain/ticker';
import { PriceSourceManager } from '../domain/price-source';
import { BitfinexPriceSource } from '../ports/bitfinex';
import { CoingeckoPriceSource } from '../ports/coingecko';
import { KrakenPriceSource } from '../ports/kraken';
import { OkxPriceSource } from '../ports/okx';
import { BinancePriceSource } from '../ports/binance';
import { ECPairFactory } from 'ecpair';
import * as ecc from 'tiny-secp256k1';

const PRIVATE_KEY = process.env.PRIVATE_KEY || '';
const IS_DEVELOPMENT = process.env.NODE_ENV === 'development';
const PRICE_PROVIDER_ADDRESS = process.env.PRICE_PROVIDER || '';
const PROVIDER = process.env.PROVIDER || '';

// oracle price sources (median will be applied)
const sources = [
  new BitfinexPriceSource(),
  new CoingeckoPriceSource(),
  new KrakenPriceSource(),
  new OkxPriceSource(),
  new BinancePriceSource(),
];

if (!PRIVATE_KEY) {
  console.error('Missing PRIVATE_KEY env var');
  process.exit(1);
}

const ECPair = ECPairFactory(ecc);
const oraclePair = ECPair.fromPrivateKey(Buffer.from(PRIVATE_KEY, 'hex'));

const oracleController = new OracleController(
  oraclePair,
  [
    Ticker.BTCUSD,
    Ticker.ETHUSD,
    Ticker.LINKUSD,
    Ticker.DOTUSD,
    Ticker.ADAUSD,
    Ticker.XRPUSD,
  ],
  new PriceSourceManager(sources, console.error),
  IS_DEVELOPMENT
);

export const PriceSubmiter = async () => {
  try {
    const provider = ethers.getDefaultProvider('goerli');
    const signer = new ethers.Wallet(PRIVATE_KEY, provider);

    const contract = new Contract(PRICE_PROVIDER_ADDRESS, abi, signer);

    //   get contract owner
    //   const owner = await contract.owner();
    //   console.log(owner);

    const results = await oracleController.getAttestationForTicker(
      Ticker.BTCUSD,
      '0',
      '0'
    );
    //   ok now submit the price to the contract

    const price = results?.lastPrice;

    await contract.changePriceNow(price);

    console.log('price submitted');
    return 1;
  } catch (e) {
    console.log('something went wrong,may be rpc node is not working');
    return 0;
  }
};
