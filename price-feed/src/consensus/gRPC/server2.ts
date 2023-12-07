import {
  Server,
  ServerCredentials,
  loadPackageDefinition,
  credentials,
} from '@grpc/grpc-js';
import path from 'path';
import { loadSync } from '@grpc/proto-loader';
import OracleTestController from '../../controllers/oracle_gRPC';
import { Ticker } from '../../domain/ticker';
import { PriceSourceManager } from '../../domain/price-source';
import { BitfinexPriceSource } from '../../ports/bitfinex';
import { CoingeckoPriceSource } from '../../ports/coingecko';
import { KrakenPriceSource } from '../../ports/kraken';
import { OkxPriceSource } from '../../ports/okx';
import { BinancePriceSource } from '../../ports/binance';
import { ECPairFactory } from 'ecpair';
import * as ecc from 'tiny-secp256k1';
import dotenv from 'dotenv';
dotenv.config();

const PRIVATE_KEY = process.env.PRIVATE_KEY || '';
const IS_DEVELOPMENT = process.env.NODE_ENV === 'development';

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

const oracleController = new OracleTestController(
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

const protoPath = path.join(__dirname, '/proto/bftconsensus.proto');
const packageDefinition = loadSync(protoPath, {
  keepCase: true,
  longs: String,
  enums: String,
  defaults: true,
  oneofs: true,
});

const bftconsensusProto: any =
  loadPackageDefinition(packageDefinition).bftconsensus;

async function sayHello(call: any, callback: any) {
  // const rest = await oracleController.getAttestationForTicker(
  //   'BTCUSD',
  //   '1632835200000',
  //   '43000'
  // );

  // console.log('rest: ', rest);

  callback(null, { message: 'Hello ' + call.request.name });
}

// call other server

function main() {
  var server = new Server();
  server.addService(bftconsensusProto.Greeter.service, { sayHello: sayHello });
  const bindAddress = '0.0.0.0';
  const port = 50052;

  // call other server port 50052
  var client = new bftconsensusProto.Greeter(
    'localhost:50051',
    credentials.createInsecure()
  );

  function callSayHello() {
    client.sayHello({ name: 'server2' }, function (err: any, response: any) {
      console.log('Greeting:', response.message);
    });
  }

  server.bindAsync(
    `${bindAddress}:${port}`,
    ServerCredentials.createInsecure(),
    () => {
      server.start();
      console.log(`gRPC server running on http://${bindAddress}:${port}`);
      callSayHello(); // Call sayHello immediately
      setInterval(callSayHello, 2000); // Call sayHello repeatedly every 2 seconds
    }
  );
}

main();