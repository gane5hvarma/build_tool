# build tool
This tool helps you to build docker images in kubernetes using kaniko
## Running
Create a .env file and copy the keys from sample.env. Fill the values in .env
Next, Build the software using the below command
```
go build -o tensorfuse
```
Run below command to build docker images in your kube cluster using s3 as build context repo. I have attached a sample app(demo-app) to test
```
./tensorfuse deploy --project-dir ${pwd}/demo-app -m s3 -b buildcontext -n default
```
To check the status of the build, you can check the pod logs. Run the below command to check pod logs
`kubectl logs ${podname}`
below is a sample log of build completion. After this you should be able to see it in dockerhub
```
INFO[0011] Taking snapshot of full filesystem...
INFO[0011] EXPOSE 5000
INFO[0011] Cmd: EXPOSE
INFO[0011] Adding exposed port: 5000/tcp
INFO[0011] ENV FLASK_APP=app.py
INFO[0011] CMD ["flask", "run", "--host=0.0.0.0", "--port=5000"]
INFO[0011] Pushing image to gane5h/build_tool:latest
INFO[0018] Pushed index.docker.io/gane5h/build_tool@sha256:b48e6996003d8957b19073da6c5086d40218f7079c327d78d9a0b58c87aad71b
```
you can pull the image and check if its running as expected
```
docker run -p 5000:5000 gane5h/build_tool:latest
```