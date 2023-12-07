import { ECPairFactory } from 'ecpair';
import * as ecc from 'tiny-secp256k1';
import express, { Application } from 'express';
import morgan from 'morgan';
import swaggerUi from 'swagger-ui-express';
import cors from 'cors';
import dotenv from 'dotenv';
dotenv.config();
import { PriceSubmiter } from './submiter';

import routerFactory from './routes';
import { Ticker } from './domain/ticker';
import { PriceSourceManager } from './domain/price-source';
import { BitfinexPriceSource } from './ports/bitfinex';
import { CoingeckoPriceSource } from './ports/coingecko';
import { KrakenPriceSource } from './ports/kraken';
import { OkxPriceSource } from './ports/okx';
import { BinancePriceSource } from './ports/binance';
import OracleController from './controllers/oracle';
import PingController from './controllers/ping';

// ENV vars
const PORT = process.env.PORT || 8000;
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
const oracle = ECPair.fromPrivateKey(Buffer.from(PRIVATE_KEY, 'hex'));

const app: Application = express();

const router = routerFactory(
  new OracleController(
    oracle,
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
  ),
  new PingController()
);

app.use(express.json());
app.use(cors());
app.use(morgan('tiny'));
app.use(express.static('public'));
app.use(router);
app.use(
  '/docs',
  swaggerUi.serve,
  swaggerUi.setup(undefined, {
    swaggerOptions: {
      url: '/swagger.json',
    },
  })
);

const args = process.argv.slice(2);
const port = args[0] || process.env.PORT || 8000;

// call price submiter at 1 minute interval
setInterval(async () => {
  await PriceSubmiter();
}, 60000);

const server = app.listen(port, () => {
  console.log('Server is running on port', port);
});

function gracefulshutdown() {
  console.log('Shutting down');
  server.close(() => {
    console.log('HTTP server closed.');
    // When server has stopped accepting connections
    // exit the process with exit status 0
    process.exit(0);
  });
}

process.on('SIGTERM', gracefulshutdown);
process.on('SIGINT', gracefulshutdown);
