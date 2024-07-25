## SSO

### NEED TO INSTALL 
1. Protocol Buffer Compiler Installation
https://grpc.io/docs/protoc-installation/#install-using-a-package-manager Г

$ apt install -y protobuf-compiler
$ protoc --version  # Ensure compiler version is 3+

2. Install go plugins 
https://grpc.io/docs/languages/go/quickstart/

3. Генерим код:
~/GolandProjects/sso/protos$ protoc -I proto proto/sso/sso.proto --go_out=./gen/go --go_opt=paths=source_relative --go-grpc_out=./gen/go/ --go-grpc_opt=paths=source_relative

Сгенирированные файлы будут лежать в protos/gen/go/sso 

4. Для автоматизации установить task
   https://taskfile.dev/api/
   sudo snap install task --classic
   
   далее просто делаем >>>  task generate
   alex@black:~/GolandProjects/sso/protos$ task generate
   task: [generate] protoc -I proto proto/sso/*.proto --go_out=./gen/go/ --go_opt=paths=source_relative --go-grpc_out=./gen/go/ --go-grpc_opt=paths=source_relative

5. 