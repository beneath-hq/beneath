FROM node:12-alpine

COPY ./beneath-frontend/package.json ./beneath-frontend/yarn.lock /source/beneath-frontend/
WORKDIR /source/beneath-frontend
RUN yarn install

COPY ./beneath-frontend /source/beneath-frontend
RUN yarn build

EXPOSE 3000
CMD ["yarn", "start"]
