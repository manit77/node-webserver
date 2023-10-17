FROM node:18.18.2-buster

WORKDIR /app

#Copy assets
ADD build /app
ADD node_modules /app/node_modules
COPY pm2.yml .

#show working directory
RUN ls /app -la

#install npm
RUN npm install pm2 -g

# Configure ENV
ENV NODE_ENV=production
ENV WEB_SERVER_PORT 8080

#expose the container port
EXPOSE 8080

# Launch the container by passing these parameters to the entrypoint
# These parameters can be overridden if you’d like
CMD ["pm2-runtime", "pm2.yml"]
