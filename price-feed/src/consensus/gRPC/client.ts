var parseArgs = require('minimist');
import path from 'path';

import {
  Server,
  ServerCredentials,
  loadPackageDefinition,
  credentials,
} from '@grpc/grpc-js';
import { loadSync } from '@grpc/proto-loader';

const protoPath = path.join(__dirname, '/proto/bftconsensus.proto');

var packageDefinition = loadSync(protoPath, {
  keepCase: true,
  longs: String,
  enums: String,
  defaults: true,
  oneofs: true,
});

var bftconsensusProto: any =
  loadPackageDefinition(packageDefinition).bftconsensus;

function main() {
  var argv = parseArgs(process.argv.slice(2), {
    string: 'target',
  });

  var target;
  if (argv.target) {
    target = argv.target;
  } else {
    target = 'localhost:50051';
  }

  var client = new bftconsensusProto.Greeter(
    target,
    credentials.createInsecure()
  );

  var user;
  if (argv._.length > 0) {
    user = argv._[0];
  } else {
    user = 'world';
  }
  // client.sayHello({ name: user }, function (err: any, response: any) {
  //   console.log('Greeting:', response.message);
  // });
  client.vote(
    {
      proposalId: 1,
      approve: true,
    },
    function (err: any, response: any) {
      console.log('Vote result:', response.message);
    }
  );
}

main();
