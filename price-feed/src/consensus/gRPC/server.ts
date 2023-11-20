import path from 'path';

import {
  Server,
  ServerCredentials,
  loadPackageDefinition,
} from '@grpc/grpc-js';
import { loadSync } from '@grpc/proto-loader';

const protoPath = path.join(__dirname, '/proto/helloworld.proto');

const packageDefinition = loadSync(protoPath, {
  keepCase: true,
  longs: String,
  enums: String,
  defaults: true,
  oneofs: true,
});

const helloWorldProto: any =
  loadPackageDefinition(packageDefinition).helloworld;

function sayHello(call: any, callback: any) {
  callback(null, { message: 'Hello ' + call.request.name });
}

function main() {
  var server = new Server();
  server.addService(helloWorldProto.Greeter.service, { sayHello: sayHello });
  const bindAddress = '0.0.0.0';
  const port = 50051;

  server.bindAsync(
    `${bindAddress}:${port}`,
    ServerCredentials.createInsecure(),
    () => {
      server.start();
      console.log(`gRPC server running on http://${bindAddress}:${port}`);
    }
  );
}

main();
