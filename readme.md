# build tool
This tool helps you to build docker images in kubernetes using kaniko
## Running
Create a .env file and copy the keys from sample.env. Fill the values in .env
Next, Build the software
```
go build -o tool
```
Run below command to build docker images in your kube cluster using s3 as build context repo. I have attached a sample app(demo-app) to test
```
./tool deploy --project-dir ${pwd}/demo-app -m s3 -b buildcontext -n default
```