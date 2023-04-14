FROM node:14.9-alpine

WORKDIR /usr/local/src
ADD package*.json /usr/local/src
RUN npm i

ADD . /usr/local/src
EXPOSE 3000

CMD ["npm", "run-script", "start"]