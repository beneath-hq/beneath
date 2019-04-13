FROM gcr.io/google_appengine/nodejs
RUN /usr/local/bin/install_node '>=8.11.1'

COPY ./web-frontend/package.json ./web-frontend/yarn.lock /source/web-frontend/
WORKDIR /source/web-frontend
RUN yarn install

# Copy code
COPY ./web-frontend /source/web-frontend
RUN yarn build

# Run npm start when booted
EXPOSE 3000
CMD ["yarn", "start"]
