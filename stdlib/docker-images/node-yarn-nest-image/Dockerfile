#ARG BUILDER_IMAGE
FROM node:18 as base
WORKDIR /usr/src/app

FROM base as builder
USER root
ARG CI_JOB_TOKEN
ENV NPM_AUTH ${CI_JOB_TOKEN}
COPY . .
RUN apk add --virtual build-dependencies && \
    apk add make gcc g++ && \
    yarn install --immutable && \
    yarn build

FROM base as release
COPY --from=builder /usr/src/app/node_modules ./node_modules
COPY --from=builder /usr/src/app/dist ./dist
COPY [ "package.json", "tsconfig.build.json", "tsconfig.json", "nest-cli.json", "./" ]

EXPOSE 80
CMD [ "yarn", "start:prod" ]

