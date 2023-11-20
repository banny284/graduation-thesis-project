import * as ecc from 'tiny-secp256k1';
import secp256k1 from 'secp256k1';

export function uint64LE(x: number): Buffer {
  const buffer = Buffer.alloc(8);
  writeUInt64LE(buffer, x, 0);
  return buffer;
}

// Copyright (c) 2011-2020 bitcoinjs-lib contributors (MIT License).
// Taken from https://github.com/bitcoinjs/bitcoinjs-lib/blob/master/ts_src/bufferutils.ts#L26-L36
export function writeUInt64LE(
  buffer: Buffer,
  value: number,
  offset: number
): number {
  verifuint(value, 0x001fffffffffffff);

  buffer.writeInt32LE(value & -1, offset);
  buffer.writeUInt32LE(Math.floor(value / 0x100000000), offset + 4);
  return offset + 8;
}

// https://github.com/feross/buffer/blob/master/index.js#L1127
function verifuint(value: number, max: number): void {
  if (typeof value !== 'number')
    throw new Error('cannot write a non-number as a number');
  if (value < 0)
    throw new Error('specified a negative value for writing an unsigned value');
  if (value > max) throw new Error('RangeError: value out of range');
  if (Math.floor(value) !== value)
    throw new Error('value has a fractional component');
}

// filter message buffer to 32 bytes

export function filterMessageTo32(message: Buffer): Buffer {
  const buffer = Buffer.alloc(32);
  message.copy(buffer, 0, 0, 32);
  return buffer;
}

// export function convertToEthAddress( ) {
//   try {

//     // Decode public key
//     const key = secp256k1.keyFromPublic('025f37d20e5b18909361e0ead7ed17c69b417bee70746c9e9c2bcb1394d921d4ae', 'hex');

//     // decompress public key
//     const decompressedKey =secp256k1.pu

//     // Convert to uncompressed format
//     const publicKey = key.getPublic().encode('hex').slice(2);

//     // Now apply keccak
//     const address = keccak256(Buffer.from(publicKey, 'hex')).slice(64 - 40);

//     console.log(`Public Key: 0x${publicKey}`);
//     console.log(`Address: 0x${address.toString()}`);
//   } catch (err) {
//     console.log(err);
//   }
// }
