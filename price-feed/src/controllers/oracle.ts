import * as ecc from 'tiny-secp256k1';
import { Get, Route, Tags, Path, Query } from 'tsoa';
import crypto from 'crypto';
import { ECPairInterface } from 'ecpair';
import { uint64LE } from '../utils/bufferutils';
import { PriceSource } from '../domain/price-source';
import { Ticker } from '../domain/ticker';
import { filterMessageTo32 } from '../utils/bufferutils';
import * as fs from 'fs';
import secp256k1 from 'secp256k1';
import { randomBytes } from 'crypto';
import keccak256 from 'keccak256';
import dotenv from 'dotenv';
dotenv.config();

export type OracleAttestation = {
  timestamp: string;
  lastPrice: string;
  attestation: {
    signature: string;
    message: string;
    messageHash: string;
    validatorAddress?: string;
  };
};

export type OracleInfo = {
  publicKey: string;
  availableTickers: string[];
};

@Route('oracle')
@Tags('Oracle')
export default class OracleController {
  constructor(
    private keyPair: ECPairInterface,
    private availableTickers: string[],
    private priceSource: PriceSource,
    private isDevelopment: boolean = false
  ) {}

  @Get('/')
  public getInfo(): OracleInfo {
    return {
      publicKey: this.keyPair.publicKey.toString('hex'),
      availableTickers: this.availableTickers,
    };
  }

  private getTimestampNowMs(): number {
    return Math.trunc(Date.now());
  }

  private isTickerAvailable(ticker: string): ticker is Ticker {
    return this.availableTickers.includes(ticker);
  }

  @Get('/:ticker')
  public async getAttestationForTicker(
    @Path() ticker: string,
    @Query() timestamp: string,
    @Query() lastPrice: string
  ): Promise<OracleAttestation | null> {
    if (!this.isTickerAvailable(ticker)) return null;

    // DEVELOPMENT ONLY: provide timestamp and lastPrice via querystring to "simulate" the oracle signing the message
    let timestampToUse = Number(timestamp);
    let lastPriceToUse = Number(lastPrice);

    // PRODUCTION: use the price source to get the timestamp and last price
    if (!this.isDevelopment) {
      const price = await this.priceSource.getPrice(ticker);
      timestampToUse = this.getTimestampNowMs();
      lastPriceToUse = Math.trunc(price);
    }

    try {
      const timpestampLE64 = uint64LE(timestampToUse);
      const priceLE64 = uint64LE(lastPriceToUse);
      const iso4217currencyCode = Buffer.from(ticker.replace('BTC', ''));
      const message = Buffer.from([
        ...timpestampLE64,
        ...priceLE64,
        ...iso4217currencyCode,
      ]);

      const hash = crypto
        .createHash('sha256')
        .update(filterMessageTo32(message))
        .digest();
      if (!this.keyPair.privateKey) throw new Error('No private key found');
      const sig = secp256k1.ecdsaSign(
        filterMessageTo32(message),
        this.keyPair.privateKey
      );

      // get wallet address from private key
      const privateKey = Buffer.from(process.env.PRIVATE_KEY!, 'hex');
      const publicKey = secp256k1.publicKeyCreate(privateKey);

      // decompress public key and use keccak256 to get the address
      const decompressedKey = secp256k1.publicKeyConvert(publicKey, false);
      const address = keccak256(Buffer.from(decompressedKey.slice(1)))
        .slice(-20)
        .toString('hex');

      return {
        timestamp: timestampToUse!.toString(),
        lastPrice: lastPriceToUse!.toString(),
        attestation: {
          signature: Buffer.from(sig.signature).toString('hex'),
          message: message.toString('hex'),
          messageHash: hash.toString('hex'),
          validatorAddress: '0x' + address,
        },
      };
    } catch (e) {
      console.error(e);
      throw new Error('Bitfinex: An error occurred while signing the message');
    }
  }
}
