
# node-webserver

Example NodeJS with Express webserver configured with cors and a public file directory. 

## Description:

The .gitlab-ci.yml file is a configuration file used by GitLab CI/CD pipelines. Here are the build steps:

1. GitLab runner will pull down the source
2. install the node packages
3. build the TypeScript app
4. execute the build.go file
5. buildconfig-prod.json will be parsed
6. build Docker image
7. publish Docker image to a private docker registry
8. deploy Docker image to Docker Swarm
9. verify the application is running 
10. check  the running app version
