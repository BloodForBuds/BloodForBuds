FROM node:22-bookworm-slim

RUN apt-get update \
    && apt-get install --yes --no-install-recommends ca-certificates curl default-jre-headless \
    && rm -rf /var/lib/apt/lists/*

RUN npm install --global firebase-tools@15.28.2

WORKDIR /firebase

COPY firebase.json ./firebase.json

EXPOSE 9099

CMD ["firebase", "emulators:start", "--only", "auth", "--project", "demo-bloodforbuds", "--non-interactive"]
