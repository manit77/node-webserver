FROM node:18.18.2-buster

WORKDIR /app

#Copy assets
ADD app /app

#show working directory
RUN ls /app -la

#install npm
RUN npm install pm2 -g
RUN pm2 install pm2-logrotate
RUN pm2 set pm2-logrotate:max_size 50M
RUN pm2 set pm2-logrotate:retain 10
RUN pm2 set pm2-logrotate:compress true
RUN pm2 flush

ENV NODE_ENV {NODE_ENV}
ENV container_port {container_port}
ENV app_name {app_name}
ENV app_version {app_version}
ENV app_builddate {app_builddate}
ENV app_git_sha {app_git_sha}
ENV database_server_name {database_server_name}
ENV database_name {database_name}
ENV database_username {database_username}
ENV database_password {database_password}

RUN echo $container_port
#expose the container port
EXPOSE {external_port}

# Launch the container by passing these parameters to the entrypoint
# These parameters can be overridden if you’d like
CMD ["pm2-runtime", "pm2.yml"]
