# go-service-template

Go service template to use as a stub to create new web services leveraging the following technologies:
- gRPC
- protobuf
- grpc-gateway
- sqlc

## Management
To initialize the environment, after the git clone launch the following command:
```bash
make init
```

To refresh the binaries in the `bin` folder, launch the `deps` make target