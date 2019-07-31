FROM node:10.15 as builder

RUN mkdir -p /usr/src/app
WORKDIR /usr/src/app
ENV PATH /usr/src/app/node_modules/.bin:$PATH

ADD . .
COPY operator_ui/package.json operator_ui/
COPY evm/package.json evm/
COPY evm/box/package.json evm/box

RUN yarn install
RUN yarn workspace @chainlink/explorer build

ENV NODE_ENV production
ENTRYPOINT [ "yarn", "workspace", "@chainlink/explorer", "run", "prod" ]
