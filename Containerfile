FROM node:18.18.2-buster

WORKDIR /app

#Copy the build directory

# Copy in package.json and package-lock.json
# Copy in source code and other assets
ADD build /app
copy node_modules /app
COPY pm2.yml .

# Install dependencies and devDependencies
RUN npm install pm2 -g

# Configure ENV
ENV NODE_ENV=production
ENV WEB_SERVER_PORT 8080

#expose the container port
EXPOSE 8080

# Launch the container by passing these parameters to the entrypoint
# These parameters can be overridden if you’d like
CMD ["pm2-runtime", "pm2.yml"]
